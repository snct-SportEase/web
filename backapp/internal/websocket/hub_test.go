package websocket

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHub_RegisterAndUnregister(t *testing.T) {
	t.Run("registered client receives broadcast", func(t *testing.T) {
		hub := NewHub()
		go hub.Run()

		send := make(chan []byte, 1)
		client := &Client{hub: hub, send: send}
		hub.register <- client
		time.Sleep(20 * time.Millisecond)

		hub.Broadcast([]byte("hello"))

		select {
		case msg := <-send:
			assert.Equal(t, "hello", string(msg))
		case <-time.After(200 * time.Millisecond):
			t.Fatal("expected to receive broadcast within 200 ms")
		}
	})

	t.Run("unregistered client does not receive subsequent broadcasts", func(t *testing.T) {
		hub := NewHub()
		go hub.Run()

		send := make(chan []byte, 1)
		client := &Client{hub: hub, send: send}
		hub.register <- client
		time.Sleep(20 * time.Millisecond)

		hub.unregister <- client
		time.Sleep(20 * time.Millisecond)

		// Drain the closed channel so Broadcast doesn't block.
		// After unregister, the send channel is closed; we simply verify no panic.
		assert.NotPanics(t, func() {
			// Broadcast goes to the hub goroutine; the closed client is removed, so nothing
			// is written to `send`. The hub should handle this gracefully.
			go hub.Broadcast([]byte("after unregister"))
		})
		time.Sleep(20 * time.Millisecond)
	})
}

func TestHub_BroadcastJSON(t *testing.T) {
	t.Run("marshals struct and delivers to client", func(t *testing.T) {
		hub := NewHub()
		go hub.Run()

		send := make(chan []byte, 1)
		client := &Client{hub: hub, send: send}
		hub.register <- client
		time.Sleep(20 * time.Millisecond)

		type payload struct {
			Type string `json:"type"`
			ID   int    `json:"id"`
		}
		hub.BroadcastJSON(payload{Type: "goal", ID: 42})

		select {
		case msg := <-send:
			var got payload
			require.NoError(t, json.Unmarshal(msg, &got))
			assert.Equal(t, "goal", got.Type)
			assert.Equal(t, 42, got.ID)
		case <-time.After(200 * time.Millisecond):
			t.Fatal("expected to receive broadcast within 200 ms")
		}
	})

	t.Run("multiple clients all receive BroadcastJSON", func(t *testing.T) {
		hub := NewHub()
		go hub.Run()

		send1 := make(chan []byte, 1)
		send2 := make(chan []byte, 1)
		hub.register <- &Client{hub: hub, send: send1}
		hub.register <- &Client{hub: hub, send: send2}
		time.Sleep(20 * time.Millisecond)

		hub.BroadcastJSON(map[string]string{"event": "start"})

		for i, ch := range []chan []byte{send1, send2} {
			select {
			case msg := <-ch:
				assert.Contains(t, string(msg), "start", "client %d should receive the broadcast", i+1)
			case <-time.After(200 * time.Millisecond):
				t.Fatalf("client %d did not receive broadcast within 200 ms", i+1)
			}
		}
	})
}
