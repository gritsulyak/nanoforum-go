package handlers

import (
	"html/template"
	"net/http"
	"strconv"

	"github.com/gritsulyak/nanoforum-go/internal/auth"
	"github.com/gritsulyak/nanoforum-go/internal/config"
	"github.com/gritsulyak/nanoforum-go/internal/models"
	"github.com/gritsulyak/nanoforum-go/internal/repository"
)

type UserStore interface {
	GetPasswordHash(username string) (string, error)
}

type PostStore interface {
	Create(username, content string) error
	List(limit, offset int) (repository.PostListResult, error)
}

type Handler struct {
	users    UserStore
	posts    PostStore
	tmpl     *template.Template
	basePath string
}

func New(users UserStore, posts PostStore, tmpl *template.Template, basePath string) *Handler {
	return &Handler{users: users, posts: posts, tmpl: tmpl, basePath: basePath}
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
			if err := h.posts.Create(username, content); err != nil {
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

	result, err := h.posts.List(pageSize, offset)
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
	_ = h.tmpl.Execute(w, PageData{
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
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, h.basePath+"/", http.StatusSeeOther)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")

	hash, err := h.users.GetPasswordHash(username)
	if err != nil {
		http.Error(w, "invalid username or password", http.StatusUnauthorized)
		return
	}

	if err := auth.CheckPassword(hash, password); err != nil {
		http.Error(w, "invalid username or password", http.StatusUnauthorized)
		return
	}

	auth.SetSession(w, username)
	http.Redirect(w, r, h.basePath+"/", http.StatusSeeOther)
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	auth.ClearSession(w)
	http.Redirect(w, r, h.basePath+"/", http.StatusSeeOther)
}
