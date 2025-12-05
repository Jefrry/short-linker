package router

import (
	"short-linker/internal/middleware"

	"github.com/go-chi/chi/v5"
)

func (r *Router) userRoutes() {
	r.router.Post("/api/user/signup", r.userHandler.Signup)
	r.router.Post("/api/user/signin", r.userHandler.Signin)

	r.router.Group(func(pr chi.Router) {
		pr.Use(middleware.AuthMiddleware([]byte(r.JWTSecret)))

		pr.Post("/api/user/signout", r.userHandler.Signout)
		pr.Get("/api/user/profile", r.userHandler.GetProfile)
		pr.Get("/api/user/urls", r.userHandler.GetLinks)
	})
}
