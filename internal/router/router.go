package router

import (
	"net/http"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"short-linker/internal/middleware"
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

func (r *Router) SetupRoutes(logger *zap.Logger) http.Handler {
	router := chi.NewRouter()
	router.Use(middleware.LoggerMiddleware(logger))
	// I know, this is an overengineered, but I wanted to learn and practice it
	router.Use(middleware.GzipMiddleware)

	r.router = router

	r.linkRoutes()

	return router
}
