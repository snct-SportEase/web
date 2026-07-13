package handler

import (
	"context"
	"errors"
	"testing"

	"google.golang.org/api/idtoken"
)

type fakeGoogleIDTokenValidator struct {
	payload *idtoken.Payload
	err     error
}

func (f fakeGoogleIDTokenValidator) Validate(context.Context, string, string) (*idtoken.Payload, error) {
	return f.payload, f.err
}

func validGoogleIDTokenPayload() *idtoken.Payload {
	return &idtoken.Payload{
		Issuer:   "https://accounts.google.com",
		Audience: "client-id",
		Subject:  "google-subject",
		Claims: map[string]any{
			"nonce":          "expected-nonce",
			"email":          "s2301059@sendai-nct.jp",
			"email_verified": true,
			"azp":            "client-id",
		},
	}
}

func TestVerifiedGoogleIDTokenEmail(t *testing.T) {
	t.Run("accepts claims only after official validation succeeds", func(t *testing.T) {
		email, err := verifiedGoogleIDTokenEmail(
			context.Background(),
			fakeGoogleIDTokenValidator{payload: validGoogleIDTokenPayload()},
			"signed-token",
			"client-id",
			"expected-nonce",
		)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if email != "s2301059@sendai-nct.jp" {
			t.Fatalf("expected email to match, got %q", email)
		}
	})

	t.Run("fails closed when signature validation fails", func(t *testing.T) {
		_, err := verifiedGoogleIDTokenEmail(
			context.Background(),
			fakeGoogleIDTokenValidator{err: errors.New("invalid signature")},
			"unsigned-token",
			"client-id",
			"expected-nonce",
		)
		if err == nil {
			t.Fatal("expected validation error")
		}
	})

	tests := []struct {
		name   string
		mutate func(*idtoken.Payload)
	}{
		{name: "rejects invalid issuer", mutate: func(payload *idtoken.Payload) { payload.Issuer = "https://attacker.example" }},
		{name: "rejects audience mismatch", mutate: func(payload *idtoken.Payload) { payload.Audience = "other-client" }},
		{name: "rejects missing subject", mutate: func(payload *idtoken.Payload) { payload.Subject = "" }},
		{name: "rejects nonce mismatch", mutate: func(payload *idtoken.Payload) { payload.Claims["nonce"] = "replayed-nonce" }},
		{name: "rejects authorized party mismatch", mutate: func(payload *idtoken.Payload) { payload.Claims["azp"] = "other-client" }},
		{name: "rejects unverified email", mutate: func(payload *idtoken.Payload) { payload.Claims["email_verified"] = false }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			payload := validGoogleIDTokenPayload()
			test.mutate(payload)
			_, err := verifiedGoogleIDTokenEmail(
				context.Background(),
				fakeGoogleIDTokenValidator{payload: payload},
				"signed-token",
				"client-id",
				"expected-nonce",
			)
			if err == nil {
				t.Fatal("expected invalid claims to be rejected")
			}
		})
	}
}
