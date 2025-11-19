package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
)

type responseWriter struct {
	http.ResponseWriter
	statusCode int
	size       int
}

type gzipResponseWriter struct {
	http.ResponseWriter
	req         *http.Request
	gz          *gzip.Writer
	wroteHeader bool
}

type gzipReadCloser struct {
	gz   *gzip.Reader
	orig io.ReadCloser
}