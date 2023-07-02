package api

import (
	_ "embed"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.uber.org/zap"
	"net/http"
)

//go:embed html/root.html
var rootPage []byte

func rootHandler(log *zap.Logger) (string, http.Handler) {
	var h http.HandlerFunc = func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			log.Info("request to non-existent path", zap.Any("path", r.URL.Path))
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte("Not Found"))
			return
		}
		w.WriteHeader(200)
		_, _ = w.Write(rootPage)

	}

	return "/", otelhttp.NewHandler(h, "root")
}
