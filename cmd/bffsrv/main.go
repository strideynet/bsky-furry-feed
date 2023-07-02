package main

import (
	"context"
	"fmt"
	texporter "github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/trace"
	"github.com/strideynet/bsky-furry-feed/api"
	"github.com/strideynet/bsky-furry-feed/feed"
	"github.com/strideynet/bsky-furry-feed/ingester"
	"github.com/strideynet/bsky-furry-feed/store"
	"go.opentelemetry.io/contrib/detectors/gcp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sys/unix"
	"os"
	"os/signal"
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

	tpOpts := []tracesdk.TracerProviderOption{
		tracesdk.WithBatcher(exp),
		tracesdk.WithResource(r),
	}
	if inProduction {
		tpOpts = append(tpOpts, tracesdk.WithSampler(tracesdk.TraceIDRatioBased(0.001)))
	}

	tp := tracesdk.NewTracerProvider(
		tpOpts...,
	)
	otel.SetTracerProvider(tp)
	// TODO: Tracer shutdown

	return tp, nil
}

func runE(log *zap.Logger) error {
	ctx, cancel := signal.NotifyContext(context.Background(), unix.SIGINT)
	defer cancel()

	log.Info("setting up services")
	_, err := tracerProvider(
		ctx,
		"http://localhost:14268/api/traces",
	)
	if err != nil {
		return fmt.Errorf("creating tracer provider: %w", err)
	}

	var storeConfig store.Config
	if inProduction {
		storeConfig.CloudSQL = &store.CloudSQLConnectorConfig{
			Instance: "bsky-furry-feed:us-east1:main-us-east",
			Database: "bff",
		}
	} else {
		storeConfig.Direct = &store.DirectConnectorConfig{
			URI: "postgres://bff:bff@localhost:5432/bff?sslmode=disable",
		}
	}
	queries, queriesClose, err := storeConfig.Connect(ctx)
	if err != nil {
		return fmt.Errorf("connecting to store: %w", err)
	}
	defer queriesClose()

	crc := ingester.NewCandidateActorCache(
		log.Named("candidate_actor_cache"),
		queries,
	)
	// Prefill the CRC before we proceed to ensure all candidate actors
	// are available to sub-services. This eliminates some potential weirdness
	// when handling events/requests shortly after process startup.
	if err := crc.Sync(ctx); err != nil {
		return fmt.Errorf("filling candidate actor cache: %w", err)
	}

	// Create an errgroup to manage the lifetimes of the subservices.
	// If one exits, all will exit.
	eg, ctx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		return crc.Start(ctx)
	})

	// Setup ingester
	fi := ingester.NewFirehoseIngester(
		log.Named("firehose_ingester"), queries, crc,
	)
	eg.Go(func() error {
		return fi.Start(ctx)
	})

	feedService := feed.ServiceWithDefaultFeeds(queries)

	// Setup the public HTTP/XRPC server
	// TODO: Make these externally configurable
	hostname := "dev-feed.ottr.sh"
	if inProduction {
		hostname = "feed.furryli.st"
	}
	listenAddr := ":1337"
	srv, err := api.New(
		log.Named("api"),
		hostname,
		listenAddr,
		feedService,
	)

	if err != nil {
		return fmt.Errorf("creating feed server: %w", err)
	}

	eg.Go(func() error {
		log.Info("feed server listening", zap.String("addr", srv.Addr))
		go func() {
			<-ctx.Done()
			if err := srv.Close(); err != nil {
				log.Error("failed to close feed server", zap.Error(err))
			}
		}()
		return srv.ListenAndServe()
	})

	// Setup private diagnostics/metrics server
	debugSrv := debugServer()
	eg.Go(func() error {
		log.Info("debug server listening", zap.String("addr", debugSrv.Addr))
		go func() {
			<-ctx.Done()
			if err := debugSrv.Close(); err != nil {
				log.Error("failed to close debug server", zap.Error(err))
			}
		}()
		return debugSrv.ListenAndServe()
	})

	log.Info("setup complete. running services")
	return eg.Wait()
}
