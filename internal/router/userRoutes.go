package router

func (r *Router) userRoutes() {
	r.router.Post("/api/user/signup", r.userHandler.Signup)
}