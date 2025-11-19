package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (r *Router) linkRoutes() {
	// Deprecated: use CreateShortLink with Content-Type: application/json instead
	r.router.Post("/", r.linkHandler.CreateShortLinkPlain)
	r.router.Post("/api/shorten", r.linkHandler.CreateShortLink)
	r.router.Get("/{id}", func(wr http.ResponseWriter, rq *http.Request) {
		id := chi.URLParam(rq, "id")
		r.linkHandler.RedirectPage(wr, rq, id)
	})
}
