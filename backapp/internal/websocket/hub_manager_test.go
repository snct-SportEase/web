package websocket

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHubManager_GetHub(t *testing.T) {
	t.Run("creates a new hub for a new topic", func(t *testing.T) {
		m := NewHubManager()
		hub := m.GetHub("tournament:1")
		require.NotNil(t, hub)
	})

	t.Run("returns the same hub for the same topic", func(t *testing.T) {
		m := NewHubManager()
		hub1 := m.GetHub("tournament:1")
		hub2 := m.GetHub("tournament:1")
		assert.Same(t, hub1, hub2, "GetHub must return the same hub for an existing topic")
	})

	t.Run("returns different hubs for different topics", func(t *testing.T) {
		m := NewHubManager()
		hub1 := m.GetHub("tournament:1")
		hub2 := m.GetHub("tournament:2")
		assert.NotSame(t, hub1, hub2, "GetHub must return distinct hubs for different topics")
	})
}

func TestHubManager_BroadcastTo(t *testing.T) {
	t.Run("no-op when topic has no hub - does not panic", func(t *testing.T) {
		m := NewHubManager()
		assert.NotPanics(t, func() {
			m.BroadcastTo("nonexistent:topic", map[string]string{"key": "value"})
		})
	})

	t.Run("delivers message to registered client", func(t *testing.T) {
		m := NewHubManager()
		hub := m.GetHub("tournament:10")

		// Register a client directly via the hub's channel (package-internal access).
		send := make(chan []byte, 1)
		client := &Client{hub: hub, send: send}
		hub.register <- client

		// Allow the hub goroutine to process the registration.
		time.Sleep(20 * time.Millisecond)

		m.BroadcastTo("tournament:10", map[string]string{"type": "test"})

		select {
		case msg := <-send:
			assert.Contains(t, string(msg), "test")
		case <-time.After(200 * time.Millisecond):
			t.Fatal("expected to receive a message within 200 ms")
		}
	})
}
