package models

import (
	"testing"
	"time"
)

func TestStructs(t *testing.T) {
	now := time.Date(2026, 8, 9, 12, 0, 0, 0, time.UTC)

	u := User{ID: 7, Username: "alice", PasswordHash: "$2a$10$abc"}
	if u.ID != 7 || u.Username != "alice" || u.PasswordHash != "$2a$10$abc" {
		t.Errorf("User fields not set: %+v", u)
	}

	p := Post{ID: 3, Username: "alice", Content: "hello", CreatedAt: now}
	if p.ID != 3 || p.Username != "alice" || p.Content != "hello" || !p.CreatedAt.Equal(now) {
		t.Errorf("Post fields not set: %+v", p)
	}
}
