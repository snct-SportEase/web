package handler

import (
	"encoding/base64"
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestEmailFromGoogleIDToken(t *testing.T) {
	t.Run("extracts verified email for matching audience", func(t *testing.T) {
		idToken := testIDToken(`{"email":"s2301059@sendai-nct.jp","email_verified":true,"aud":"client-id","exp":%d}`, time.Now().Add(time.Hour).Unix())

		email, err := emailFromGoogleIDToken(idToken, "client-id")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if email != "s2301059@sendai-nct.jp" {
			t.Fatalf("expected email to match, got %q", email)
		}
	})

	t.Run("rejects audience mismatch", func(t *testing.T) {
		idToken := testIDToken(`{"email":"s2301059@sendai-nct.jp","email_verified":true,"aud":"other-client","exp":%d}`, time.Now().Add(time.Hour).Unix())

		if _, err := emailFromGoogleIDToken(idToken, "client-id"); err == nil {
			t.Fatal("expected audience mismatch error")
		}
	})

	t.Run("rejects unverified email", func(t *testing.T) {
		idToken := testIDToken(`{"email":"s2301059@sendai-nct.jp","email_verified":false,"aud":"client-id","exp":%d}`, time.Now().Add(time.Hour).Unix())

		if _, err := emailFromGoogleIDToken(idToken, "client-id"); err == nil {
			t.Fatal("expected unverified email error")
		}
	})
}

func testIDToken(payloadFormat string, exp int64) string {
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"none"}`))
	payload := base64.RawURLEncoding.EncodeToString([]byte(fmt.Sprintf(payloadFormat, exp)))
	return strings.Join([]string{header, payload, "signature"}, ".")
}
