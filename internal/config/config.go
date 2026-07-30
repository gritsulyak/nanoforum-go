package config

import "os"

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
