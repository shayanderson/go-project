package env

import (
	"os"
	"strconv"
)

// Bool retrieves the boolean value of the environment variable by key
func Bool(key string, fallback bool) bool {
	v, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}
	return parseBool(v)
}

// Int retrieves the integer value of the environment variable by key
func Int(key string, fallback int) int {
	v, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}
	i, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return i
}

// Must checks that the given environment variables are set and not empty
// panics if any are missing or empty
func Must(keys ...string) {
	for _, key := range keys {
		v, ok := os.LookupEnv(key)
		if !ok || v == "" {
			panic("required env var missing or empty: " + key)
		}
	}
}

// MustBool retrieves the boolean value of the environment variable by key
// panics if the variable is missing or empty
func MustBool(key string) bool {
	v, ok := os.LookupEnv(key)
	if !ok || v == "" {
		panic("required env var missing or empty: " + key)
	}
	return parseBool(v)
}

// MustInt retrieves the integer value of the environment variable by key
// panics if the variable is missing, empty, or invalid integer
func MustInt(key string) int {
	v, ok := os.LookupEnv(key)
	if !ok || v == "" {
		panic("required env var missing or empty: " + key)
	}
	i, err := strconv.Atoi(v)
	if err != nil {
		panic("invalid int value for env var: " + key)
	}
	return i
}

// MustString retrieves the string value of the environment variable by key
// panics if the variable is missing or empty
func MustString(key string) string {
	v, ok := os.LookupEnv(key)
	if !ok || v == "" {
		panic("required env var missing or empty: " + key)
	}
	return v
}

// parseBool parses a string into a boolean value
func parseBool(v string) bool {
	switch v {
	case "1", "true", "TRUE", "True", "yes", "YES", "on", "ON":
		return true
	}
	return false
}

// String retrieves the string value of the environment variable by key
func String(key, fallback string) string {
	v, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}
	return v
}
