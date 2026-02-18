package httptransport

import "github.com/gin-gonic/gin"

type Handler interface {
	RegisterRoutes(router gin.IRouter)
}

type Router struct {
	engine   gin.IRouter
	handlers []Handler
}

func NewRouter(engine gin.IRouter, handlers []Handler) *Router {
	return &Router{
		engine:   engine,
		handlers: handlers,
	}
}

func (r *Router) RegisterRoutes() {
	for _, handler := range r.handlers {
		handler.RegisterRoutes(r.engine)
	}
}
