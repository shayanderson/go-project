package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/shayanderson/go-project/app/config"
)

// Handler is a http handler that returns an error
type Handler func(http.ResponseWriter, *http.Request) error

// ServeHTTP implements the http.Handler interface
func (r Handler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if err := r(w, req); err != nil {
		// #todo use cust error handler
		slog.Error("http handler error", "err", err)
		_ = WriteJSON(
			w,
			http.StatusInternalServerError,
			map[string]string{"error": "internal server error"},
		)
	}
}

// Server is an http server
type Server struct {
	Router *router
	server *http.Server
}

// New creates a new Server
func New(port int) *Server {
	s := &Server{
		Router: newRouter(http.NewServeMux()),
	}
	s.server = &http.Server{
		Addr:              fmt.Sprintf(":%d", port),
		Handler:           s.Router,
		ReadHeaderTimeout: 3 * time.Second,
	}
	return s
}

// Start starts the server
func (s *Server) Start() error {
	slog.Info("starting server", "port", config.Config.ServerPort)
	return s.server.ListenAndServe()
}

// Stop stops the server
func (s *Server) Stop(ctx context.Context) error {
	slog.Info("stopping server")
	ctx, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
	defer cancel()
	return s.server.Shutdown(ctx)
}

// ReadJSON reads a JSON request
func ReadJSON(r *http.Request, payload *any) error {
	return json.NewDecoder(r.Body).Decode(payload)
}

// WriteJSON writes a JSON response, with status code and sets content type to application/json
func WriteJSON(w http.ResponseWriter, code int, payload any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	return json.NewEncoder(w).Encode(payload)
}
