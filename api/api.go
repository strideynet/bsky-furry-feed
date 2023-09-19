package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"connectrpc.com/connect"
	"connectrpc.com/otelconnect"
	"github.com/rs/cors"
	"github.com/strideynet/bsky-furry-feed/bluesky"
	"github.com/strideynet/bsky-furry-feed/feed"
	"github.com/strideynet/bsky-furry-feed/proto/bff/v1/bffv1pbconnect"
	"github.com/strideynet/bsky-furry-feed/store"
	"go.uber.org/zap"
)

func handleErr(w http.ResponseWriter, log *zap.Logger, err error) {
	log.Error("failed to handle request", zap.Error(err))
	w.WriteHeader(500)
	_, _ = w.Write([]byte(fmt.Sprintf("failed to handle request: %s", err)))
}

func jsonHandler(log *zap.Logger, h func(r *http.Request) (any, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		respBody, err := h(r)
		if err != nil {
			handleErr(w, log, err)
			return
		}
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(200)
		encoder := json.NewEncoder(w)
		_ = encoder.Encode(respBody)
	}
}

type feedService interface {
	Metas() []feed.Meta
	GetFeedPosts(ctx context.Context, feedKey string, cursor string, limit int) (posts []feed.Post, err error)
}

func New(
	ctx context.Context,
	log *zap.Logger,
	hostname string,
	listenAddr string,
	feedService feedService,
	pgxStore *store.PGXStore,
	pdsHost string,
	bskyCredentials *bluesky.Credentials,
	authEngine *AuthEngine,
) (*http.Server, error) {
	mux := &http.ServeMux{}

	c := cors.New(cors.Options{
		AllowedOrigins: []string{
			"https://admin.furryli.st",
			"https://*.vercel.app",
			"https://furryli.st",
			"http://localhost:*",
			"https://buf.build",
		},
		AllowCredentials: true,
		AllowedHeaders: []string{
			"*",
		},
	})

	// Mount xrpc handlers
	didEndpointPath, didHandler, err := didHandler(log, hostname)
	if err != nil {
		return nil, fmt.Errorf("creating did handler: %w", err)
	}
	mux.Handle(didEndpointPath, didHandler)
	mux.Handle(getFeedSkeletonHandler(log, feedService))
	mux.Handle(describeFeedGeneratorHandler(log, hostname, feedService))

	client, err := bluesky.ClientFromCredentials(ctx, pdsHost, bskyCredentials)
	if err != nil {
		return nil, fmt.Errorf("creating bluesky client: %w", err)
	}

	// Mount Buf Connect services
	modSvcHandler := &ModerationServiceHandler{
		store:      pgxStore,
		log:        log,
		client:     client,
		authEngine: authEngine,
	}
	interceptors := connect.WithInterceptors(
		unaryLoggingInterceptor(log),
		otelconnect.NewInterceptor(),
	)
	mux.Handle(
		bffv1pbconnect.NewModerationServiceHandler(
			modSvcHandler,
			interceptors,
		),
	)
	mux.Handle(
		bffv1pbconnect.NewPublicServiceHandler(
			&PublicServiceHandler{
				feedService: feedService,
			},
			interceptors,
		),
	)
	mux.Handle(
		bffv1pbconnect.NewUserServiceHandler(
			&UserServiceHandler{
				authEngine: authEngine,
			}, interceptors,
		),
	)

	// Mount root/not found handler
	mux.Handle(rootHandler(log))

	return &http.Server{
		Addr:    listenAddr,
		Handler: c.Handler(mux),
	}, nil
}

func unaryLoggingInterceptor(log *zap.Logger) connect.UnaryInterceptorFunc {
	interceptor := func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			res, err := next(ctx, req)
			if err != nil {
				log.Error(
					"gRPC request failed",
					zap.String("procedure", req.Spec().Procedure),
					zap.Error(err),
				)
			} else {
				log.Info(
					"gRPC request handled",
					zap.String("procedure", req.Spec().Procedure),
				)
			}
			return res, err
		}
	}
	return interceptor
}
