package middleware

import (
	"log/slog"
	"net/http"
)

func ExampleMiddleware(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		slog.Info("example middleware", "path", r.URL.Path)
		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}

func ExampleHandlerMiddleware(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		slog.Info("example handler middleware", "path", r.URL.Path)
		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}
