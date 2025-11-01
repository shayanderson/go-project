package app

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"sync/atomic"
	"time"
)

// httpRootPattern is a pattern that matches the root path "/"
const httpRootPattern = "/{$}"

// StatusError is an error with an associated HTTP status code
type StatusError interface {
	error
	Status() int
}

// statusError is a simple implementation of StatusError
type statusError struct {
	err    error
	status int
}

// Error implements the error interface
func (s statusError) Error() string {
	return s.err.Error()
}

// Status implements the StatusError interface
func (s statusError) Status() int {
	return s.status
}

// Handler is a http handler that returns an error
type Handler func(http.ResponseWriter, *http.Request) error

// ServeHTTP implements the http.Handler interface
func (r Handler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// log request
	slog.Info(
		fmt.Sprintf(
			"[http] %s http://%s%s %s from %s",
			req.Method,
			req.Host,
			req.RequestURI,
			req.Proto,
			req.RemoteAddr,
		),
	)

	if hErr := r(w, req); hErr != nil {
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
			"[http] %s http://%s%s %s from %s (%d)",
			req.Method,
			req.Host,
			req.RequestURI,
			req.Proto,
			req.RemoteAddr,
			err.Status(),
		), slog.String("err", err.Error()))
		// write error response
		w.Header().Set("Content-Type", "application/json")
		status := err.Status()
		if status < 400 || status > 599 {
			status = http.StatusInternalServerError
		}
		w.WriteHeader(status)
		if err := json.NewEncoder(w).Encode(map[string]string{"error": err.Error()}); err != nil {
			slog.Error("[http] failed to write error response", slog.String("err", err.Error()))
		}
	}
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

// ServerOptions holds the configuration options for the Server
type ServerOptions struct {
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
	options    ServerOptions
	middleware []Middleware
	mux        *http.ServeMux
	server     *http.Server
	stopping   atomic.Bool
}

// newServer creates a new Server instance
func newServer(options ServerOptions) *Server {
	s := &Server{
		options: options,
		mux:     http.NewServeMux(),
	}
	s.server = &http.Server{
		Addr:         options.Addr,
		Handler:      s.mux,
		IdleTimeout:  options.IdleTimeout,
		ReadTimeout:  options.ReadTimeout,
		WriteTimeout: options.WriteTimeout,
	}
	return s
}

// Handle registers a new route with a handler
func (s *Server) Handle(pattern string, handler Handler, middlewares ...Middleware) {
	s.mux.Handle(pattern, chain(handler, middlewares...))
}

// Mux returns the underlying http.ServeMux
func (s *Server) Mux() *http.ServeMux {
	return s.mux
}

// Start starts the HTTP server
func (s *Server) Start() error {
	h := Handler(func(w http.ResponseWriter, r *http.Request) error {
		s.mux.ServeHTTP(w, r)
		return nil
	})
	for i := len(s.middleware) - 1; i >= 0; i-- {
		h = s.middleware[i](h)
	}
	s.server.Handler = h

	var err error
	if s.options.CertFile != "" && s.options.CertKeyFile != "" {
		slog.Info(
			"[http] starting server", slog.String("addr", s.options.Addr), slog.Bool("tls", true),
		)
		err = s.server.ListenAndServeTLS(s.options.CertFile, s.options.CertKeyFile)
	} else {
		slog.Info(
			"[http] starting server", slog.String("addr", s.options.Addr), slog.Bool("tls", false),
		)
		err = s.server.ListenAndServe()
	}
	if err != nil && err == http.ErrServerClosed && s.stopping.Load() {
		// server is stopping, ignore error
		return nil
	}
	return err
}

// Stop stops the HTTP server
func (s *Server) Stop() error {
	slog.Info("[http] stopping server")
	s.stopping.Store(true)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	return s.server.Shutdown(ctx)
}

// Use adds middleware to the server
func (s *Server) Use(middlewares ...Middleware) {
	s.middleware = append(s.middleware, middlewares...)
}
