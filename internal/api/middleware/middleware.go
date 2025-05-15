package middleware

import (
	"net/http"

	"github.com/dylan0804/image-processing-tool/internal/api/logger"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func LoggingMiddleware(log *zap.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
			r.Header.Set("X-Request-ID", requestID)
		}

		reqLogger := log.With(
			zap.String("request_id", requestID),
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
			zap.String("remote_addr", r.RemoteAddr),
			zap.String("user_agent", r.UserAgent()),
		)

		// attach logger to context
		ctx := r.Context()
		ctx = logger.WithLogger(ctx, reqLogger)
		r = r.WithContext(ctx)

		reqLogger.Info("Request started")

		next.ServeHTTP(w, r)
	})
}