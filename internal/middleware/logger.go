package middleware

import (
	"net/http"
	"time"
	"go.uber.org/zap"
)


func LoggerMiddleware(logger *zap.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			rw := &responseWriter{ResponseWriter: w}

			next.ServeHTTP(rw, r)

			logger.Info("HTTP Request",
				zap.String("URI", r.RequestURI),
				zap.String("method", r.Method),
				zap.Duration("duration", time.Since(start)),
				zap.Int("status", rw.statusCode),
				zap.Int("size", rw.size),
			)
		})
	}
}