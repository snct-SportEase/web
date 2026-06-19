package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestSetSessionTokenCookieUsesSecureAttributes(t *testing.T) {
	tests := []struct {
		name       string
		value      string
		expiration time.Time
	}{
		{
			name:       "creates session cookie",
			value:      "session-token",
			expiration: time.Now().Add(24 * time.Hour),
		},
		{
			name:       "clears session cookie",
			value:      "",
			expiration: time.Now().Add(-1 * time.Hour),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()

			setSessionTokenCookie(w, tt.value, tt.expiration)

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
			if !cookie.Secure {
				t.Fatal("expected cookie to be Secure")
			}
			if cookie.SameSite != http.SameSiteLaxMode {
				t.Fatalf("expected SameSite=Lax, got %v", cookie.SameSite)
			}
		})
	}
}
