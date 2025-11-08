package service

import (
	"log/slog"

	"github.com/shayanderson/go-project/v2/internal/server"
)

// ExampleMiddleware is an example middleware
type ExampleMiddleware struct{}

// Handle implements the middleware interface
func (m ExampleMiddleware) Handle(next server.Handler) server.Handler {
	return func(c *server.Context) error {
		slog.Info("example middleware invoked")
		return next(c)
	}
}
