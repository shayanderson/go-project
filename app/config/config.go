package config

import (
	"os"
	"strconv"
)

// Config is the global configuration for the application
var Config config

// init initializes the config
func init() {
	Config = newConfig()
}

// config is the configuration for the application
type config struct {
	// Debug is the debug mode flag
	Debug bool

	// ServerPort is the http server port
	ServerPort int
}

// newConfig creates a new config with default values
func newConfig() config {
	return config{
		Debug:      envVar("DEBUG", "0") == "1",
		ServerPort: envVarInt("PORT", 8080),
	}
}

// envVar returns the environment variable value or the fallback value if not set or empty
func envVar(key, fallback string) string {
	v, ok := os.LookupEnv(key)
	if !ok || v == "" {
		return fallback
	}
	return v
}

// envVarInt returns the environment variable value as an int or the fallback value if not set
// or empty, panics if the value is not a valid int
func envVarInt(key string, fallback int) int {
	v, ok := os.LookupEnv(key)
	if !ok || v == "" {
		return fallback
	}
	i, err := strconv.Atoi(v)
	if err != nil {
		panic("invalid int value for " + key)
	}
	return i
}
