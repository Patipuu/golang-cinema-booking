package middleware

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

// LoggerMiddleware logs method, path, remote addr, and duration using Zap.
func LoggerMiddleware(logger *zap.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			path := r.URL.Path
			method := r.Method
			remote := r.RemoteAddr
			next.ServeHTTP(w, r)
			duration := time.Since(start)
			logger.Info("request",
				zap.String("method", method),
				zap.String("path", path),
				zap.String("remote", remote),
				zap.Duration("duration", duration),
			)
		})
	}
}
