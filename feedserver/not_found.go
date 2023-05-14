package feedserver

import (
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.uber.org/zap"
	"net/http"
)

func notFoundHandler(log *zap.Logger) (string, http.Handler) {
	var h http.HandlerFunc = func(w http.ResponseWriter, r *http.Request) {
		log.Info("request to non-existent path", zap.Any("path", r.URL.Path))
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Not Found"))
	}

	return "/", otelhttp.NewHandler(h, "not_found")
}
