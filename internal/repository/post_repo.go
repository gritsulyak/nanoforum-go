package repository

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/gritsulyak/nanoforum-go/internal/config"
	"github.com/gritsulyak/nanoforum-go/internal/models"
)

type PostRepo struct {
	db    *sql.DB
	cache *postsCache
}

func NewPostRepo(db *sql.DB) *PostRepo {
	return &PostRepo{
		db:    db,
		cache: newPostsCache(config.PostsCacheTTL()),
	}
}

type postsCache struct {
	mu      sync.Mutex
	ttl     time.Duration
	entries map[string]postCacheEntry
}

type postCacheEntry struct {
	result PostListResult
	expiry time.Time
}

func newPostsCache(ttl time.Duration) *postsCache {
	return &postsCache{
		ttl:     ttl,
		entries: make(map[string]postCacheEntry),
	}
}

func (c *postsCache) key(limit, offset int) string {
	return fmt.Sprintf("%d:%d", limit, offset)
}

func (c *postsCache) get(limit, offset int) (PostListResult, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	key := c.key(limit, offset)
	e, ok := c.entries[key]
	if !ok {
		return PostListResult{}, false
	}
	if time.Now().After(e.expiry) {
		delete(c.entries, key)
		return PostListResult{}, false
	}
	return e.result, true
}

func (c *postsCache) set(limit, offset int, result PostListResult) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries[c.key(limit, offset)] = postCacheEntry{
		result: result,
		expiry: time.Now().Add(c.ttl),
	}
}

func (c *postsCache) invalidate() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries = make(map[string]postCacheEntry)
}

func (r *PostRepo) Create(ctx context.Context, username, content string) error {
	_, err := r.db.ExecContext(ctx, "INSERT INTO posts (username, content) VALUES (?, ?)", username, content)
	if err == nil {
		r.cache.invalidate()
	}
	return err
}

// PostListResult holds the posts and a flag indicating if more posts exist.
type PostListResult struct {
	Posts   []models.Post
	HasNext bool
}

// List retrieves posts with pagination.
// Results are cached per (limit, offset) for the configured TTL
// and invalidated whenever a new post is created.
// We fetch limit + 1 to efficiently determine if a next page exists.
func (r *PostRepo) List(ctx context.Context, limit, offset int) (PostListResult, error) {
	if res, ok := r.cache.get(limit, offset); ok {
		return res, nil
	}

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

	result := PostListResult{Posts: posts, HasNext: hasNext}
	if err := rows.Err(); err != nil {
		return PostListResult{}, err
	}

	r.cache.set(limit, offset, result)
	return result, nil
}
