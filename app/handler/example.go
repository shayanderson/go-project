package handler

import (
	"net/http"

	"github.com/shayanderson/go-project/server"
)

type ExampleHandler struct{}

func NewExampleHandler() *ExampleHandler {
	return &ExampleHandler{}
}

func (h *ExampleHandler) Get(w http.ResponseWriter, r *http.Request) error {
	return server.WriteJSON(
		w,
		http.StatusOK,
		map[string]string{"message": "example"},
	)
}

func (h *ExampleHandler) GetEchoName(w http.ResponseWriter, r *http.Request) error {
	name := r.PathValue("name")
	return server.WriteJSON(
		w,
		http.StatusOK,
		map[string]string{"name": name},
	)
}
