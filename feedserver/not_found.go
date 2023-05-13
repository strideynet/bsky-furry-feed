package feedserver

import (
	"go.uber.org/zap"
	"net/http"
)

func notFoundHandler(log *zap.Logger) (string, http.HandlerFunc) {
	return "/", func(w http.ResponseWriter, r *http.Request) {
		log.Info("request to non-existent path", zap.Any("path", r.URL.Path))
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Not Found"))
	}
}
