package config

import (
	"os"
	"testing"
	"time"
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

func TestPostsCacheTTL(t *testing.T) {
	tests := []struct {
		name string
		env  string
		want time.Duration
	}{
		{"unset", "", time.Second},
		{"valid", "500ms", 500 * time.Millisecond},
		{"valid seconds", "2s", 2 * time.Second},
		{"invalid", "abc", time.Second},
		{"zero", "0s", time.Second},
		{"negative", "-1s", time.Second},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("POSTS_CACHE_TTL", tt.env)
			if got := PostsCacheTTL(); got != tt.want {
				t.Errorf("PostsCacheTTL() = %v, want %v", got, tt.want)
			}
		})
	}

	t.Run("unset var", func(t *testing.T) {
		unset(t, "POSTS_CACHE_TTL")
		if got := PostsCacheTTL(); got != time.Second {
			t.Errorf("PostsCacheTTL() = %v, want %v", got, time.Second)
		}
	})
}

func TestPprofEnabled(t *testing.T) {
	tests := []struct {
		name string
		env  string
		want bool
	}{
		{"unset", "", false},
		{"enabled 1", "1", true},
		{"enabled true", "true", true},
		{"enabled yes", "yes", true},
		{"enabled on", "on", true},
		{"enabled upper", "TRUE", true},
		{"disabled 0", "0", false},
		{"disabled false", "false", false},
		{"whitespace", " 1 ", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("PPROF", tt.env)
			if got := PprofEnabled(); got != tt.want {
				t.Errorf("PprofEnabled() = %v, want %v", got, tt.want)
			}
		})
	}

	t.Run("unset", func(t *testing.T) {
		unset(t, "PPROF")
		if got := PprofEnabled(); got {
			t.Errorf("PprofEnabled() = %v, want false", got)
		}
	})
}

func TestDebug(t *testing.T) {
	tests := []struct {
		name string
		env  string
		want bool
	}{
		{"unset", "", false},
		{"enabled 1", "1", true},
		{"enabled true", "true", true},
		{"disabled 0", "0", false},
		{"disabled false", "false", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("DEBUG", tt.env)
			if got := Debug(); got != tt.want {
				t.Errorf("Debug() = %v, want %v", got, tt.want)
			}
		})
	}

	t.Run("unset", func(t *testing.T) {
		unset(t, "DEBUG")
		if got := Debug(); got {
			t.Errorf("Debug() = %v, want false", got)
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
