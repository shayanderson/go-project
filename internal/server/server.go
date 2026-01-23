package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"sync/atomic"
	"time"
)

// RootPattern is a pattern that matches the root path "/"
const RootPattern = "/{$}"

// Handler is a http handler that returns an error
type Handler func(*Context) error

// Serve serves an HTTP request
func (h Handler) Serve(c *Context) {
	req := c.Request()
	if !c.isMiddleware() {
		// log request when not in middleware
		slog.Info(
			fmt.Sprintf(
				"http: %s http://%s%s %s from %s",
				req.Method,
				req.Host,
				req.RequestURI,
				req.Proto,
				req.RemoteAddr,
			),
		)
	}

	if hErr := h(c); hErr != nil {
		var err StatusError
		if sErr, ok := hErr.(StatusError); ok {
			err = sErr
		} else {
			err = statusError{
				err:    hErr,
				status: http.StatusInternalServerError,
			}
		}
		// log error
		slog.Error(fmt.Sprintf(
			"http: %s http://%s%s %s from %s (%d)",
			req.Method,
			req.Host,
			req.RequestURI,
			req.Proto,
			req.RemoteAddr,
			err.Status(),
		), slog.String("err", err.Error()))
		// write error response
		status := err.Status()
		if status < 400 || status > 599 {
			status = http.StatusInternalServerError
		}
		c.Status(status)
		if err := c.JSON(map[string]string{"error": err.Error()}); err != nil {
			panic("http server failed to write error response: " + err.Error())
		}
	}
}

// ServeHTTP serves an HTTP request
func (r Handler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	c := NewContext(w, req)
	defer c.request.Body.Close()

	r.Serve(c)
}

// Middleware is a function that wraps a Handler
type Middleware func(Handler) Handler

// chain applies middleware to a handler
func chain(h Handler, middleware ...Middleware) Handler {
	for _, m := range middleware {
		h = m(h)
	}
	return h
}

// Options holds the configuration options for the Server
type Options struct {
	// Addr is the address to listen on
	Addr string
	// CertFile is the path to the TLS certificate file
	CertFile string
	// CertKeyFile is the path to the TLS certificate key file
	CertKeyFile string
	// IdleTimeout is the maximum amount of time to wait for the next request
	// when keep-alive is enabled
	IdleTimeout time.Duration
	// ReadHeaderTimeout is the amount of time allowed to read request headers
	ReadHeaderTimeout time.Duration
	// ReadTimeout is the maximum duration for reading the entire request, including the body
	ReadTimeout time.Duration
	// WriteTimeout is the maximum duration before timing out writes of the response
	WriteTimeout time.Duration
}

// Server is a simple HTTP server with middleware support
type Server struct {
	middleware []Middleware
	mux        *http.ServeMux
	opts       Options
	server     *http.Server
	stopping   atomic.Bool
}

// New creates a new server instance
func New(opts Options) *Server {
	if opts.ReadHeaderTimeout == 0 {
		opts.ReadHeaderTimeout = 3 * time.Second
	}
	if opts.ReadTimeout == 0 {
		opts.ReadTimeout = 5 * time.Second
	}
	if opts.WriteTimeout == 0 {
		opts.WriteTimeout = 5 * time.Second
	}

	s := &Server{
		opts: opts,
		mux:  http.NewServeMux(),
	}
	s.server = &http.Server{
		Addr:              opts.Addr,
		Handler:           s.mux,
		IdleTimeout:       opts.IdleTimeout,
		ReadHeaderTimeout: opts.ReadHeaderTimeout,
		ReadTimeout:       opts.ReadTimeout,
		WriteTimeout:      opts.WriteTimeout,
	}
	return s
}

// Handle registers a new route with a handler
func (s *Server) Handle(pattern string, handler Handler, middleware ...Middleware) {
	s.mux.Handle(pattern, chain(handler, middleware...))
}

// Mux returns the underlying http.ServeMux
func (s *Server) Mux() *http.ServeMux {
	return s.mux
}

// Start starts the HTTP server
func (s *Server) Start() error {
	// base handler to start the chain
	h := Handler(func(c *Context) error {
		s.mux.ServeHTTP(c.Writer(), c.Request())
		return nil
	})

	// apply middleware
	for i := len(s.middleware) - 1; i >= 0; i-- {
		h = s.middleware[i](h)
	}

	// wrap base handler
	s.server.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c := NewContext(w, r)
		c.middleware()
		h.Serve(c)
	})

	slog.Info("http server starting", slog.String("addr", s.opts.Addr))
	var err error
	if s.opts.CertFile != "" && s.opts.CertKeyFile != "" {
		err = s.server.ListenAndServeTLS(s.opts.CertFile, s.opts.CertKeyFile)
	} else {
		err = s.server.ListenAndServe()
	}
	if err != nil && err == http.ErrServerClosed && s.stopping.Load() {
		return nil
	}
	return err
}

// Stop stops the HTTP server
func (s *Server) Stop() error {
	slog.Info("http server stopping")
	s.stopping.Store(true)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	return s.server.Shutdown(ctx)
}

// Use adds middleware to the server
func (s *Server) Use(middleware ...Middleware) {
	s.middleware = append(s.middleware, middleware...)
}
