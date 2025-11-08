package tests

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/shayanderson/go-project/v2/entity"
	"github.com/shayanderson/go-project/v2/infra/cache"
	"github.com/shayanderson/go-project/v2/internal/test"
	"github.com/shayanderson/go-project/v2/service/item"
)

// newMockItemStore creates a mock item store with predefined data
func newMockItemStore() item.Store {
	s := cache.New[entity.Item, int]()
	s.Put(1, entity.Item{ID: 1, Name: "test item 1"})
	return s
}

func TestItem_Get(t *testing.T) {
	t.Parallel()
	s := newTestServer() // create a new test server
	defer s.Stop()       // ensure server is stopped/cleaned up after test

	// perform `GET /items` request
	res, err := s.Client().Get(s.URL("/items"))
	// assert no error
	test.NoError(t, err)
	// ensure response body is closed after reading
	defer res.Body.Close()

	// assert status code OK
	test.Equal(t, http.StatusOK, res.StatusCode)
	// assert content type JSON
	test.Equal(t, "application/json", res.Header.Get("Content-Type"))

	// read response body
	body, err := io.ReadAll(res.Body)
	// assert no error reading body
	test.NoError(t, err)

	var r []entity.Item
	// unmarshal response body
	test.NoError(t, json.Unmarshal(body, &r))
	// assert response content has 1 item
	test.Equal(t, 1, len(r))
	// assert item id is 1
	test.Equal(t, 1, r[0].ID)
	// assert item name is "test item 1"
	test.Equal(t, "test item 1", r[0].Name)
}
