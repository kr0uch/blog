package middlewares

import (
	"blog/internal/logger"
	"net/http"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

func LoggerMiddleware(zapLogger logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			method := r.Method
			path := r.URL.Path
			requestID := r.Header.Get("X-Request-ID")
			if requestID == "" {
				requestID = uuid.New().String()
			}

			reqLogger := zapLogger.WithFields(
				zap.String(logger.RequestIDKey, requestID),
				zap.String(logger.MethodKey, method),
				zap.String(logger.PathKey, path))

			ctx := logger.LoggerWithContext(r.Context(), reqLogger)
			r = r.WithContext(ctx)

			w.Header().Set("X-Request-ID", requestID)
			next.ServeHTTP(w, r)

			latency := time.Since(start)
			reqLogger.Info("Request completed",
				zap.Duration("latency", latency),
			)
		})
	}
}
