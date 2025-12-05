package router

import (
	"short-linker/internal/middleware"

	"github.com/go-chi/chi/v5"
	"net/http"
)

func (r *Router) linkRoutes() {
	r.router.Group(func(pr chi.Router) {
		pr.Use(middleware.OptionalAuthMiddleware([]byte(r.JWTSecret)))

		// Deprecated: use CreateShortLink with Content-Type: application/json instead
		pr.Post("/", r.linkHandler.CreateShortLinkPlain)
		pr.Post("/api/shorten", r.linkHandler.CreateShortLink)
		pr.Post("/api/shorten/batch", r.linkHandler.CreateShortLinkBatch)
	})
	r.router.Get("/{id}", func(wr http.ResponseWriter, rq *http.Request) {
		id := chi.URLParam(rq, "id")
		r.linkHandler.RedirectPage(wr, rq, id)
	})
}
