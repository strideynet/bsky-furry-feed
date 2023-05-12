package main

import (
	"context"
	"fmt"
	"github.com/oklog/run"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	"golang.org/x/exp/slog"
	"net/http"
	"net/http/pprof"
	"os"
)

var tracer = otel.Tracer("github.com/strideynet/bsky-furry-feed")

func main() {
	log := slog.New(slog.NewTextHandler(os.Stderr, nil))
	err := runE(log)
	if err != nil {
		panic(err)
	}
}

func tracerProvider(url string) (*tracesdk.TracerProvider, error) {
	exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(url)))
	if err != nil {
		return nil, fmt.Errorf("creating jaeger exporter: %w", err)
	}

	tp := tracesdk.NewTracerProvider(
		tracesdk.WithBatcher(exp),
	)

	return tp, nil
}

func runE(log *slog.Logger) error {
	log.Info("setting up services")
	tp, err := tracerProvider("http://localhost:14268/api/traces")
	if err != nil {
		return fmt.Errorf("creating tracer provider: %w", err)
	}
	otel.SetTracerProvider(tp)

	// TODO: Tracer shutdown

	runGroup := run.Group{}

	// Setup ingester
	fi := &FirehoseIngester{
		stop:        make(chan struct{}),
		log:         log.WithGroup("firehoseIngester"),
		usersGetter: NewStaticCandidateUsers(),
	}
	runGroup.Add(fi.Start, func(_ error) {
		fi.Stop()
	})

	// Setup the public HTTP/XRPC server
	srv := feedServer(log)
	runGroup.Add(func() error {
		log.Info("feed server listening", "addr", srv.Addr)
		return srv.ListenAndServe()
	}, func(err error) {
		srv.Close()
	})

	// Setup private diagnostics/metrics server
	debugSrv := debugServer()
	runGroup.Add(func() error {
		log.Info("debug server listening", "addr", debugSrv.Addr)
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
