package repository

import (
	"context"
	"database/sql"
	"errors"
)

type UserRepo struct{ db *sql.DB }

func NewUserRepo(db *sql.DB) *UserRepo { return &UserRepo{db: db} }

func (r *UserRepo) Create(ctx context.Context, username, passwordHash string) error {
	_, err := r.db.ExecContext(ctx, "INSERT INTO users (username, password_hash) VALUES (?, ?)", username, passwordHash)
	return err
}

func (r *UserRepo) GetPasswordHash(ctx context.Context, username string) (string, error) {
	var hash string
	err := r.db.QueryRowContext(ctx, "SELECT password_hash FROM users WHERE username = ?", username).Scan(&hash)
	if errors.Is(err, sql.ErrNoRows) {
		return "", errors.New("user not found")
	}
	return hash, err
}
