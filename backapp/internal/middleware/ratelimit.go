package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type ipEntry struct {
	mu      sync.Mutex
	count   int
	resetAt time.Time
}

type rateLimitStore struct {
	mu      sync.RWMutex
	entries map[string]*ipEntry
}

func newRateLimitStore(cleanupInterval time.Duration) *rateLimitStore {
	s := &rateLimitStore{
		entries: make(map[string]*ipEntry),
	}
	go s.cleanup(cleanupInterval)
	return s
}

func (s *rateLimitStore) cleanup(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for range ticker.C {
		now := time.Now()
		s.mu.Lock()
		for ip, e := range s.entries {
			e.mu.Lock()
			if now.After(e.resetAt) {
				delete(s.entries, ip)
			}
			e.mu.Unlock()
		}
		s.mu.Unlock()
	}
}

func (s *rateLimitStore) allow(ip string, limit int, window time.Duration) bool {
	s.mu.RLock()
	e, exists := s.entries[ip]
	s.mu.RUnlock()

	if !exists {
		s.mu.Lock()
		// double-check after acquiring write lock
		e, exists = s.entries[ip]
		if !exists {
			e = &ipEntry{}
			s.entries[ip] = e
		}
		s.mu.Unlock()
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	now := time.Now()
	if now.After(e.resetAt) {
		e.count = 0
		e.resetAt = now.Add(window)
	}
	e.count++
	return e.count <= limit
}

// RateLimit returns a middleware that limits requests to `limit` per `window` per client IP.
func RateLimit(limit int, window time.Duration) gin.HandlerFunc {
	store := newRateLimitStore(window * 2)
	return func(c *gin.Context) {
		if !store.allow(c.ClientIP(), limit, window) {
			c.Header("Retry-After", "60")
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "Too many requests"})
			c.Abort()
			return
		}
		c.Next()
	}
}
