package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestSetSessionTokenCookieSecurityMatchesRequest(t *testing.T) {
	tests := []struct {
		name          string
		proto         string
		forwardedHost string
		value         string
		expiration    time.Time
		secure        bool
	}{
		{
			name:       "creates secure session cookie for https",
			proto:      "https",
			value:      "session-token",
			expiration: time.Now().Add(24 * time.Hour),
			secure:     true,
		},
		{
			name:          "creates non-secure session cookie for localhost even behind forwarded https",
			proto:         "https",
			forwardedHost: "localhost:3300",
			value:         "session-token",
			expiration:    time.Now().Add(24 * time.Hour),
			secure:        false,
		},
		{
			name:       "creates non-secure session cookie for http",
			proto:      "http",
			value:      "session-token",
			expiration: time.Now().Add(24 * time.Hour),
			secure:     false,
		},
		{
			name:       "clears session cookie",
			proto:      "https",
			value:      "",
			expiration: time.Now().Add(-1 * time.Hour),
			secure:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, "/", nil)
			r.Header.Set("X-Forwarded-Proto", tt.proto)
			if tt.forwardedHost != "" {
				r.Header.Set("X-Forwarded-Host", tt.forwardedHost)
			}

			setSessionTokenCookie(w, r, tt.value, tt.expiration)

			cookies := w.Result().Cookies()
			if len(cookies) != 1 {
				t.Fatalf("expected exactly one cookie, got %d", len(cookies))
			}

			cookie := cookies[0]
			if cookie.Name != "session_token" {
				t.Fatalf("expected session_token cookie, got %q", cookie.Name)
			}
			if cookie.Value != tt.value {
				t.Fatalf("expected cookie value %q, got %q", tt.value, cookie.Value)
			}
			if cookie.Path != "/" {
				t.Fatalf("expected cookie path /, got %q", cookie.Path)
			}
			if !cookie.HttpOnly {
				t.Fatal("expected cookie to be HttpOnly")
			}
			if cookie.Secure != tt.secure {
				t.Fatalf("expected Secure=%v, got %v", tt.secure, cookie.Secure)
			}
			if cookie.SameSite != http.SameSiteLaxMode {
				t.Fatalf("expected SameSite=Lax, got %v", cookie.SameSite)
			}
		})
	}
}
