package repository

import (
	"context"
	"database/sql"
	"log"

	"github.com/gritsulyak/nanoforum-go/internal/models"
)

type PostRepo struct{ db *sql.DB }

func NewPostRepo(db *sql.DB) *PostRepo { return &PostRepo{db: db} }

func (r *PostRepo) Create(ctx context.Context, username, content string) error {
	_, err := r.db.ExecContext(ctx, "INSERT INTO posts (username, content) VALUES (?, ?)", username, content)
	return err
}

// PostListResult holds the posts and a flag indicating if more posts exist.
type PostListResult struct {
	Posts   []models.Post
	HasNext bool
}

// List retrieves posts with pagination.
// We fetch limit + 1 to efficiently determine if a next page exists.
func (r *PostRepo) List(ctx context.Context, limit, offset int) (PostListResult, error) {
	queryLimit := limit + 1
	rows, err := r.db.QueryContext(ctx, "SELECT id, username, content, created_at FROM posts ORDER BY id DESC LIMIT ? OFFSET ?", queryLimit, offset)
	if err != nil {
		return PostListResult{}, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("error closing rows: %v", err)
		}
	}()

	var posts []models.Post
	for rows.Next() {
		var p models.Post
		if err := rows.Scan(&p.ID, &p.Username, &p.Content, &p.CreatedAt); err != nil {
			return PostListResult{}, err
		}
		posts = append(posts, p)
	}

	hasNext := false
	if len(posts) > limit {
		hasNext = true
		posts = posts[:limit] // Drop the extra item used for checking
	}

	return PostListResult{Posts: posts, HasNext: hasNext}, rows.Err()
}
