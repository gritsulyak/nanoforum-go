package handlers

import (
	"context"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gritsulyak/nanoforum-go/internal/auth"
	"github.com/gritsulyak/nanoforum-go/internal/config"
	"github.com/gritsulyak/nanoforum-go/internal/models"
	"github.com/gritsulyak/nanoforum-go/internal/repository"
)

type UserStore interface {
	GetPasswordHash(ctx context.Context, username string) (string, error)
}

type PostStore interface {
	Create(ctx context.Context, username, content string) error
	List(ctx context.Context, limit, offset int) (repository.PostListResult, error)
}

type Handler struct {
	users    UserStore
	posts    PostStore
	tmpl     *template.Template
	basePath string
	debug    bool
}

func New(users UserStore, posts PostStore, tmpl *template.Template, basePath string) *Handler {
	return &Handler{
		users:    users,
		posts:    posts,
		tmpl:     tmpl,
		basePath: basePath,
		debug:    config.Debug(),
	}
}

// logTiming emits the duration of a handler stage when DEBUG is enabled.
func (h *Handler) logTiming(stage string, start time.Time) {
	if h.debug {
		log.Printf("timing %s=%s", stage, time.Since(start))
	}
}

type PageData struct {
	CurrentUser string
	Posts       []models.Post
	BasePath    string
	// Pagination fields
	CurrentPage int
	HasPrev     bool
	HasNext     bool
	PrevPage    int
	NextPage    int
	PageSize    int
}

func (h *Handler) Forum(w http.ResponseWriter, r *http.Request) {
	username := auth.CurrentUser(r)
	if r.Method == http.MethodPost {
		if username == "" {
			http.Error(w, "auth required", http.StatusUnauthorized)
			return
		}
		content := r.FormValue("content")
		if content != "" {
			if err := h.posts.Create(r.Context(), username, content); err != nil {
				http.Error(w, "can't create post", http.StatusInternalServerError)
				return
			}
		}
		// Redirect to page 1 after posting
		http.Redirect(w, r, h.basePath+"/", http.StatusSeeOther)
		return
	}

	// Pagination logic
	pageSize := config.PageSize()
	page := 1
	if p := r.URL.Query().Get("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}
	offset := (page - 1) * pageSize

	start := time.Now()
	result, err := h.posts.List(r.Context(), pageSize, offset)
	h.logTiming("forum.posts.List", start)
	if err != nil {
		http.Error(w, "can't load posts", http.StatusInternalServerError)
		return
	}

	hasPrev := page > 1
	prevPage := page - 1
	if prevPage < 1 {
		prevPage = 1
	}
	nextPage := page + 1

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	start = time.Now()
	err = h.tmpl.Execute(w, PageData{
		CurrentUser: username,
		Posts:       result.Posts,
		BasePath:    h.basePath,
		CurrentPage: page,
		HasPrev:     hasPrev,
		HasNext:     result.HasNext,
		PrevPage:    prevPage,
		NextPage:    nextPage,
		PageSize:    pageSize,
	})
	h.logTiming("forum.tmpl.Execute", start)
	if err != nil {
		log.Printf("error executing template: %v", err)
	}
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, h.basePath+"/", http.StatusSeeOther)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")

	start := time.Now()
	hash, err := h.users.GetPasswordHash(r.Context(), username)
	h.logTiming("login.GetPasswordHash", start)
	if err != nil {
		http.Error(w, "invalid username or password", http.StatusUnauthorized)
		return
	}

	start = time.Now()
	if err := auth.CheckPassword(hash, password); err != nil {
		h.logTiming("login.CheckPassword", start)
		http.Error(w, "invalid username or password", http.StatusUnauthorized)
		return
	}
	h.logTiming("login.CheckPassword", start)

	auth.SetSession(w, r, username)
	http.Redirect(w, r, h.basePath+"/", http.StatusSeeOther)
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	auth.ClearSession(w, r)
	http.Redirect(w, r, h.basePath+"/", http.StatusSeeOther)
}
