package handlers

import (
	"html/template"
	"net/http"

	"github.com/gritsulyak/nanoforum-go/internal/auth"
	"github.com/gritsulyak/nanoforum-go/internal/models"
	"github.com/gritsulyak/nanoforum-go/internal/repository"
)

type PageData struct {
	CurrentUser string
	Posts       []models.Post
}

type Handler struct {
	users *repository.UserRepo
	posts *repository.PostRepo
	tmpl  *template.Template
}

func New(users *repository.UserRepo, posts *repository.PostRepo, tmpl *template.Template) *Handler {
	return &Handler{users: users, posts: posts, tmpl: tmpl}
}

func (h *Handler) Forum(w http.ResponseWriter, r *http.Request) {
	username := auth.CurrentUser(r)

	if r.Method == http.MethodPost {
		if username == "" {
			http.Error(w, "Только авторизованные пользователи могут писать", http.StatusUnauthorized)
			return
		}
		content := r.FormValue("content")
		if content != "" {
			if err := h.posts.Create(username, content); err != nil {
				http.Error(w, "Не удалось сохранить сообщение", http.StatusInternalServerError)
				return
			}
		}
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	posts, err := h.posts.List()
	if err != nil {
		http.Error(w, "Ошибка загрузки сообщений", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = h.tmpl.Execute(w, PageData{CurrentUser: username, Posts: posts})
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")

	hash, err := h.users.GetPasswordHash(username)
	if err != nil {
		http.Error(w, "Неверный логин или пароль", http.StatusUnauthorized)
		return
	}

	if err := auth.CheckPassword(hash, password); err != nil {
		http.Error(w, "Неверный логин или пароль", http.StatusUnauthorized)
		return
	}

	auth.SetSession(w, username)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	auth.ClearSession(w)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
