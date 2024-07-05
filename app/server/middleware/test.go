package middleware

import (
	"log/slog"
	"net/http"
)

func TestMiddleware(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		slog.Info("test middleware", "path", r.URL.Path)
		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}

func TestHandlerMiddleware(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		slog.Info("test handler middleware", "path", r.URL.Path)
		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}
