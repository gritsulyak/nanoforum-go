package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

const defaultDBPath = "./forum.db"

// DBPath returns the database path
// from the environment variable DB_PATH
// or the default path if not set
// "./forum.db", when running locally,
// or "/app/data/forum.db"
// when running in a Docker container.
// In the container, DB_PATH is set to
// /app/data/forum.db via docker-compose.yml.
func DBPath() string {
	if v, ok := os.LookupEnv("DB_PATH"); ok && v != "" {
		return v
	}
	return defaultDBPath
}

// BasePath returns the base path for the forum, which can be set via the environment variable BASE_PATH.
// This is useful when the forum is running behind a reverse proxy (e.g., "/forum").
// By default, it returns an empty string, meaning the forum runs at the root of the domain.
func BasePath() string {
	v, ok := os.LookupEnv("BASE_PATH")
	if !ok {
		return ""
	}
	return "/" + strings.Trim(v, "/")
}

// PageSize returns the number of posts to display per page.
// It reads from the PAGE_SIZE environment variable, defaulting to 10.
func PageSize() int {
	v, ok := os.LookupEnv("PAGE_SIZE")
	if !ok {
		return 10
	}
	size, err := strconv.Atoi(v)
	if err != nil || size <= 0 {
		return 10 // Fallback to default if invalid
	}
	return size
}

// PprofEnabled reports whether the /debug/pprof profiling endpoint
// should be exposed. Controlled by the PPROF environment variable.
func PprofEnabled() bool {
	return envBool("PPROF")
}

// PostsCacheTTL returns the TTL for the posts list cache,
// read from the POSTS_CACHE_TTL environment variable (Go duration),
// defaulting to 1 second.
func PostsCacheTTL() time.Duration {
	if v, ok := os.LookupEnv("POSTS_CACHE_TTL"); ok && v != "" {
		if d, err := time.ParseDuration(v); err == nil && d > 0 {
			return d
		}
	}
	return time.Second
}

// Debug reports whether per-request latency breakdown logging
// is enabled. Controlled by the DEBUG environment variable.
func Debug() bool {
	return envBool("DEBUG")
}

func envBool(key string) bool {
	v, ok := os.LookupEnv(key)
	if !ok {
		return false
	}
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}
