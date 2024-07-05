package server

import "net/http"

// router is an http router
type router struct {
	mux *http.ServeMux
	mw  []Middleware
}

// newRouter creates a new router
func newRouter(mux *http.ServeMux) *router {
	return &router{
		mux: mux,
		mw:  []Middleware{},
	}
}

// handle adds a handler to the router
func (r *router) handle(method, pattern string, handler Handler, middleware ...Middleware) {
	var h http.Handler = handler
	for i := len(middleware) - 1; i >= 0; i-- {
		h = middleware[i](h)
	}
	r.mux.Handle(method+" "+pattern, h)
}

// Delete adds a DELETE handler to the router
func (r *router) Delete(pattern string, handler Handler, middleware ...Middleware) {
	r.handle(http.MethodDelete, pattern, handler, middleware...)
}

// Get adds a GET handler to the router
func (r *router) Get(pattern string, handler Handler, middleware ...Middleware) {
	r.handle(http.MethodGet, pattern, handler, middleware...)
}

// Handle adds a handler to the router
func (r *router) Handle(method string, pattern string, handler Handler, middleware ...Middleware) {
	r.handle(method, pattern, handler, middleware...)
}

// Patch adds a PATCH handler to the router
func (r *router) Patch(pattern string, handler Handler, middleware ...Middleware) {
	r.handle(http.MethodPatch, pattern, handler, middleware...)
}

// Post adds a POST handler to the router
func (r *router) Post(pattern string, handler Handler, middleware ...Middleware) {
	r.handle(http.MethodPost, pattern, handler, middleware...)
}

// Put adds a PUT handler to the router
func (r *router) Put(pattern string, handler Handler, middleware ...Middleware) {
	r.handle(http.MethodPut, pattern, handler, middleware...)
}

// Use adds middleware to the router middleware stack
func (r *router) Use(mw ...Middleware) {
	r.mw = append(r.mw, mw...)
}

// ServeHTTP implements the http.Handler interface
func (r *router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	var h http.Handler = r.mux
	for i := len(r.mw) - 1; i >= 0; i-- {
		h = r.mw[i](h)
	}
	h.ServeHTTP(w, req)
}
