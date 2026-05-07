package env

import (
	"os"
	"strconv"
	"strings"
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

// MustStrings retrieves a slice of strings from the environment variable by key, split by comma
// panics if the variable is missing or empty
func MustStrings(key string) []string {
	v, ok := os.LookupEnv(key)
	if !ok || v == "" {
		panic("required env var missing or empty: " + key)
	}
	return splitAndTrim(v, ",")
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

// Strings retrieves a slice of strings from the environment variable by key, split by comma
func Strings(key string, fallback []string) []string {
	v, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}
	return splitAndTrim(v, ",")
}

// splitAndTrim splits a string by the given separator and trims whitespace from each element
func splitAndTrim(s, sep string) []string {
	parts := make([]string, 0)
	for v := range strings.SplitSeq(s, sep) {
		v = strings.TrimSpace(v)
		if v != "" {
			parts = append(parts, v)
		}
	}
	return parts
}
