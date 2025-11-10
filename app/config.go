package app

import (
	"time"

	"github.com/shayanderson/go-project/v2/internal/env"
)

// Config holds the application configuration
type Config struct {
	Debug                       bool
	HTTPBindLimitReadSize       int64
	HTTPServerAddr              string
	HTTPServerReadHeaderTimeout time.Duration
	HTTPServerReadTimeout       time.Duration
	HTTPServerWriteTimeout      time.Duration
}

// NewConfig creates a new Config instance with default values
func NewConfig() (Config, error) {
	c := Config{}

	c.Debug = env.Bool("DEBUG_MODE", false)

	// http bind limit
	c.HTTPBindLimitReadSize = 20 * 1024 * 1024 // 20 MB

	// http server
	c.HTTPServerAddr = env.String("HTTP_SERVER_ADDR", ":8080")
	c.HTTPServerReadHeaderTimeout = 3 * time.Second
	c.HTTPServerReadTimeout = 3 * time.Second
	c.HTTPServerWriteTimeout = 5 * time.Second

	return c, nil
}
