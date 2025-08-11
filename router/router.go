package router

import (
	"net/http"

	"github.com/atlastore/belt/server/options"
	"github.com/gofiber/fiber/v3"
)

var (
	_ options.Router = &Router{}
)

type Method string
const (
	GET Method = http.MethodGet
	POST Method = http.MethodPost
	PUT Method = http.MethodPut
	DELETE Method = http.MethodDelete
	PATCH Method = http.MethodPatch
	OPTIONS Method = http.MethodOptions
	HEAD Method = http.MethodHead
	ALL Method = "ALL"
)

type route struct {
	path string
	method Method
	handler fiber.Handler
	middleware []fiber.Handler
}

type RouteGroup struct {
	prefix string
	middleware []fiber.Handler
	routes []route
}

func (g *RouteGroup) Add(path string, method Method, handler fiber.Handler, middleware ...fiber.Handler) {
	g.routes = append(g.routes, route{
		path: path,
		method: method,
		handler: handler,
		middleware: middleware,
	})
}

type Router struct {
	globalMiddleware []fiber.Handler
	routes []route
	groups []*RouteGroup
}

func New() *Router {
	return &Router{}
}


func (r *Router) AddGlobalMiddleware(middleware ...fiber.Handler) {
	r.globalMiddleware = append(r.globalMiddleware, middleware...)
}

func (r *Router) Add(path string, method Method, handler fiber.Handler, middleware ...fiber.Handler) {
	r.routes = append(r.routes, route{
		path: path,
		method: method,
		handler: handler,
		middleware: middleware,
	})
}

func (r *Router) Group(prefix string, middleware ...fiber.Handler) *RouteGroup {
	group := &RouteGroup{
		prefix: prefix,
		middleware: middleware,
	}
	
	r.groups = append(r.groups, group)
	
	return group
}

func (r *Router) RegisterRoutes(app *fiber.App) {
	if len(r.globalMiddleware) > 0 {
		app.Use(r.globalMiddleware)
	}	

	for _, route := range r.routes {
		register(app, route.method, route.path, route.handler, route.middleware...)
	}

	for _, group := range r.groups {
		gr := app.Group(group.prefix, group.middleware...)
		for _, route := range group.routes {
			register(gr, route.method, route.path, route.handler, route.middleware...)
		}
	}

}

func register(app fiber.Router, method Method, path string, handler fiber.Handler, middleware ...fiber.Handler) {
	switch method {
	case GET:
		app.Get(path, handler, middleware...)
	case POST:
		app.Post(path, handler, middleware...)
	case PUT:
		app.Put(path, handler, middleware...)
	case DELETE:
		app.Delete(path, handler, middleware...)
	case PATCH:
		app.Patch(path, handler, middleware...)
	case OPTIONS:
		app.Options(path, handler, middleware...)
	case HEAD:
		app.Head(path, handler, middleware...)
	case ALL:
		app.All(path, handler, middleware...)
	default:
		app.All(path, handler, middleware...)
	}
}