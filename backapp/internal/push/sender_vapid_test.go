package push

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"io"
	"math/big"
	"net/http"
	"strings"
	"testing"
	"time"

	"backapp/internal/models"

	webpush "github.com/SherClockHolmes/webpush-go"
)

func TestSendBatchBuildsValidVAPIDAuthorization(t *testing.T) {
	authKey, p256dhKey := validSubscriptionKeys(t)
	privateKey, publicKey, err := webpush.GenerateVAPIDKeys()
	if err != nil {
		t.Fatal(err)
	}
	publicBytes, err := base64.RawURLEncoding.DecodeString(publicKey)
	if err != nil {
		t.Fatal(err)
	}
	x, y := elliptic.Unmarshal(elliptic.P256(), publicBytes)
	verificationKey := &ecdsa.PublicKey{Curve: elliptic.P256(), X: x, Y: y}

	tests := []struct {
		name       string
		subscriber string
		wantSub    string
	}{
		{"default contact", "", "https://nitsche-gyouji.com"},
		{"blank contact", " \t", "https://nitsche-gyouji.com"},
		{"email address", "push@example.com", "mailto:push@example.com"},
		{"mailto URI", "mailto:push@example.com", "mailto:push@example.com"},
		{"duplicate prefix", "mailto:mailto:push@example.com", "mailto:push@example.com"},
		{"case and whitespace", " \tMAILTO:push@example.com \n", "mailto:push@example.com"},
		{"empty mailto", "mailto:", "https://nitsche-gyouji.com"},
		{"HTTPS contact", " https://example.com/contact ", "https://example.com/contact"},
	}
	for _, origin := range []string{"https://web.push.apple.com", "https://fcm.googleapis.com"} {
		for _, test := range tests {
			t.Run(origin+"/"+test.name, func(t *testing.T) {
				var authorization string
				client := vapidHTTPClientFunc(func(request *http.Request) (*http.Response, error) {
					authorization = request.Header.Get("Authorization")
					return &http.Response{StatusCode: http.StatusCreated, Body: io.NopCloser(strings.NewReader(""))}, nil
				})
				sender := newSender(Config{
					VAPIDPublicKey: publicKey, VAPIDPrivateKey: privateKey, Subscriber: test.subscriber,
				}, newHostPolicy(nil), client)
				before := time.Now()
				results := sender.SendBatch(context.Background(), []byte(`{"title":"test"}`), []models.PushSubscription{{
					Endpoint: origin + "/push/test-token", AuthKey: authKey, P256dhKey: p256dhKey,
				}}, 60)
				if results[0].Err != nil || results[0].StatusCode != http.StatusCreated {
					t.Fatalf("send failed: status=%d err=%v", results[0].StatusCode, results[0].Err)
				}

				// 実際にwebpush-goが生成したJWTを検証し、内部設定の比較だけで済ませない。
				if !strings.HasPrefix(authorization, "vapid t=") {
					t.Fatal("missing VAPID Authorization header")
				}
				token, encodedKey, ok := strings.Cut(strings.TrimPrefix(authorization, "vapid t="), ", k=")
				if !ok || encodedKey != publicKey {
					t.Fatal("Authorization public key does not match configured VAPID key")
				}
				parts := strings.Split(token, ".")
				if len(parts) != 3 {
					t.Fatal("invalid JWT structure")
				}
				var header struct {
					Algorithm string `json:"alg"`
				}
				decodeVAPIDJSON(t, parts[0], &header)
				if header.Algorithm != "ES256" {
					t.Fatalf("algorithm = %q, want ES256", header.Algorithm)
				}
				var claims struct {
					Subject  string `json:"sub"`
					Audience string `json:"aud"`
					Expires  int64  `json:"exp"`
				}
				decodeVAPIDJSON(t, parts[1], &claims)
				if claims.Subject != test.wantSub {
					t.Fatalf("sub = %q, want %q", claims.Subject, test.wantSub)
				}
				if claims.Audience != origin {
					t.Fatalf("aud = %q, want %q", claims.Audience, origin)
				}
				if claims.Expires <= before.Unix() || claims.Expires > time.Now().Add(24*time.Hour).Unix() {
					t.Fatal("JWT expiration must be in the next 24 hours")
				}
				signature, err := base64.RawURLEncoding.DecodeString(parts[2])
				if err != nil || len(signature) != 64 {
					t.Fatal("invalid ES256 signature encoding")
				}
				digest := sha256.Sum256([]byte(parts[0] + "." + parts[1]))
				if !ecdsa.Verify(verificationKey, digest[:], new(big.Int).SetBytes(signature[:32]), new(big.Int).SetBytes(signature[32:])) {
					t.Fatal("JWT signature does not match configured VAPID public key")
				}
			})
		}
	}
}

func decodeVAPIDJSON(t *testing.T, encoded string, target any) {
	t.Helper()
	data, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, target); err != nil {
		t.Fatal(err)
	}
}

type vapidHTTPClientFunc func(*http.Request) (*http.Response, error)

func (f vapidHTTPClientFunc) Do(request *http.Request) (*http.Response, error) {
	return f(request)
}
