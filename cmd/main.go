package main

import (
	"encoding/json"
	"github.com/oklog/run"
	"go.opentelemetry.io/otel"
	"golang.org/x/exp/slog"
	"net/http"
	"os"
	"strconv"
)

var tracer = otel.Tracer("github.com/strideynet/bsky-furry-feed")

func main() {
	log := slog.New(slog.NewTextHandler(os.Stderr, nil))
	err := runE(log)
	if err != nil {
		panic(err)
	}
}

func runE(log *slog.Logger) error {
	runGroup := run.Group{}

	recordHandler := &RecordHandler{
		log: log.WithGroup("recordHandler"),
	}

	fh := &WebSocketFirehose{
		stop:                make(chan struct{}),
		log:                 log.WithGroup("firehose"),
		RecordCreateHandler: recordHandler.HandleCreate,
	}
	runGroup.Add(fh.Start, func(_ error) {
		fh.Stop()
	})

	srv := feedServer(log)
	runGroup.Add(func() error {
		return srv.ListenAndServe()
	}, func(err error) {
		srv.Close()
	})

	return runGroup.Run()
}

type getFeedSkeletonParameters struct {
	cursor string
	limit  int
	feed   string
}

func feedServer(log *slog.Logger) *http.Server {
	mux := &http.ServeMux{}
	mux.Handle("/xrpc/app.bsky.feed.getFeedSkeleton", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		params := getFeedSkeletonParameters{
			cursor: q.Get("cursor"),
			feed:   q.Get("feed"),
		}
		limitStr := q.Get("limit")
		if limitStr != "" {
			limit, err := strconv.Atoi(limitStr)
			if err != nil {
				panic(err)
			}
			params.limit = limit
		}

		w.WriteHeader(200)
		output := map[string]any{
			"cursor": "my-cursor",
			"feed":   []any{},
		}
		encoder := json.NewEncoder(w)
		encoder.Encode(output)
	}))
	mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Info("request")
		w.WriteHeader(http.StatusTeapot)
		w.Write([]byte("boo!"))
	}))

	return &http.Server{
		Addr:    ":1337",
		Handler: mux,
	}
}
