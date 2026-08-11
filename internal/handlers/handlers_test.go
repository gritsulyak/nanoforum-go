package handlers

import (
	"context"
	"errors"
	"html/template"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gritsulyak/nanoforum-go/internal/auth"
	"github.com/gritsulyak/nanoforum-go/internal/models"
	"github.com/gritsulyak/nanoforum-go/internal/repository"
)

type mockPostStore struct {
	listRes   repository.PostListResult
	listErr   error
	createErr error

	gotLimit  int
	gotOffset int
	created   []string
}

func (m *mockPostStore) Create(ctx context.Context, username, content string) error {
	m.created = append(m.created, username+":"+content)
	return m.createErr
}

func (m *mockPostStore) List(ctx context.Context, limit, offset int) (repository.PostListResult, error) {
	m.gotLimit, m.gotOffset = limit, offset
	return m.listRes, m.listErr
}

type mockUserStore struct {
	hash string
	err  error
}

func (m *mockUserStore) GetPasswordHash(ctx context.Context, username string) (string, error) {
	return m.hash, m.err
}

func newTestHandler(posts *mockPostStore, users *mockUserStore, basePath string) *Handler {
	tmpl := template.Must(template.New("index").Parse(`<html>{{.CurrentUser}}{{range .Posts}}@{{.Username}}{{end}}</html>`))
	h := New(users, posts, tmpl, basePath)
	h.debug = false
	return h
}

func newDebugHandler(posts *mockPostStore, users *mockUserStore, basePath string) *Handler {
	h := newTestHandler(posts, users, basePath)
	h.debug = true
	return h
}

func TestForumGetDebugTiming(t *testing.T) {
	posts := &mockPostStore{
		listRes: repository.PostListResult{Posts: []models.Post{{ID: 1, Username: "alice", Content: "hello"}}},
	}
	h := newDebugHandler(posts, &mockUserStore{}, "")
	rec := doRequest(t, h, http.MethodGet, "/", nil, false)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if !strings.Contains(rec.Body.String(), "@alice") {
		t.Errorf("body = %q, want it to contain @alice", rec.Body.String())
	}
}

func TestForumGetTemplateError(t *testing.T) {
	tmpl := template.Must(template.New("index").Parse(`{{.NoSuchField}}`))
	posts := &mockPostStore{
		listRes: repository.PostListResult{Posts: []models.Post{{ID: 1, Username: "alice"}}},
	}
	h := &Handler{users: &mockUserStore{}, posts: posts, tmpl: tmpl, basePath: ""}
	rec := doRequest(t, h, http.MethodGet, "/", nil, false)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d (template error must not break the handler)", rec.Code, http.StatusOK)
	}
}

func doRequest(t *testing.T, h *Handler, method, target string, form url.Values, loggedIn bool) *httptest.ResponseRecorder {
	t.Helper()
	var body *strings.Reader
	if form != nil {
		body = strings.NewReader(form.Encode())
	} else {
		body = strings.NewReader("")
	}
	req := httptest.NewRequest(method, target, body)
	if form != nil {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	if loggedIn {
		req.AddCookie(&http.Cookie{Name: "session_user", Value: "alice"})
	}
	rec := httptest.NewRecorder()
	switch target {
	case "/login":
		h.Login(rec, req)
	case "/logout":
		h.Logout(rec, req)
	default:
		h.Forum(rec, req)
	}
	return rec
}

func TestForumGet(t *testing.T) {
	post := models.Post{ID: 1, Username: "alice", Content: "hello"}

	tests := []struct {
		name       string
		path       string
		loggedIn   bool
		listRes    repository.PostListResult
		listErr    error
		wantStatus int
		wantBody   string
		wantPage   int
		wantOffset int
	}{
		{
			name:       "default page",
			path:       "/",
			listRes:    repository.PostListResult{Posts: []models.Post{post}, HasNext: true},
			wantStatus: http.StatusOK,
			wantBody:   "@alice",
			wantPage:   1,
		},
		{
			name:       "logged in",
			path:       "/",
			loggedIn:   true,
			listRes:    repository.PostListResult{},
			wantStatus: http.StatusOK,
			wantBody:   "alice",
			wantPage:   1,
		},
		{
			name:       "page two",
			path:       "/?page=2",
			listRes:    repository.PostListResult{},
			wantStatus: http.StatusOK,
			wantPage:   2,
			wantOffset: 10,
		},
		{
			name:       "invalid page falls back",
			path:       "/?page=abc",
			listRes:    repository.PostListResult{},
			wantStatus: http.StatusOK,
			wantPage:   1,
			wantOffset: 0,
		},
		{
			name:       "non-positive page falls back",
			path:       "/?page=0",
			listRes:    repository.PostListResult{},
			wantStatus: http.StatusOK,
			wantPage:   1,
			wantOffset: 0,
		},
		{
			name:       "list error",
			path:       "/",
			listErr:    errors.New("boom"),
			wantStatus: http.StatusInternalServerError,
			wantBody:   "can't load posts",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			posts := &mockPostStore{listRes: tt.listRes, listErr: tt.listErr}
			h := newTestHandler(posts, &mockUserStore{}, "")
			rec := doRequest(t, h, http.MethodGet, tt.path, nil, tt.loggedIn)

			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantStatus)
			}
			if tt.wantBody != "" && !strings.Contains(rec.Body.String(), tt.wantBody) {
				t.Errorf("body = %q, want it to contain %q", rec.Body.String(), tt.wantBody)
			}
			if tt.wantPage != 0 && posts.gotOffset != tt.wantOffset {
				t.Errorf("List() offset = %d, want %d", posts.gotOffset, tt.wantOffset)
			}
			if tt.listErr == nil && posts.gotLimit != 10 {
				t.Errorf("List() limit = %d, want %d", posts.gotLimit, 10)
			}
		})
	}
}

func TestForumPost(t *testing.T) {
	tests := []struct {
		name         string
		loggedIn     bool
		content      string
		createErr    error
		basePath     string
		wantStatus   int
		wantLocation string
		wantBody     string
	}{
		{
			name:       "not logged in",
			content:    "hello",
			wantStatus: http.StatusUnauthorized,
			wantBody:   "auth required",
		},
		{
			name:         "empty content",
			loggedIn:     true,
			wantStatus:   http.StatusSeeOther,
			wantLocation: "/",
		},
		{
			name:         "create success with base path",
			loggedIn:     true,
			content:      "hello world",
			basePath:     "/forum",
			wantStatus:   http.StatusSeeOther,
			wantLocation: "/forum/",
		},
		{
			name:       "create error",
			loggedIn:   true,
			content:    "hello",
			createErr:  errors.New("boom"),
			wantStatus: http.StatusInternalServerError,
			wantBody:   "can't create post",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			posts := &mockPostStore{createErr: tt.createErr}
			h := newTestHandler(posts, &mockUserStore{}, tt.basePath)
			form := url.Values{}
			if tt.content != "" {
				form.Set("content", tt.content)
			}
			rec := doRequest(t, h, http.MethodPost, "/", form, tt.loggedIn)

			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantStatus)
			}
			if tt.wantStatus == http.StatusSeeOther {
				if loc := rec.Header().Get("Location"); loc != tt.wantLocation {
					t.Errorf("Location = %q, want %q", loc, tt.wantLocation)
				}
			}
			if tt.wantBody != "" && !strings.Contains(rec.Body.String(), tt.wantBody) {
				t.Errorf("body = %q, want it to contain %q", rec.Body.String(), tt.wantBody)
			}
			if tt.wantStatus == http.StatusSeeOther && tt.content != "" {
				if len(posts.created) != 1 {
					t.Errorf("Create() called %d times, want 1", len(posts.created))
				}
			}
		})
	}
}

func TestLogin(t *testing.T) {
	validHash, err := auth.HashPassword("correct")
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name       string
		method     string
		username   string
		password   string
		userErr    error
		wantStatus int
		wantCookie string
	}{
		{
			name:       "GET redirects",
			method:     http.MethodGet,
			wantStatus: http.StatusSeeOther,
		},
		{
			name:       "success",
			method:     http.MethodPost,
			username:   "alice",
			password:   "correct",
			wantStatus: http.StatusSeeOther,
			wantCookie: "alice",
		},
		{
			name:       "user not found",
			method:     http.MethodPost,
			username:   "ghost",
			password:   "correct",
			userErr:    errors.New("user not found"),
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "wrong password",
			method:     http.MethodPost,
			username:   "alice",
			password:   "wrong",
			wantStatus: http.StatusUnauthorized,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			users := &mockUserStore{hash: validHash, err: tt.userErr}
			h := newTestHandler(&mockPostStore{}, users, "")
			form := url.Values{"username": {tt.username}, "password": {tt.password}}
			rec := doRequest(t, h, tt.method, "/login", form, false)

			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantStatus)
			}
			if tt.wantCookie != "" {
				cookies := rec.Result().Cookies()
				found := false
				for _, c := range cookies {
					if c.Name == "session_user" && c.Value == tt.wantCookie {
						found = true
					}
				}
				if !found {
					t.Errorf("expected session_user cookie = %q, got %v", tt.wantCookie, cookies)
				}
			}
		})
	}
}

func TestLogout(t *testing.T) {
	h := newTestHandler(&mockPostStore{}, &mockUserStore{}, "/forum")
	rec := doRequest(t, h, http.MethodGet, "/logout", nil, true)

	if rec.Code != http.StatusSeeOther {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusSeeOther)
	}
	if loc := rec.Header().Get("Location"); loc != "/forum/" {
		t.Errorf("Location = %q, want %q", loc, "/forum/")
	}
	found := false
	for _, c := range rec.Result().Cookies() {
		if c.Name == "session_user" && c.Value == "" {
			found = true
		}
	}
	if !found {
		t.Error("expected cleared session_user cookie")
	}
}
