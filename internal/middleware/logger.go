package middleware

import (
	"net/http"
	"time"

	"short-linker/internal/logger"
)

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	size, err := rw.ResponseWriter.Write(b)
	rw.size += size
	return size, err
}

func LoggerMiddleware(l logger.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			rw := &responseWriter{ResponseWriter: w}

			next.ServeHTTP(rw, r)

			l.Info("HTTP Request",
				logger.String("URI", r.RequestURI),
				logger.String("method", r.Method),
				logger.Duration("duration", time.Since(start)),
				logger.Int("status", rw.statusCode),
				logger.Int("size", rw.size),
			)
		})
	}
}
