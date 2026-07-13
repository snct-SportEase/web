package push

import (
	"context"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/base64"
	"io"
	"net"
	"net/http"
	"net/netip"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"backapp/internal/models"

	webpush "github.com/SherClockHolmes/webpush-go"
)

func TestValidateSubscription(t *testing.T) {
	authKey, p256dhKey := validSubscriptionKeys(t)
	sender := NewSender(Config{})

	tests := []struct {
		name     string
		endpoint string
		auth     string
		p256dh   string
		wantErr  bool
	}{
		{name: "valid FCM endpoint", endpoint: "https://fcm.googleapis.com/fcm/send/token", auth: authKey, p256dh: p256dhKey},
		{name: "valid Apple endpoint", endpoint: "https://web.push.apple.com/push/token", auth: authKey, p256dh: p256dhKey},
		{name: "valid WNS endpoint", endpoint: "https://wns2-test.notify.windows.com/w/token", auth: authKey, p256dh: p256dhKey},
		{name: "http rejected", endpoint: "http://fcm.googleapis.com/fcm/send/token", auth: authKey, p256dh: p256dhKey, wantErr: true},
		{name: "private IP rejected", endpoint: "https://127.0.0.1/push", auth: authKey, p256dh: p256dhKey, wantErr: true},
		{name: "userinfo rejected", endpoint: "https://user@fcm.googleapis.com/push", auth: authKey, p256dh: p256dhKey, wantErr: true},
		{name: "nonstandard port rejected", endpoint: "https://fcm.googleapis.com:8443/push", auth: authKey, p256dh: p256dhKey, wantErr: true},
		{name: "unknown host rejected", endpoint: "https://example.com/push", auth: authKey, p256dh: p256dhKey, wantErr: true},
		{name: "invalid auth rejected", endpoint: "https://fcm.googleapis.com/push", auth: "bad", p256dh: p256dhKey, wantErr: true},
		{name: "invalid p256dh rejected", endpoint: "https://fcm.googleapis.com/push", auth: authKey, p256dh: "bad", wantErr: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := sender.ValidateSubscription(test.endpoint, test.auth, test.p256dh)
			if test.wantErr && err == nil {
				t.Fatal("expected validation error")
			}
			if !test.wantErr && err != nil {
				t.Fatalf("unexpected validation error: %v", err)
			}
		})
	}
}

func TestHostPolicyWildcardRequiresDotBoundary(t *testing.T) {
	policy := newHostPolicy([]string{"*.push.example"})
	if !policy.allows("region.push.example") {
		t.Fatal("expected wildcard subdomain to be allowed")
	}
	if policy.allows("push.example") {
		t.Fatal("wildcard must not allow the suffix root")
	}
	if policy.allows("evilpush.example") {
		t.Fatal("wildcard must require a dot boundary")
	}
}

func TestRestrictedDialerRejectsBlockedResolution(t *testing.T) {
	dialCalled := false
	dialer := &restrictedDialer{
		allowedHosts: newHostPolicy([]string{"push.example"}),
		resolver: resolverFunc(func(context.Context, string, string) ([]netip.Addr, error) {
			return []netip.Addr{netip.MustParseAddr("169.254.169.254")}, nil
		}),
		dialer: dialerFunc(func(context.Context, string, string) (net.Conn, error) {
			dialCalled = true
			return nil, nil
		}),
	}

	if _, err := dialer.DialContext(context.Background(), "tcp", "push.example:443"); err == nil {
		t.Fatal("expected blocked address error")
	}
	if dialCalled {
		t.Fatal("network dial must not occur for a blocked address")
	}
}

func TestRestrictedDialerRejectsMixedPublicAndPrivateResolution(t *testing.T) {
	dialCalled := false
	dialer := &restrictedDialer{
		allowedHosts: newHostPolicy([]string{"push.example"}),
		resolver: resolverFunc(func(context.Context, string, string) ([]netip.Addr, error) {
			return []netip.Addr{
				netip.MustParseAddr("8.8.8.8"),
				netip.MustParseAddr("10.0.0.10"),
			}, nil
		}),
		dialer: dialerFunc(func(context.Context, string, string) (net.Conn, error) {
			dialCalled = true
			return nil, nil
		}),
	}

	if _, err := dialer.DialContext(context.Background(), "tcp", "push.example:443"); err == nil {
		t.Fatal("expected mixed resolution to be rejected")
	}
	if dialCalled {
		t.Fatal("network dial must not occur when any resolved address is blocked")
	}
}

func TestNewSenderRejectsRedirects(t *testing.T) {
	service := NewSender(Config{}).(*sender)
	client := service.client.(*http.Client)
	request, err := http.NewRequest(http.MethodGet, "https://fcm.googleapis.com/redirect", nil)
	if err != nil {
		t.Fatalf("create request: %v", err)
	}
	if err := client.CheckRedirect(request, nil); err == nil {
		t.Fatal("expected redirects to be rejected")
	}
}

func TestBlockedIPRanges(t *testing.T) {
	for _, address := range []string{
		"127.0.0.1",
		"10.0.0.1",
		"100.100.100.200",
		"169.254.169.254",
		"192.168.1.1",
		"::1",
		"fd00::1",
		"fe80::1",
		"64:ff9b::7f00:1",
		"2002:7f00:1::",
	} {
		if !isBlockedIP(netip.MustParseAddr(address)) {
			t.Fatalf("address %s must be blocked", address)
		}
	}
	if isBlockedIP(netip.MustParseAddr("8.8.8.8")) {
		t.Fatal("public address must not be blocked")
	}
}

func TestRestrictedDialerPinsValidatedPublicAddress(t *testing.T) {
	var dialedAddress string
	var peer net.Conn
	dialer := &restrictedDialer{
		allowedHosts: newHostPolicy([]string{"push.example"}),
		resolver: resolverFunc(func(context.Context, string, string) ([]netip.Addr, error) {
			return []netip.Addr{netip.MustParseAddr("8.8.8.8")}, nil
		}),
		dialer: dialerFunc(func(_ context.Context, _ string, address string) (net.Conn, error) {
			dialedAddress = address
			client, server := net.Pipe()
			peer = server
			return client, nil
		}),
	}

	connection, err := dialer.DialContext(context.Background(), "tcp", "push.example:443")
	if err != nil {
		t.Fatalf("unexpected dial error: %v", err)
	}
	defer connection.Close()
	defer peer.Close()
	if dialedAddress != "8.8.8.8:443" {
		t.Fatalf("dialed address = %q, want validated IP", dialedAddress)
	}
}

func TestSendBatchHonorsConcurrencyLimit(t *testing.T) {
	authKey, p256dhKey := validSubscriptionKeys(t)
	privateKey, publicKey, err := webpush.GenerateVAPIDKeys()
	if err != nil {
		t.Fatalf("generate VAPID keys: %v", err)
	}

	client := &countingHTTPClient{delay: 15 * time.Millisecond}
	allowed := newHostPolicy([]string{"push.example"})
	sender := newSender(Config{
		VAPIDPublicKey:  publicKey,
		VAPIDPrivateKey: privateKey,
		MaxConcurrency:  2,
	}, allowed, client)

	subscriptions := make([]models.PushSubscription, 6)
	for i := range subscriptions {
		subscriptions[i] = models.PushSubscription{
			ID:        i + 1,
			UserID:    "user",
			Endpoint:  "https://push.example/send/" + string(rune('a'+i)),
			AuthKey:   authKey,
			P256dhKey: p256dhKey,
		}
	}

	results := sender.SendBatch(context.Background(), []byte("payload"), subscriptions, 60)
	if len(results) != len(subscriptions) {
		t.Fatalf("result count = %d, want %d", len(results), len(subscriptions))
	}
	for _, result := range results {
		if result.Err != nil || result.StatusCode != http.StatusCreated {
			t.Fatalf("unexpected send result: status=%d err=%v", result.StatusCode, result.Err)
		}
	}
	if got := client.maximum.Load(); got > 2 {
		t.Fatalf("maximum concurrency = %d, want <= 2", got)
	}
}

func validSubscriptionKeys(t *testing.T) (string, string) {
	t.Helper()
	auth := []byte("0123456789abcdef")
	_, x, y, err := elliptic.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("generate subscription key: %v", err)
	}
	publicKey := elliptic.Marshal(elliptic.P256(), x, y)
	return base64.RawURLEncoding.EncodeToString(auth), base64.RawURLEncoding.EncodeToString(publicKey)
}

type resolverFunc func(context.Context, string, string) ([]netip.Addr, error)

func (f resolverFunc) LookupNetIP(ctx context.Context, network, host string) ([]netip.Addr, error) {
	return f(ctx, network, host)
}

type dialerFunc func(context.Context, string, string) (net.Conn, error)

func (f dialerFunc) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	return f(ctx, network, address)
}

type countingHTTPClient struct {
	current atomic.Int32
	maximum atomic.Int32
	delay   time.Duration
	mu      sync.Mutex
}

func (c *countingHTTPClient) Do(_ *http.Request) (*http.Response, error) {
	current := c.current.Add(1)
	defer c.current.Add(-1)
	for {
		maximum := c.maximum.Load()
		if current <= maximum || c.maximum.CompareAndSwap(maximum, current) {
			break
		}
	}
	time.Sleep(c.delay)
	return &http.Response{
		StatusCode: http.StatusCreated,
		Body:       io.NopCloser(strings.NewReader("")),
	}, nil
}
