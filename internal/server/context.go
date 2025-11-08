package server

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"sync/atomic"
)

// LimitReadSize is the maximum size of a request body that will be read
// defaults to 10 MB
// set to 0 to disable limit
var LimitReadSize int64 = 10 * 1024 * 1024 // 10 MB

// Context represents the context of an HTTP request
type Context struct {
	ctx         context.Context
	isMW        bool
	request     *http.Request
	writer      http.ResponseWriter
	wroteStatus atomic.Bool
}

// NewContext creates a new Context
func NewContext(w http.ResponseWriter, r *http.Request) *Context {
	return &Context{
		ctx:     r.Context(),
		request: r,
		writer:  w,
	}
}

// Bind binds the request body as JSON to the given struct
func (c *Context) Bind(v any) error {
	if c.request.Header.Get("Content-Type") != "application/json" {
		return statusError{
			err:    errors.New("invalid content type, expected application/json"),
			status: http.StatusBadRequest,
		}
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

// isMiddleware returns true if the context is being used in middleware
func (c *Context) isMiddleware() bool {
	return c.isMW
}

// JSON writes the given value as JSON to the response
// if an error is provided, it returns that error instead
// URL query parameter "pretty" can be used to pretty-print the JSON
func (c *Context) JSON(v any, err ...error) error {
	if len(err) > 0 && err[0] != nil {
		return err[0]
	}

	w := c.Writer()
	w.Header().Set("Content-Type", "application/json")
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

// Status sets the HTTP status code for the response
func (c *Context) Status(code int) {
	if c.wroteStatus.Load() {
		slog.Warn(
			"status code already written, ignoring additional status code",
			slog.Int("code", code),
		)
		return
	}
	c.writer.WriteHeader(code)
	c.wroteStatus.Store(true)
}

// Writer returns the underlying http.ResponseWriter
func (c *Context) Writer() http.ResponseWriter {
	return c.writer
}
