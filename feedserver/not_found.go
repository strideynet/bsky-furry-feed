package feedserver

import (
	"go.uber.org/zap"
	"net/http"
)

func notFoundHandler(log *zap.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Info("request", zap.Any("r", r))
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Not Found"))
	}
}
