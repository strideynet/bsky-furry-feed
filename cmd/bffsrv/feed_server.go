package main

import (
	"encoding/json"
	"go.uber.org/zap"
	"net/http"
	"strconv"
)

type getFeedSkeletonParameters struct {
	cursor string
	limit  int
	feed   string
}

func feedServer(log *zap.Logger) *http.Server {
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
		// TODO: Make struct type for response
		output := map[string]any{
			"cursor": "my-cursor",
			"feed":   []any{},
		}
		encoder := json.NewEncoder(w)
		encoder.Encode(output)
		// TODO: Handle err.
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
