package server

import (
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"
	"time"
)

// Middleware is a http middleware
type Middleware func(http.Handler) http.Handler

// responseWriter is a http.ResponseWriter wrapper
type responseWriter struct {
	w      *http.ResponseWriter
	status *int
}

// Header implements the http.ResponseWriter interface
func (r responseWriter) Header() http.Header {
	return (*r.w).Header()
}

// Write implements the http.ResponseWriter interface
func (r responseWriter) Write(b []byte) (int, error) {
	return (*r.w).Write(b)
}

// WriteHeader implements the http.ResponseWriter interface
func (r responseWriter) WriteHeader(status int) {
	(*r.status) = status
	(*r.w).WriteHeader(status)
}

// LoggerMiddleware logs http requests
func LoggerMiddleware(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		status := 0
		rw := responseWriter{
			w:      &w,
			status: &status,
		}

		defer func() {
			scheme := "http"
			if r.TLS != nil {
				scheme = "https"
			}

			slog.Info(
				fmt.Sprintf(
					"[http] %s %s://%s%s %s",
					r.Method,
					scheme,
					r.Host,
					r.RequestURI,
					r.Proto,
				),
				"from", r.RemoteAddr,
				"status", *rw.status,
				"took", time.Since(start).String(),
			)
		}()

		next.ServeHTTP(rw, r)
	}

	return http.HandlerFunc(fn)
}

// RecoverMiddleware recovers from panics
func RecoverMiddleware(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.Header().Set("Connection", "close")
				slog.Error(
					"[http] recovering from panic",
					"err",
					err,
					"trace",
					string(debug.Stack()),
				)
				_ = WriteJSON(
					w,
					http.StatusInternalServerError,
					map[string]string{"error": "internal server error"},
				)
			}
		}()

		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}
