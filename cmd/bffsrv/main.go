package main

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/oklog/run"
	"github.com/strideynet/bsky-furry-feed/store"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.uber.org/zap"
	"net/http"
	"net/http/pprof"
	"os"
)

var tracer = otel.Tracer("github.com/strideynet/bsky-furry-feed")

func main() {
	log, _ := zap.NewDevelopment()
	err := runE(log)
	if err != nil {
		panic(err)
	}
}

func tracerProvider(ctx context.Context, url string) (*tracesdk.TracerProvider, error) {
	exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(url)))
	if err != nil {
		return nil, fmt.Errorf("creating jaeger exporter: %w", err)
	}

	r, err := resource.New(
		ctx,
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

	pool, err := pgxpool.New(ctx, localDBURL)
	if err != nil {
		return fmt.Errorf("creating db pool: %w", err)
	}
	defer pool.Close()
	crc := &candidateRepositoryCache{
		store: store.New(pool),
		log:   log.Named("candidate_repositories_cache"),
	}
	if err := crc.fetch(ctx); err != nil {
		return fmt.Errorf("filling candidate repository cache: %w", err)
	}

	// Setup ingester
	fi := &FirehoseIngester{
		stop: make(chan struct{}),
		log:  log.Named("firehose_ingester"),
		crc:  crc,
	}
	runGroup.Add(fi.Start, func(_ error) {
		fi.Stop()
	})

	// Setup the public HTTP/XRPC server
	srv := feedServer(log)
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

func debugServer() *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	mux.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	})
	mux.HandleFunc("/livez", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	})

	return &http.Server{
		Addr:    "127.0.0.1:1338",
		Handler: mux,
	}
}
