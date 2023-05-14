package main

import (
	"cloud.google.com/go/cloudsqlconn"
	"context"
	"fmt"
	texporter "github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/trace"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/oklog/run"
	"github.com/strideynet/bsky-furry-feed/feedserver"
	"github.com/strideynet/bsky-furry-feed/ingester"
	"github.com/strideynet/bsky-furry-feed/store"
	"go.opentelemetry.io/contrib/detectors/gcp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.uber.org/zap"
	"net"
	"os"
)

var tracer = otel.Tracer("bffsrv")

// TODO: Better, more granular, configuration.
// A `inGCP` would make more sense rather than `isProduction`
var inProduction = os.Getenv("ENV") == "production"

func main() {
	log, _ := zap.NewProduction()
	err := runE(log)
	if err != nil {
		log.Fatal("exited with error", zap.Error(err))
	}
}

func tracerProvider(ctx context.Context, url string) (*tracesdk.TracerProvider, error) {
	var exp tracesdk.SpanExporter
	var err error
	if inProduction {
		exp, err = texporter.New()
		if err != nil {
			return nil, fmt.Errorf("creating gcp trace exporter: %w", err)
		}
	} else {
		exp, err = jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(url)))
		if err != nil {
			return nil, fmt.Errorf("creating jaeger exporter: %w", err)
		}
	}

	r, err := resource.New(
		ctx,
		resource.WithDetectors(gcp.NewDetector()),
		resource.WithTelemetrySDK(),
		resource.WithAttributes(
			semconv.ServiceName("bffsrv"),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("creating resource attributes: %w", err)
	}

	tp := tracesdk.NewTracerProvider(
		tracesdk.WithBatcher(exp),
		tracesdk.WithResource(r),
	)
	otel.SetTracerProvider(tp)
	// TODO: Tracer shutdown

	return tp, nil
}

const localDBURL = "postgres://bff:bff@localhost:5432/bff?sslmode=disable"

func connectDB(ctx context.Context) (*pgxpool.Pool, error) {
	// TODO: Make this less horrible.
	// We should check env var for GCP Cloud SQL mode
	// We should detect the service account email.
	if inProduction {
		d, err := cloudsqlconn.NewDialer(context.Background(), cloudsqlconn.WithIAMAuthN())
		if err != nil {
			return nil, fmt.Errorf("creating cloud sql dialer: %w", err)
		}
		// TODO: Make this configurable
		cfg, err := pgxpool.ParseConfig("user=849144245446-compute@developer database=bff")
		if err != nil {
			return nil, fmt.Errorf("parsing cloud sql config: %w", err)
		}
		cfg.ConnConfig.DialFunc = func(ctx context.Context, _, _ string) (net.Conn, error) {
			return d.Dial(ctx, "bsky-furry-feed:us-east1:main-us-east")
		}
		return pgxpool.NewWithConfig(ctx, cfg)
	}
	return pgxpool.New(ctx, localDBURL)
}

func runE(log *zap.Logger) error {
	ctx := context.Background()
	log.Info("setting up services")
	_, err := tracerProvider(
		ctx,
		"http://localhost:14268/api/traces",
	)
	if err != nil {
		return fmt.Errorf("creating tracer provider: %w", err)
	}
	runGroup := run.Group{}

	pool, err := connectDB(ctx)
	if err != nil {
		return fmt.Errorf("connecting to db: %w", err)
	}
	defer pool.Close()
	queries := store.New(pool)
	crc := ingester.NewCandidateRepositoryCache(
		log.Named("candidate_repositories_cache"),
		queries,
	)
	if err := crc.Fetch(ctx); err != nil {
		return fmt.Errorf("filling candidate repository cache: %w", err)
	}

	// Setup ingester
	fi := ingester.NewFirehoseIngester(
		log.Named("firehose_ingester"), queries, crc,
	)
	runGroup.Add(fi.Start, func(_ error) {
		fi.Stop()
	})

	// Setup the public HTTP/XRPC server
	// TODO: Make these externally configurable
	hostname := "dev-feed.ottr.sh"
	if inProduction {
		hostname = "feed.ottr.sh"
	}
	listenAddr := ":1337"
	srv := feedserver.New(
		log.Named("feed_server"),
		queries,
		hostname,
		listenAddr,
	)
	runGroup.Add(func() error {
		log.Info("feed server listening", zap.String("addr", srv.Addr))
		return srv.ListenAndServe()
	}, func(err error) {
		srv.Close()
	})

	// Setup private diagnostics/metrics server
	debugSrv := debugServer()
	runGroup.Add(func() error {
		log.Info("debug server listening", zap.String("addr", debugSrv.Addr))
		return debugSrv.ListenAndServe()
	}, func(err error) {
		debugSrv.Close()
	})

	runGroup.Add(run.SignalHandler(context.Background(), os.Interrupt))

	log.Info("setup complete. running services")
	return runGroup.Run()
}
