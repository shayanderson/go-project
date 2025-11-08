package tests

import (
	"github.com/shayanderson/go-project/v2/internal/server"
	"github.com/shayanderson/go-project/v2/service"
)

// test_test.go - test helpers

// newTestServer creates a test server with the API routes registered
func newTestServer() *server.TestServer {
	srv := server.NewTestServer()
	service.NewAPI(srv, service.Infra{
		// use a mock item store for testing
		ItemStore: newMockItemStore(),
	})
	return srv
}
