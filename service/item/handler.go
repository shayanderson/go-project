package item

import (
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
	return c.JSON(h.service.Get(c))
}

// Post handles POST /items
func (h *Handler) Post(c *server.Context) error {
	return c.JSON(h.service.Create(c))
}
