package config

import (
	"os"
	"testing"
)

// unset removes key from the environment and restores the previous value
// when the test finishes.
func unset(t *testing.T, key string) {
	t.Helper()
	t.Setenv(key, "placeholder-for-cleanup")
	_ = os.Unsetenv(key)
}

func TestDBPath(t *testing.T) {
	tests := []struct {
		name string
		env  string
		want string
	}{
		{"unset", "", "./forum.db"},
		{"custom", "/app/data/forum.db", "/app/data/forum.db"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("DB_PATH", tt.env)
			if got := DBPath(); got != tt.want {
				t.Errorf("DBPath() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestBasePath(t *testing.T) {
	tests := []struct {
		name string
		env  string
		want string
	}{
		{"empty", "", "/"},
		{"with slashes", "/forum/", "/forum"},
		{"no slashes", "forum", "/forum"},
		{"root", "/", "/"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("BASE_PATH", tt.env)
			if got := BasePath(); got != tt.want {
				t.Errorf("BasePath() = %q, want %q", got, tt.want)
			}
		})
	}

	t.Run("unset", func(t *testing.T) {
		unset(t, "BASE_PATH")
		if got := BasePath(); got != "" {
			t.Errorf("BasePath() = %q, want %q", got, "")
		}
	})
}

func TestPageSize(t *testing.T) {
	tests := []struct {
		name string
		env  string
		want int
	}{
		{"valid", "25", 25},
		{"invalid int", "abc", 10},
		{"zero", "0", 10},
		{"negative", "-3", 10},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("PAGE_SIZE", tt.env)
			if got := PageSize(); got != tt.want {
				t.Errorf("PageSize() = %d, want %d", got, tt.want)
			}
		})
	}

	t.Run("unset", func(t *testing.T) {
		unset(t, "PAGE_SIZE")
		if got := PageSize(); got != 10 {
			t.Errorf("PageSize() = %d, want %d", got, 10)
		}
	})
}
