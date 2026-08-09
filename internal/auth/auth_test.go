package auth

import (
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestHashPassword(t *testing.T) {
	for _, pw := range []string{"secret", "пароль", "longer password with spaces 123", "unicode ñ €"} {
		pw := pw
		t.Run(pw, func(t *testing.T) {
			hash, err := HashPassword(pw)
			if err != nil {
				t.Fatalf("HashPassword() error = %v", err)
			}
			if hash == "" || hash == pw {
				t.Fatal("HashPassword() returned empty or plaintext hash")
			}
			if err := CheckPassword(hash, pw); err != nil {
				t.Errorf("CheckPassword() with correct password returned error: %v", err)
			}
		})
	}
}

func TestHashPasswordTooLong(t *testing.T) {
	if _, err := HashPassword(strings.Repeat("a", 73)); err == nil {
		t.Fatal("HashPassword() expected error for password longer than 72 bytes")
	}
}

func TestCheckPasswordErrors(t *testing.T) {
	hash, err := HashPassword("correct")
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name     string
		hash     string
		password string
	}{
		{"wrong password", hash, "wrong"},
		{"invalid hash", "not-a-bcrypt-hash", "correct"},
		{"empty password", hash, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := CheckPassword(tt.hash, tt.password); err == nil {
				t.Fatal("CheckPassword() expected error, got nil")
			}
		})
	}
}

func TestCurrentUser(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: "alice"})

	if got := CurrentUser(req); got != "alice" {
		t.Errorf("CurrentUser() with cookie = %q, want %q", got, "alice")
	}

	if got := CurrentUser(httptest.NewRequest(http.MethodGet, "/", nil)); got != "" {
		t.Errorf("CurrentUser() without cookie = %q, want %q", got, "")
	}
}

func TestSetSession(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	SetSession(rec, req, "alice")

	cookie := findSessionCookie(t, rec.Result().Cookies())
	if cookie.Value != "alice" {
		t.Errorf("cookie value = %q, want %q", cookie.Value, "alice")
	}
	if !cookie.HttpOnly {
		t.Error("cookie not HttpOnly")
	}
	if cookie.SameSite != http.SameSiteLaxMode {
		t.Errorf("cookie SameSite = %v, want %v", cookie.SameSite, http.SameSiteLaxMode)
	}
	if cookie.Path != "/" {
		t.Errorf("cookie Path = %q, want %q", cookie.Path, "/")
	}
	if cookie.Secure {
		t.Error("cookie Secure set over plain HTTP")
	}
	if !cookie.Expires.After(time.Now()) {
		t.Error("cookie Expires not in the future")
	}
}

func TestSetSessionTLS(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.TLS = &tls.ConnectionState{}
	SetSession(rec, req, "alice")

	cookie := findSessionCookie(t, rec.Result().Cookies())
	if !cookie.Secure {
		t.Error("cookie not Secure over TLS")
	}
}

func TestClearSession(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	ClearSession(rec, req)

	cookie := findSessionCookie(t, rec.Result().Cookies())
	if cookie.Value != "" {
		t.Errorf("cookie value = %q, want empty", cookie.Value)
	}
	if cookie.Secure {
		t.Error("cookie Secure set over plain HTTP")
	}
	if !cookie.Expires.Before(time.Now()) {
		t.Error("cookie Expires not in the past")
	}
}

func TestClearSessionTLS(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.TLS = &tls.ConnectionState{}
	ClearSession(rec, req)

	cookie := findSessionCookie(t, rec.Result().Cookies())
	if !cookie.Secure {
		t.Error("cookie not Secure over TLS")
	}
}

func findSessionCookie(t *testing.T, cookies []*http.Cookie) *http.Cookie {
	t.Helper()
	for _, c := range cookies {
		if c.Name == sessionCookieName {
			return c
		}
	}
	t.Fatalf("cookie %q not found in %v", sessionCookieName, cookies)
	return nil
}
