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

// PostListResult holds the posts and a flag indicating if more posts exist.
type PostListResult struct {
	Posts   []models.Post
	HasNext bool
}

// List retrieves posts with pagination.
// We fetch limit + 1 to efficiently determine if a next page exists.
func (r *PostRepo) List(limit, offset int) (PostListResult, error) {
	queryLimit := limit + 1
	rows, err := r.db.Query("SELECT id, username, content FROM posts ORDER BY id DESC LIMIT ? OFFSET ?", queryLimit, offset)
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
		if err := rows.Scan(&p.ID, &p.Username, &p.Content); err != nil {
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
