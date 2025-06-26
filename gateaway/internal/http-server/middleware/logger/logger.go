package mvlogger

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/middleware"
	"go.uber.org/zap"
)

func New(next http.Handler, logger *zap.Logger)  http.Handler {
	log := logger.With(zap.String("component", "middleware/logger"))
	log.Info("logger middleware enabled")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		entry := log.With(
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
			zap.String("remote_addr", r.RemoteAddr),
			zap.String("user_agent", r.UserAgent()),
			zap.String("request_id", middleware.GetReqID(r.Context())),
		)

		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		t := time.Now()

		defer func() {
			entry.Info("request completed",
				zap.Int("status", ww.Status()),
				zap.Int("bytes", ww.BytesWritten()),
				zap.String("duration", time.Since(t).String()),
			)
		}()
		next.ServeHTTP(w, r)
	})
}