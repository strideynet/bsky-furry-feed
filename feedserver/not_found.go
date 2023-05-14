package feedserver

import (
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.uber.org/zap"
	"net/http"
)

func rootHandler(log *zap.Logger) (string, http.Handler) {
	var h http.HandlerFunc = func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			log.Info("request to non-existent path", zap.Any("path", r.URL.Path))
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("Not Found"))
			return
		}
		w.WriteHeader(200)
		w.Write([]byte("bluesky-furry-feed"))

	}

	return "/", otelhttp.NewHandler(h, "root")
}
