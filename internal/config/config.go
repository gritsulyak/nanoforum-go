package config

import (
	"os"
	"strconv"
	"strings"
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
