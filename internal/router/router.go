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
	pingHandler *handler.PingHandler
	router      *chi.Mux
}

func NewRouter(pingHandler *handler.PingHandler, linkHandler *handler.LinkHandler) *Router {
	return &Router{
		pingHandler: pingHandler,
		linkHandler: linkHandler,
	}
}

func (r *Router) SetupRoutes(logger *zap.Logger) http.Handler {
	router := chi.NewRouter()
	router.Use(middleware.LoggerMiddleware(logger))
	// I know, this is an overengineered, but I wanted to learn and practice it
	router.Use(middleware.GzipMiddleware)

	r.router = router

	r.baseRoutes()
	r.linkRoutes()

	return router
}
