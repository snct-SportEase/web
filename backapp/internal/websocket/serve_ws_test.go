package websocket

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/require"
)

func startServeWsServer(t *testing.T, hubPtr *Hub) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ServeWs(func() *Hub { return hubPtr }, w, r, "")
	}))
	t.Cleanup(srv.Close)
	return srv
}

func waitPong(t *testing.T, conn *websocket.Conn, wait time.Duration) error {
	t.Helper()
	got := make(chan struct{}, 1)
	conn.SetPongHandler(func(string) error {
		select {
		case got <- struct{}{}:
		default:
		}
		return nil
	})
	go func() {
		_ = conn.SetReadDeadline(time.Now().Add(wait))
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				return
			}
		}
	}()
	if err := conn.WriteControl(websocket.PingMessage, nil, time.Now().Add(time.Second)); err != nil {
		return err
	}
	select {
	case <-got:
		return nil
	case <-time.After(wait):
		return errDeadline
	}
}

type deadlineErr struct{}

func (deadlineErr) Error() string { return "i/o deadline" }

var errDeadline = deadlineErr{}

func TestServeWsHealthyHubResponds(t *testing.T) {
	hub := NewHub()
	go hub.Run()
	srv := startServeWsServer(t, hub)

	conn, _, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(srv.URL, "http"), nil)
	require.NoError(t, err)
	defer conn.Close()

	require.NoError(t, waitPong(t, conn, 5*time.Second))
}

func TestServeWsDoesNotBlockAfterHubRunExited(t *testing.T) {
	deadHub := NewHub(func() {})
	go deadHub.Run()
	seed := &Client{hub: deadHub, send: make(chan []byte, 1)}
	deadHub.register <- seed
	deadHub.unregister <- seed
	select {
	case <-deadHub.stopped:
	case <-time.After(2 * time.Second):
		t.Fatal("seed hub did not exit")
	}

	srv := startServeWsServer(t, deadHub)
	done := make(chan error, 1)
	go func() {
		conn, _, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(srv.URL, "http"), nil)
		if err != nil {
			// Connection may fail or close immediately after upgrade — both are OK.
			done <- nil
			return
		}
		defer conn.Close()
		// If upgrade succeeded, the server must close promptly instead of hanging.
		_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
		_, _, readErr := conn.ReadMessage()
		done <- readErr
	}()

	select {
	case <-done:
		// ServeWs returned without blocking forever.
	case <-time.After(3 * time.Second):
		t.Fatal("ServeWs blocked on register after hub teardown")
	}
}
