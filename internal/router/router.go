package router

import (
	"github.com/go-chi/chi/v5"
	"net/http"

	"short-linker/internal/handler"
)

type Router struct {
	linkHandler *handler.LinkHandler
	router      *chi.Mux
}

func NewRouter(linkHandler *handler.LinkHandler) *Router {
	return &Router{
		linkHandler: linkHandler,
	}
}

func (r *Router) SetupRoutes() http.Handler {
	router := chi.NewRouter()
	r.router = router

	r.linkRoutes()

	return router
}
