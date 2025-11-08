package service

import "github.com/shayanderson/go-project/v2/service/item"

// router sets up the API routes
func (a *API) router() {
	// item
	itemService := item.New(a.infra.ItemStore)
	itemHandler := item.NewHandler(itemService)
	a.server.Handle("GET /items", itemHandler.Get)
	a.server.Handle("POST /items", itemHandler.Post)

	// other routes added here...
}
