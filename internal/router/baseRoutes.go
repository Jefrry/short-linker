package router

func (r *Router) baseRoutes() {
	r.router.Get("/ping", r.pingHandler.Ping)
}