package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/joho/godotenv"
	"github.com/strideynet/bsky-furry-feed/api"
	"github.com/strideynet/bsky-furry-feed/bluesky"
	"github.com/strideynet/bsky-furry-feed/feed"
	"github.com/strideynet/bsky-furry-feed/ingester"
	"github.com/strideynet/bsky-furry-feed/store"
	"go.opentelemetry.io/contrib/detectors/gcp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sys/unix"
)

// TODO: Better, more granular, env configuration.
// A `inGCP` would make more sense rather than `isProduction`
type mode string

var (
	productionMode mode = "production"
	feedDevMode    mode = "feedDev"
	devMode        mode = "dev"
)

func getMode() (mode, error) {
	switch os.Getenv("ENV") {
	case "production":
		return productionMode, nil
	case "feedDev":
		return feedDevMode, nil
	case "dev":
		return devMode, nil
	default:
		return "", fmt.Errorf("unrecognized mode: %s", os.Getenv("ENV"))
	}
}

// TODO: Add Roles to Candidate Actor schema (or a seperate schema for actual
// feed users)
var moderatorDIDs = []string{
	// Noah
	"did:plc:dllwm3fafh66ktjofzxhylwk",
	// Newton
	"did:plc:ouytv644apqbu2pm7fnp7qrj",
	// Kio
	"did:plc:o74zbazekchwk2v4twee4ekb",
	// Kev
	"did:plc:bv2ckchoc76yobfhkrrie4g6",
	// KandyElmo
	"did:plc:xtuer24wmkwitvr2ajjeewra",
}

func main() {
	log, _ := zap.NewProduction()

	if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		log.Fatal("could not load existing .env file", zap.Error(err))
	}

	if err := runE(log); err != nil {
		log.Fatal("exited with error", zap.Error(err))
	}
}

func setupTracing(ctx context.Context, url string, mode mode) (func(), error) {
	var exp tracesdk.SpanExporter
	var err error
	if mode == productionMode {
		exp, err = otlptracehttp.New(ctx)
		if err != nil {
			return nil, fmt.Errorf("creating http trace exporter: %w", err)
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

	tracerProvider := tracesdk.NewTracerProvider(
		tracesdk.WithBatcher(exp),
		tracesdk.WithResource(r),
	)
	otel.SetTracerProvider(tracerProvider)

	return func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
		defer cancel()
		_ = tracerProvider.Shutdown(ctx)
	}, nil
}

func runE(log *zap.Logger) error {
	mode, err := getMode()
	if err != nil {
		return err
	}
	bskyCredentials, err := bluesky.CredentialsFromEnv()
	if err != nil {
		return fmt.Errorf("loading bsky credentials: %w", err)
	}

	ingesterEnabled := os.Getenv("BFF_INGESTER_ENABLED") == "1"
	apiEnabled := os.Getenv("BFF_API_ENABLED") == "1"

	log.Info("starting", zap.String("mode", string(mode)))

	ctx, cancel := signal.NotifyContext(context.Background(), unix.SIGINT)
	defer cancel()

	log.Info("setting up services")
	shutdownTrace, err := setupTracing(
		ctx,
		"http://localhost:14268/api/traces",
		mode,
	)
	if err != nil {
		return fmt.Errorf("creating tracer providers: %w", err)
	}
	defer shutdownTrace()

	var poolConnector store.PoolConnector
	switch mode {
	case productionMode:
		poolConnector = &store.CloudSQLConnector{
			Instance: "bsky-furry-feed:us-east1:main-us-east",
			Database: "bff",
			// TODO: Fetch this from an env var or from adc
			Username: "849144245446-compute@developer",
		}
	case feedDevMode:
		poolConnector = &store.CloudSQLConnector{
			Instance: "bsky-furry-feed:us-east1:main-us-east",
			Database: "bff",
			// TODO: Fetch this from an env var or from adc
			Username: "noah@noahstride.co.uk",
		}
	case devMode:
		poolConnector = &store.DirectConnector{
			URI: "postgres://bff:bff@localhost:5432/bff?sslmode=disable",
		}
	default:
		return fmt.Errorf("unhandled mode: %s", mode)
	}
	pgxStore, err := store.ConnectPGXStore(ctx, log.Named("store"), poolConnector)
	if err != nil {
		return fmt.Errorf("connecting to store: %w", err)
	}
	defer pgxStore.Close()

	// Create an errgroup to manage the lifetimes of the subservices.
	// If one exits, all will exit.
	eg, ctx := errgroup.WithContext(ctx)

	if ingesterEnabled {
		log.Info("setting up ingester")
		actorCache := ingester.NewActorCache(
			log.Named("actor_cache"),
			pgxStore,
		)
		// Prefill the actor cache before we proceed to ensure all actors
		// are available to sub-services. This eliminates some potential weirdness
		// when handling events/requests shortly after process startup.
		if err := actorCache.Sync(ctx); err != nil {
			return fmt.Errorf("filling candidate actor cache: %w", err)
		}
		eg.Go(func() error {
			return actorCache.Start(ctx)
		})

		fi := ingester.NewFirehoseIngester(
			log.Named("firehose_ingester"), pgxStore, actorCache, "wss://bsky.social",
		)
		eg.Go(func() error {
			return fi.Start(ctx)
		})
	}

	if apiEnabled {
		log.Info("setting up api")
		feedService := feed.ServiceWithDefaultFeeds(pgxStore)

		// Setup the public HTTP/XRPC server
		// TODO: Make these externally configurable
		hostname := "dev-feed.ottr.sh"
		if mode == productionMode {
			hostname = "feed.furryli.st"
		}
		listenAddr := ":1337"
		srv, err := api.New(
			log.Named("api"),
			hostname,
			listenAddr,
			feedService,
			pgxStore,
			bskyCredentials,
			&api.AuthEngine{
				ActorGetter:    pgxStore,
				TokenValidator: api.BSkyTokenValidator(bluesky.DefaultPDSHost),
			},
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
	}

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
