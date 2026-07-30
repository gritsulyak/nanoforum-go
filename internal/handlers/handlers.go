package handlers

import (
	"html/template"
	"net/http"

	"github.com/gritsulyak/nanoforum-go/internal/auth"
	"github.com/gritsulyak/nanoforum-go/internal/models"
	"github.com/gritsulyak/nanoforum-go/internal/repository"
)

type Handler struct {
	users    *repository.UserRepo
	posts    *repository.PostRepo
	tmpl     *template.Template
	basePath string
}

func New(users *repository.UserRepo, posts *repository.PostRepo, tmpl *template.Template, basePath string) *Handler {
	return &Handler{users: users, posts: posts, tmpl: tmpl, basePath: basePath}
}

type PageData struct {
	CurrentUser string
	Posts       []models.Post
	BasePath    string
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
		http.Redirect(w, r, h.basePath+"/", http.StatusSeeOther)
		return
	}

	posts, err := h.posts.List()
	if err != nil {
		http.Error(w, "can't load posts", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = h.tmpl.Execute(w, PageData{CurrentUser: username, Posts: posts, BasePath: h.basePath})
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
