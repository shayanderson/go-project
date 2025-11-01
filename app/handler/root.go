package handler

import "net/http"

type Root struct{}

func NewRoot() *Root {
	return &Root{}
}

func (r *Root) Index(w http.ResponseWriter, req *http.Request) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte(`{"status":"ok"}`))
	return err
}

func (r *Root) NotFound(w http.ResponseWriter, req *http.Request) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotFound)
	_, err := w.Write([]byte(`{"error":"not found"}`))
	return err
}
