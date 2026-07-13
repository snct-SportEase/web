package push

import (
	"context"
	"crypto/elliptic"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"strings"
	"time"

	"backapp/internal/models"

	webpush "github.com/SherClockHolmes/webpush-go"
	"golang.org/x/sync/errgroup"
)

const (
	MaxSubscriptionRequestBytes int64 = 4 << 10
	MaxSubscriptionsPerUser           = 5
	defaultMaxConcurrency             = 32
	maxEndpointLength                 = 500
	maxErrorBodyBytes           int64 = 4 << 10
	requestTimeout                    = 10 * time.Second
)

var defaultAllowedHostPatterns = []string{
	"fcm.googleapis.com",
	"updates.push.services.mozilla.com",
	"*.push.apple.com",
	"*.notify.windows.com",
}

type Sender interface {
	Enabled() bool
	ValidateSubscription(endpoint, authKey, p256dhKey string) error
	SendBatch(ctx context.Context, payload []byte, subscriptions []models.PushSubscription, ttl int) []Result
}

type Result struct {
	Subscription        models.PushSubscription
	StatusCode          int
	InvalidSubscription bool
	Err                 error
}

type Config struct {
	VAPIDPublicKey  string
	VAPIDPrivateKey string
	AllowedHosts    []string
	Subscriber      string
	MaxConcurrency  int
}

type sender struct {
	publicKey      string
	privateKey     string
	subscriber     string
	allowedHosts   hostPolicy
	client         webpush.HTTPClient
	maxConcurrency int
}

func NewSender(cfg Config) Sender {
	allowedHosts := newHostPolicy(cfg.AllowedHosts)
	secureDialer := &restrictedDialer{
		allowedHosts: allowedHosts,
		resolver:     net.DefaultResolver,
		dialer: &net.Dialer{
			Timeout:   3 * time.Second,
			KeepAlive: 30 * time.Second,
		},
	}

	transport := &http.Transport{
		Proxy:                  nil,
		DialContext:            secureDialer.DialContext,
		ForceAttemptHTTP2:      true,
		MaxIdleConns:           defaultMaxConcurrency,
		MaxIdleConnsPerHost:    8,
		MaxConnsPerHost:        16,
		IdleConnTimeout:        30 * time.Second,
		TLSHandshakeTimeout:    3 * time.Second,
		ResponseHeaderTimeout:  5 * time.Second,
		ExpectContinueTimeout:  time.Second,
		MaxResponseHeaderBytes: 32 << 10,
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   requestTimeout,
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
			return errors.New("web push redirects are not allowed")
		},
	}

	return newSender(cfg, allowedHosts, client)
}

func newSender(cfg Config, allowedHosts hostPolicy, client webpush.HTTPClient) Sender {
	if allowedHosts.empty() {
		allowedHosts = newHostPolicy(cfg.AllowedHosts)
	}
	subscriber := strings.TrimSpace(cfg.Subscriber)
	if subscriber == "" {
		subscriber = "mailto:notifications@sportease.local"
	}
	maxConcurrency := cfg.MaxConcurrency
	if maxConcurrency <= 0 || maxConcurrency > defaultMaxConcurrency {
		maxConcurrency = defaultMaxConcurrency
	}

	return &sender{
		publicKey:      cfg.VAPIDPublicKey,
		privateKey:     cfg.VAPIDPrivateKey,
		subscriber:     subscriber,
		allowedHosts:   allowedHosts,
		client:         client,
		maxConcurrency: maxConcurrency,
	}
}

func (s *sender) Enabled() bool {
	return s.publicKey != "" && s.privateKey != ""
}

func (s *sender) ValidateSubscription(endpoint, authKey, p256dhKey string) error {
	if _, err := validateEndpoint(endpoint, s.allowedHosts); err != nil {
		return err
	}

	auth, err := decodeSubscriptionKey(authKey)
	if err != nil || len(auth) != 16 {
		return errors.New("auth key must be a 16-byte base64url value")
	}

	p256dh, err := decodeSubscriptionKey(p256dhKey)
	if err != nil || len(p256dh) != 65 || p256dh[0] != 4 {
		return errors.New("p256dh key must be an uncompressed P-256 public key")
	}
	x, y := elliptic.Unmarshal(elliptic.P256(), p256dh)
	if x == nil || y == nil || !elliptic.P256().IsOnCurve(x, y) {
		return errors.New("p256dh key is not a valid P-256 point")
	}

	return nil
}

func (s *sender) SendBatch(ctx context.Context, payload []byte, subscriptions []models.PushSubscription, ttl int) []Result {
	results := make([]Result, len(subscriptions))
	if len(subscriptions) == 0 {
		return results
	}

	var group errgroup.Group
	group.SetLimit(s.maxConcurrency)
	for i := range subscriptions {
		i := i
		group.Go(func() error {
			results[i] = s.sendOne(ctx, payload, subscriptions[i], ttl)
			return nil
		})
	}
	_ = group.Wait()
	return results
}

func (s *sender) sendOne(ctx context.Context, payload []byte, sub models.PushSubscription, ttl int) Result {
	result := Result{Subscription: sub}
	if !s.Enabled() {
		result.Err = errors.New("web push is disabled")
		return result
	}
	if err := s.ValidateSubscription(sub.Endpoint, sub.AuthKey, sub.P256dhKey); err != nil {
		result.InvalidSubscription = true
		result.Err = fmt.Errorf("invalid stored push subscription: %w", err)
		return result
	}

	requestCtx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()

	response, err := webpush.SendNotificationWithContext(requestCtx, payload, &webpush.Subscription{
		Endpoint: sub.Endpoint,
		Keys: webpush.Keys{
			Auth:   sub.AuthKey,
			P256dh: sub.P256dhKey,
		},
	}, &webpush.Options{
		HTTPClient:      s.client,
		Subscriber:      s.subscriber,
		VAPIDPublicKey:  s.publicKey,
		VAPIDPrivateKey: s.privateKey,
		TTL:             ttl,
	})
	if err != nil {
		result.Err = err
		return result
	}
	if response == nil {
		result.Err = errors.New("push service returned no response")
		return result
	}
	defer response.Body.Close()

	result.StatusCode = response.StatusCode
	if response.StatusCode >= http.StatusBadRequest {
		_, readErr := io.Copy(io.Discard, io.LimitReader(response.Body, maxErrorBodyBytes))
		if readErr != nil {
			result.Err = fmt.Errorf("read push error response: %w", readErr)
			return result
		}
	}

	return result
}

func EndpointLogID(endpoint string) string {
	host := "invalid-host"
	if parsed, err := url.Parse(endpoint); err == nil && parsed.Hostname() != "" {
		host = strings.ToLower(parsed.Hostname())
	}
	digest := sha256.Sum256([]byte(endpoint))
	return fmt.Sprintf("%s#%x", host, digest[:6])
}

type hostPolicy struct {
	exact    map[string]struct{}
	suffixes []string
}

func newHostPolicy(patterns []string) hostPolicy {
	if len(patterns) == 0 {
		patterns = defaultAllowedHostPatterns
	}
	policy := hostPolicy{exact: make(map[string]struct{}, len(patterns))}
	for _, pattern := range patterns {
		pattern = strings.ToLower(strings.TrimSuffix(strings.TrimSpace(pattern), "."))
		if strings.HasPrefix(pattern, "*.") {
			suffix := strings.TrimPrefix(pattern, "*.")
			if validHostname(suffix) {
				policy.suffixes = append(policy.suffixes, suffix)
			}
			continue
		}
		if validHostname(pattern) {
			policy.exact[pattern] = struct{}{}
		}
	}
	return policy
}

func (p hostPolicy) allows(host string) bool {
	host = strings.ToLower(strings.TrimSuffix(host, "."))
	if _, ok := p.exact[host]; ok {
		return true
	}
	for _, suffix := range p.suffixes {
		if strings.HasSuffix(host, "."+suffix) {
			return true
		}
	}
	return false
}

func (p hostPolicy) empty() bool {
	return len(p.exact) == 0 && len(p.suffixes) == 0
}

func validateEndpoint(endpoint string, allowedHosts hostPolicy) (*url.URL, error) {
	if endpoint == "" || len(endpoint) > maxEndpointLength || strings.TrimSpace(endpoint) != endpoint {
		return nil, errors.New("endpoint length or whitespace is invalid")
	}
	u, err := url.Parse(endpoint)
	if err != nil || !u.IsAbs() || u.Opaque != "" {
		return nil, errors.New("endpoint is not an absolute URL")
	}
	if !strings.EqualFold(u.Scheme, "https") {
		return nil, errors.New("endpoint must use https")
	}
	if u.User != nil || u.Fragment != "" {
		return nil, errors.New("endpoint must not contain user information or a fragment")
	}
	if port := u.Port(); port != "" && port != "443" {
		return nil, errors.New("endpoint must use port 443")
	}
	host := strings.ToLower(strings.TrimSuffix(u.Hostname(), "."))
	if net.ParseIP(host) != nil || !validHostname(host) {
		return nil, errors.New("endpoint must use a valid DNS hostname")
	}
	if !allowedHosts.allows(host) {
		return nil, fmt.Errorf("push service host %q is not allowed", host)
	}
	return u, nil
}

func validHostname(host string) bool {
	if host == "" || len(host) > 253 {
		return false
	}
	for _, label := range strings.Split(host, ".") {
		if label == "" || len(label) > 63 || label[0] == '-' || label[len(label)-1] == '-' {
			return false
		}
		for _, char := range label {
			if (char < 'a' || char > 'z') && (char < '0' || char > '9') && char != '-' {
				return false
			}
		}
	}
	return true
}

func decodeSubscriptionKey(value string) ([]byte, error) {
	if value == "" || strings.ContainsAny(value, " \t\r\n") {
		return nil, errors.New("subscription key is empty or contains whitespace")
	}
	trimmed := strings.TrimRight(value, "=")
	if strings.Contains(trimmed, "=") {
		return nil, errors.New("subscription key has invalid padding")
	}
	return base64.RawURLEncoding.DecodeString(trimmed)
}

type netIPResolver interface {
	LookupNetIP(ctx context.Context, network, host string) ([]netip.Addr, error)
}

type contextDialer interface {
	DialContext(ctx context.Context, network, address string) (net.Conn, error)
}

type restrictedDialer struct {
	allowedHosts hostPolicy
	resolver     netIPResolver
	dialer       contextDialer
}

func (d *restrictedDialer) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return nil, fmt.Errorf("invalid push address: %w", err)
	}
	host = strings.ToLower(strings.TrimSuffix(host, "."))
	if !d.allowedHosts.allows(host) {
		return nil, fmt.Errorf("push service host %q is not allowed", host)
	}
	if port != "443" {
		return nil, fmt.Errorf("push service port %q is not allowed", port)
	}

	addresses, err := d.resolver.LookupNetIP(ctx, "ip", host)
	if err != nil {
		return nil, fmt.Errorf("resolve push service: %w", err)
	}
	if len(addresses) == 0 {
		return nil, errors.New("push service resolved to no addresses")
	}
	for _, address := range addresses {
		if isBlockedIP(address) {
			return nil, fmt.Errorf("push service resolved to blocked address %s", address)
		}
	}

	var lastErr error
	for _, resolved := range addresses {
		connection, dialErr := d.dialer.DialContext(ctx, network, net.JoinHostPort(resolved.Unmap().String(), port))
		if dialErr == nil {
			return connection, nil
		}
		lastErr = dialErr
	}
	return nil, fmt.Errorf("connect to push service: %w", lastErr)
}

var blockedPrefixes = []netip.Prefix{
	netip.MustParsePrefix("0.0.0.0/8"),
	netip.MustParsePrefix("10.0.0.0/8"),
	netip.MustParsePrefix("100.64.0.0/10"),
	netip.MustParsePrefix("127.0.0.0/8"),
	netip.MustParsePrefix("169.254.0.0/16"),
	netip.MustParsePrefix("172.16.0.0/12"),
	netip.MustParsePrefix("192.0.0.0/24"),
	netip.MustParsePrefix("192.0.2.0/24"),
	netip.MustParsePrefix("192.168.0.0/16"),
	netip.MustParsePrefix("198.18.0.0/15"),
	netip.MustParsePrefix("198.51.100.0/24"),
	netip.MustParsePrefix("203.0.113.0/24"),
	netip.MustParsePrefix("224.0.0.0/4"),
	netip.MustParsePrefix("240.0.0.0/4"),
	netip.MustParsePrefix("::/128"),
	netip.MustParsePrefix("::1/128"),
	netip.MustParsePrefix("64:ff9b::/96"),
	netip.MustParsePrefix("64:ff9b:1::/48"),
	netip.MustParsePrefix("100::/64"),
	netip.MustParsePrefix("2001::/32"),
	netip.MustParsePrefix("2001:db8::/32"),
	netip.MustParsePrefix("2002::/16"),
	netip.MustParsePrefix("fc00::/7"),
	netip.MustParsePrefix("fec0::/10"),
	netip.MustParsePrefix("fe80::/10"),
	netip.MustParsePrefix("ff00::/8"),
}

func isBlockedIP(address netip.Addr) bool {
	if !address.IsValid() {
		return true
	}
	address = address.Unmap()
	if !address.IsGlobalUnicast() {
		return true
	}
	for _, prefix := range blockedPrefixes {
		if prefix.Contains(address) {
			return true
		}
	}
	return false
}
