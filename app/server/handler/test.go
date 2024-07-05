package handler

import (
	"net/http"

	"github.com/shayanderson/go-project/app/server"
)

type TestHandler struct{}

func NewTestHandler() *TestHandler {
	return &TestHandler{}
}

func (h *TestHandler) Get(w http.ResponseWriter, r *http.Request) error {
	return server.WriteJSON(
		w,
		http.StatusOK,
		map[string]string{"message": "test"},
	)
}

func (h *TestHandler) GetEchoName(w http.ResponseWriter, r *http.Request) error {
	name := r.PathValue("name")
	return server.WriteJSON(
		w,
		http.StatusOK,
		map[string]string{"name": name},
	)
}
