package item

import (
	"log/slog"

	"github.com/shayanderson/go-project/v2/internal/assert"
	"github.com/shayanderson/go-project/v2/internal/server"
)

// Handler handles item HTTP requests
type Handler struct {
	service *Service
}

// NewHandler creates a new item handler
func NewHandler(service *Service) *Handler {
	assert.NotNil(service, "service is nil")
	return &Handler{service: service}
}

// Get handles GET /items
func (h *Handler) Get(c *server.Context) error {
	val := c.Get("example") // example of retrieving a value set by middleware
	slog.Info("example value from context", "value", val)
	r, err := h.service.Get(c)
	if err != nil {
		return err
	}
	return c.JSON(r)
}

// Post handles POST /items
func (h *Handler) Post(c *server.Context) error {
	r, err := h.service.Create(c)
	if err != nil {
		return err
	}
	return c.JSON(r)
}
