package service

import (
	"log/slog"

	"github.com/shayanderson/go-project/v2/internal/server"
)

// ExampleMiddleware is an example middleware
type ExampleMiddleware struct{}

// Handle implements the middleware interface
func (m ExampleMiddleware) Handle(next server.HandlerFunc) server.HandlerFunc {
	return func(c *server.Context) error {
		slog.Info("example middleware invoked")
		c.Set("example", "value from middleware") // example of setting a value in the context
		return next(c)
	}
}
