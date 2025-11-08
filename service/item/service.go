package item

import (
	"github.com/shayanderson/go-project/v2/entity"
	"github.com/shayanderson/go-project/v2/internal/assert"
	"github.com/shayanderson/go-project/v2/internal/server"
)

// Store defines the item storage interface
type Store interface {
	All() []entity.Item
	Put(int, entity.Item)
}

// Service handles item-related operations
type Service struct {
	store Store
}

// New creates a new item service
func New(store Store) *Service {
	assert.NotNil(store, "store is nil")
	return &Service{
		store: store,
	}
}

// Create creates a new item
func (s *Service) Create(c *server.Context) (entity.Item, error) {
	var v entity.Item
	if err := c.Bind(&v); err != nil {
		return entity.Item{}, err
	}
	s.store.Put(v.ID, v)
	return v, nil
}

// Get retrieves all items
func (s *Service) Get(c *server.Context) ([]entity.Item, error) {
	return s.store.All(), nil
}
