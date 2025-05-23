package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"time"

	"github.com/grafana/pyroscope-go"
	"github.com/strideynet/bsky-furry-feed/bfflog"
	"github.com/strideynet/bsky-furry-feed/scoring"
	"github.com/strideynet/bsky-furry-feed/worker"

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

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	slog.SetDefault(log)

	if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		log.Error(
			"could not load existing .env file",
			bfflog.Err(err),
		)
		os.Exit(1)
	}

	if err := runE(log); err != nil {
		log.Error(
			"exited with error",
			bfflog.Err(err),
		)
		os.Exit(1)
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

func runE(log *slog.Logger) error {
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
	scoreMaterializerEnabled := os.Getenv("BFF_SCORE_MATERIALIZER_ENABLED") == "1"
	backgroundWorkerEnabled := os.Getenv("BFF_BACKGROUND_WORKER_ENABLED") == "1"

	log.Info("starting bffsrv", slog.String("mode", string(mode)))

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

	if mode == productionMode {
		prof, err := pyroscope.Start(pyroscope.Config{
			ApplicationName: "bffsrv",

			// replace this with the address of pyroscope server
			ServerAddress:     "https://profiles-prod-001.grafana.net",
			BasicAuthUser:     os.Getenv("PYROSCOPE_USER"),
			BasicAuthPassword: os.Getenv("PYROSCOPE_PASSWORD"),

			// you can disable logging by setting this to nil
			Logger: &bfflog.PyroscopeSlogAdapter{
				Slog: bfflog.ChildLogger(log, "pyroscope"),
			},

			// you can provide static tags via a map:
			Tags: map[string]string{"hostname": os.Getenv("HOSTNAME")},

			ProfileTypes: []pyroscope.ProfileType{
				pyroscope.ProfileCPU,
				pyroscope.ProfileAllocObjects,
				pyroscope.ProfileAllocSpace,
				pyroscope.ProfileInuseObjects,
				pyroscope.ProfileInuseSpace,
			},
		})
		if err != nil {
			log.Error("fail to initialize pyroscope", bfflog.Err(err))
		} else {
			defer func() {
				err := prof.Stop()
				if err != nil {
					log.Info("error stopping prof", bfflog.Err(err))
				}
			}()
		}
	}

	var poolConnector store.PoolConnector
	switch mode {
	case productionMode:
		poolConnector = &store.DirectConnector{
			URI: os.Getenv("DB_URI"),
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
	pgxStore, err := store.ConnectPGXStore(
		ctx, bfflog.ChildLogger(log, "store"), poolConnector,
	)
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
			bfflog.ChildLogger(log, "actor_cache"),
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
			bfflog.ChildLogger(log, "ingester"), pgxStore, actorCache, "",
		)
		eg.Go(func() error {
			return fi.Start(ctx)
		})
	}

	if apiEnabled {
		log.Info("setting up api")
		feedService := feed.ServiceWithDefaultFeeds(pgxStore)

		// Setup the public HTTP/XRPC server
		hostname := os.Getenv("BFF_HOSTNAME")
		if hostname == "" {
			return fmt.Errorf("BFF_HOSTNAME not set")
		}
		listenAddr := ":1337"
		srv, err := api.New(
			ctx,
			bfflog.ChildLogger(log, "api"),
			hostname,
			listenAddr,
			feedService,
			pgxStore,
			bluesky.DefaultPDSHost,
			&api.AuthEngine{
				ActorGetter:    pgxStore,
				TokenValidator: api.BSkyTokenValidator(bluesky.DefaultPDSHost),
				Log:            bfflog.ChildLogger(log, "auth_engine"),
			},
		)
		if err != nil {
			return fmt.Errorf("creating feed server: %w", err)
		}

		eg.Go(func() error {
			log.Info("feed server listening", slog.String("addr", srv.Addr))
			go func() {
				<-ctx.Done()
				if err := srv.Close(); err != nil {
					log.Error("failed to close feed server", bfflog.Err(err))
				}
			}()
			return srv.ListenAndServe()
		})
	}

	if scoreMaterializerEnabled {
		log.Info("starting scoring materializer")
		hm := scoring.NewMaterializer(
			bfflog.ChildLogger(log, "scoring_materializer"),
			pgxStore,
			scoring.Opts{
				MaterializationInterval: 1 * time.Minute,
				RetentionPeriod:         15 * time.Minute,
				LookbackPeriod:          24 * time.Hour,
			},
		)
		eg.Go(func() error {
			return hm.Run(ctx)
		})
	}

	if backgroundWorkerEnabled {
		log.Info("starting background worker")

		eg.Go(func() error {
			worker, err := worker.New(
				ctx,
				bfflog.ChildLogger(log, "background_worker"),
				bluesky.DefaultPDSHost,
				bskyCredentials,
				pgxStore,
			)
			if err != nil {
				return fmt.Errorf("initializing worker: %w", err)
			}

			return worker.Run(ctx)
		})
	}

	// Setup private diagnostics/metrics server
	debugSrv := debugServer()
	eg.Go(func() error {
		log.Info("debug server listening", slog.String("addr", debugSrv.Addr))
		go func() {
			<-ctx.Done()
			if err := debugSrv.Close(); err != nil {
				log.Error("failed to close debug server", bfflog.Err(err))
			}
		}()
		return debugSrv.ListenAndServe()
	})

	log.Info("setup complete. running services")
	return eg.Wait()
}
