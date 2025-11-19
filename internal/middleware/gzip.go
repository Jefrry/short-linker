package middleware

import (
	"compress/gzip"
	"net/http"
	"strings"
)

func (g *gzipReadCloser) Read(p []byte) (int, error) {
	return g.gz.Read(p)
}

func (g *gzipReadCloser) Close() error {
	_ = g.gz.Close()
	return g.orig.Close()
}

func (g *gzipResponseWriter) WriteHeader(status int) {
	if g.wroteHeader {
		return
	}
	g.wroteHeader = true

	if g.Header().Get("Content-Encoding") == "" && acceptsGzip(g.req) && shouldCompressContentType(g.Header().Get("Content-Type")) {
		g.Header().Set("Content-Encoding", "gzip")
		g.Header().Del("Content-Length")
		vary := g.Header().Get("Vary")
		if vary == "" {
			g.Header().Set("Vary", "Accept-Encoding")
		} else if !strings.Contains(vary, "Accept-Encoding") {
			g.Header().Set("Vary", vary+", Accept-Encoding")
		}

		gz := gzip.NewWriter(g.ResponseWriter)
		g.gz = gz
	}

	g.ResponseWriter.WriteHeader(status)
}

func (g *gzipResponseWriter) Write(b []byte) (int, error) {
	if !g.wroteHeader {
		g.WriteHeader(http.StatusOK)
	}

	if g.gz != nil {
		return g.gz.Write(b)
	}
	
	return g.ResponseWriter.Write(b)
}

func GzipMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if enc := r.Header.Get("Content-Encoding"); enc != "" {
			if strings.Contains(strings.ToLower(enc), "gzip") {
				origBody := r.Body
				gz, err := gzip.NewReader(origBody)
				if err != nil {
					_ = origBody.Close()
					http.Error(w, "invalid gzip body", http.StatusBadRequest)
					return
				}

				r.Body = &gzipReadCloser{gz: gz, orig: origBody}
				r.Header.Del("Content-Encoding")
			}
		}

		grw := &gzipResponseWriter{ResponseWriter: w, req: r}
		defer func() {
			if grw.gz != nil {
				_ = grw.gz.Close()
			}
		}()

		next.ServeHTTP(grw, r)
	})
}

func acceptsGzip(r *http.Request) bool {
	ae := r.Header.Get("Accept-Encoding")
	return strings.Contains(strings.ToLower(ae), "gzip")
}

func shouldCompressContentType(ct string) bool {
	ct = strings.ToLower(ct)
	if ct == "" {
		return false
	}

	if strings.HasPrefix(ct, "application/json") || strings.HasPrefix(ct, "text/html") {
		return true
	}

	return false
}
