package service

import (
	"github.com/shayanderson/go-project/v2/internal/assert"
	"github.com/shayanderson/go-project/v2/internal/server"
	"github.com/shayanderson/go-project/v2/service/item"
)

// Infra holds the dependencies required by the API
type Infra struct {
	// ItemStore is the item store
	ItemStore item.Store
}

// Server defines the interface for the server used by the API
type Server interface {
	Handle(string, server.HandlerFunc, ...server.Middleware)
	Start() error
	Stop() error
	Use(...server.Middleware)
}

// API represents the API service
type API struct {
	infra  Infra
	server Server
}

// NewAPI creates a new API instance
func NewAPI(srv Server, infra Infra) *API {
	assert.NotNil(srv, "server is nil")
	assert.NotNil(infra.ItemStore, "ItemStore is nil")
	a := &API{server: srv, infra: infra}

	// setup middleware
	srv.Use(ExampleMiddleware{}.Handle)

	// setup routes
	a.router()
	return a
}

// Start starts the API server
func (a *API) Start() error {
	return a.server.Start()
}

// Stop stops the API server
func (a *API) Stop() error {
	return a.server.Stop()
}
