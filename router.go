package router

import (
	"net/http"
	"strings"
)

func New() *Router {
	return newRouter([]func(http.Handler) http.Handler{})
}

func newRouter(middlewares []func(http.Handler) http.Handler) *Router {
	return &Router{
		serveMux:    http.NewServeMux(),
		middlewares: middlewares,
	}
}

type Router struct {
	serveMux    *http.ServeMux
	middlewares []func(http.Handler) http.Handler
}

func (r *Router) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	chainMiddleware(r.middlewares)(r.serveMux).ServeHTTP(res, req)
}

func (r *Router) HandleFunc(pattern string, handlerFunc http.HandlerFunc) {
	r.Handle(pattern, handlerFunc)
}

func (r *Router) Handle(pattern string, handler http.Handler) {
	r.serveMux.Handle(pattern, handler)
}

func (r *Router) Group(pattern string, fn func(*Router)) {
	if pattern[len(pattern)-1] != '/' {
		pattern += "/"
	}

	subRouter := newRouter(r.middlewares)
	fn(subRouter)
	r.serveMux.Handle(pattern, http.StripPrefix(strings.TrimSuffix(pattern, "/"), subRouter))
}

func (r *Router) Pipe(middleware func(http.Handler) http.Handler) {
	r.middlewares = append(r.middlewares, middleware)
}

func chainMiddleware(middlewares []func(http.Handler) http.Handler) func(http.Handler) http.Handler {
	return func(final http.Handler) http.Handler {
		last := final
		for i := len(middlewares) - 1; i >= 0; i-- {
			last = middlewares[i](last)
		}
		return last
	}
}
