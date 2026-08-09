package db

import (
	"context"
	"database/sql"

	_ "modernc.org/sqlite"
)

var open = sql.Open

func New(path string) (*sql.DB, error) {
	conn, err := open("sqlite", path)
	if err != nil {
		return nil, err
	}
	if err := migrate(conn); err != nil {
		return nil, err
	}
	return conn, nil
}

func migrate(conn *sql.DB) error {
	ctx := context.Background()
	if _, err := conn.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT UNIQUE NOT NULL,
		password_hash TEXT NOT NULL
	)`); err != nil {
		return err
	}
	_, err := conn.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS posts (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT NOT NULL,
		content TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`)
	return err
}
