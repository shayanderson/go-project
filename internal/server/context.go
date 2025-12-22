package server

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"strings"
	"sync/atomic"
)

// LimitReadSize is the maximum size of a request body that will be read
// defaults to 10 MB
// set to 0 to disable limit
var LimitReadSize int64 = 10 * 1024 * 1024 // 10 MB

// responseWriter is a wrapper around http.ResponseWriter that tracks if the header has been written
type responseWriter struct {
	http.ResponseWriter
	headerWritten *atomic.Bool
}

// Flush implements the http.Flusher interface
func (w *responseWriter) Flush() {
	if f, ok := w.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

// Hijack implements the http.Hijacker interface
func (w *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	h, ok := w.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, errors.New("hijacker not supported")
	}
	return h.Hijack()
}

// Push implements the http.Pusher interface
func (w *responseWriter) Push(target string, opts *http.PushOptions) error {
	if p, ok := w.ResponseWriter.(http.Pusher); ok {
		return p.Push(target, opts)
	}
	return http.ErrNotSupported
}

// Write writes the given bytes to the response
func (w *responseWriter) Write(b []byte) (int, error) {
	if w.headerWritten.CompareAndSwap(false, true) {
		w.ResponseWriter.WriteHeader(http.StatusOK)
	}
	return w.ResponseWriter.Write(b)
}

// WriteHeader writes the HTTP status code to the response
func (w *responseWriter) WriteHeader(status int) {
	if w.headerWritten.CompareAndSwap(false, true) {
		w.ResponseWriter.WriteHeader(status)
		return
	}
	// ignore duplicate header writes
}

// Context represents the context of an HTTP request
type Context struct {
	ctx     context.Context
	isMW    bool
	request *http.Request
	writer  http.ResponseWriter
}

// newContext creates a new Context
func newContext(w http.ResponseWriter, r *http.Request) *Context {
	written := &atomic.Bool{}
	return &Context{
		ctx:     r.Context(),
		request: r,
		writer:  &responseWriter{ResponseWriter: w, headerWritten: written},
	}
}

// Bind binds the request body as JSON to the given struct
func (c *Context) Bind(v any) error {
	ct := c.request.Header.Get("Content-Type")
	if !strings.HasPrefix(ct, "application/json") {
		return Error(http.StatusBadRequest, "invalid content type, expected application/json")
	}
	var dec *json.Decoder
	if LimitReadSize > 0 {
		dec = json.NewDecoder(io.LimitReader(c.request.Body, LimitReadSize))
	} else {
		dec = json.NewDecoder(c.request.Body)
	}
	return dec.Decode(v)
}

// Context returns the underlying context.Context
func (c *Context) Context() context.Context {
	return c.ctx
}

// Get retrieves a value from the context by key
func (c *Context) Get(key any) any {
	return c.ctx.Value(key)
}

// isMiddleware returns true if the context is being used in middleware
func (c *Context) isMiddleware() bool {
	return c.isMW
}

// JSON writes the given value as JSON to the response
// if an error is provided, it returns that error instead
// URL query parameter "pretty" can be used to pretty-print the JSON
func (c *Context) JSON(v any, err ...error) error {
	w := c.Writer()
	w.Header().Set("Content-Type", "application/json")

	if len(err) > 0 && err[0] != nil {
		return err[0] // return the error instead
	}

	enc := json.NewEncoder(w)
	if pretty := c.request.URL.Query().Has("pretty"); pretty {
		enc.SetIndent("", "  ")
	}
	return enc.Encode(v)
}

// middleware marks the context as being used in middleware
func (c *Context) middleware() {
	c.isMW = true
}

// Param retrieves a path parameter by key
// returns an error if the parameter is not found
func (c *Context) Param(key string) (string, error) {
	v := c.request.PathValue(key)
	if v == "" {
		return "", statusError{
			err:    errors.New("required param not found: " + key),
			status: http.StatusBadRequest,
		}
	}
	return v, nil
}

// Request returns the underlying http.Request
func (c *Context) Request() *http.Request {
	return c.request
}

// Set sets a value in the context by key
func (c *Context) Set(key, value any) {
	c.ctx = context.WithValue(c.ctx, key, value)
	c.request = c.request.WithContext(c.ctx)
}

// Status sets the HTTP status code for the response
func (c *Context) Status(code int) {
	c.writer.WriteHeader(code)
}

// Writer returns the underlying http.ResponseWriter
func (c *Context) Writer() http.ResponseWriter {
	return c.writer
}
