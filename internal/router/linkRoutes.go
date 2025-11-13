package router

import (
	"github.com/go-chi/chi/v5"
	"net/http"
)

func (r *Router) linkRoutes() {
	r.router.Post("/", r.linkHandler.CreateShortLink)
	r.router.Get("/{id}", func(wr http.ResponseWriter, rq *http.Request) {
		id := chi.URLParam(rq, "id")
		r.linkHandler.RedirectPage(wr, rq, id)
	})
}