package db

import (
	"database/sql"
	"errors"
	"path/filepath"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestNew(t *testing.T) {
	path := filepath.Join(t.TempDir(), "forum.db")
	conn, err := New(path)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer func() { _ = conn.Close() }()

	rows, err := conn.Query("SELECT name FROM sqlite_master WHERE type = 'table' AND name IN ('users', 'posts') ORDER BY name")
	if err != nil {
		t.Fatalf("Query() error = %v", err)
	}
	defer func() { _ = rows.Close() }()

	var tables []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			t.Fatal(err)
		}
		tables = append(tables, name)
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	if len(tables) != 2 || tables[0] != "posts" || tables[1] != "users" {
		t.Fatalf("expected tables [posts users], got %v", tables)
	}
}

func TestNewOpenError(t *testing.T) {
	orig := open
	open = func(driver, dsn string) (*sql.DB, error) { return nil, errors.New("open failed") }
	defer func() { open = orig }()

	if _, err := New(":memory:"); err == nil {
		t.Fatal("New() expected error from sql.Open, got nil")
	}
}

func TestNewMigrateError(t *testing.T) {
	tests := []struct {
		name  string
		setup func(mock sqlmock.Sqlmock)
	}{
		{
			name: "users table",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("CREATE TABLE IF NOT EXISTS users").WillReturnError(errors.New("boom"))
			},
		},
		{
			name: "posts table",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("CREATE TABLE IF NOT EXISTS users").WillReturnResult(sqlmock.NewResult(0, 0))
				mock.ExpectExec("CREATE TABLE IF NOT EXISTS posts").WillReturnError(errors.New("boom"))
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB, mock, err := sqlmock.New()
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = mockDB.Close() }()

			orig := open
			open = func(driver, dsn string) (*sql.DB, error) { return mockDB, nil }
			defer func() { open = orig }()

			tt.setup(mock)
			if _, err := New(":memory:"); err == nil {
				t.Fatal("New() expected error during migrate, got nil")
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}
