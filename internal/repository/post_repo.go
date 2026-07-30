package repository

import (
	"database/sql"
	"log"

	"github.com/gritsulyak/nanoforum-go/internal/models"
)

type PostRepo struct{ db *sql.DB }

func NewPostRepo(db *sql.DB) *PostRepo { return &PostRepo{db: db} }

func (r *PostRepo) Create(username, content string) error {
	_, err := r.db.Exec("INSERT INTO posts (username, content) VALUES (?, ?)", username, content)
	return err
}

func (r *PostRepo) List() ([]models.Post, error) {
	rows, err := r.db.Query("SELECT id, username, content FROM posts ORDER BY id DESC")
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("error closing rows: %v", err)
		}
	}()

	var posts []models.Post
	for rows.Next() {
		var p models.Post
		if err := rows.Scan(&p.ID, &p.Username, &p.Content); err != nil {
			return nil, err
		}
		posts = append(posts, p)
	}
	return posts, rows.Err()
}
