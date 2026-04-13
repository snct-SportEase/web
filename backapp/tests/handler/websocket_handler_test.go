package handler_test

import (
	"backapp/internal/handler"
	"backapp/internal/websocket"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	gowebsocket "github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// dialWS is a helper that connects to the given WebSocket URL and fails the test on error.
func dialWS(t *testing.T, wsURL string) *gowebsocket.Conn {
	t.Helper()
	conn, _, err := gowebsocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err, "failed to connect to WebSocket")
	return conn
}

// readWithDeadline reads one message within the given duration, returning (message, true) on
// success or ("", false) if the deadline expires.
func readWithDeadline(conn *gowebsocket.Conn, d time.Duration) (string, bool) {
	conn.SetReadDeadline(time.Now().Add(d))
	_, msg, err := conn.ReadMessage()
	if err != nil {
		return "", false
	}
	return string(msg), true
}

func newWSTestServer(t *testing.T) (*httptest.Server, *websocket.HubManager) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	hubManager := websocket.NewHubManager()
	wsHandler := handler.NewWebSocketHandler(hubManager)

	router := gin.New()
	router.GET("/ws/tournaments/:tournament_id", wsHandler.ServeTournamentWebSocket)
	router.GET("/ws/progress", wsHandler.ServeProgressWebSocket)

	server := httptest.NewServer(router)
	t.Cleanup(server.Close)
	return server, hubManager
}

func TestWebSocketHandler_ServeTournamentWebSocket(t *testing.T) {
	t.Run("client receives broadcast", func(t *testing.T) {
		server, hubManager := newWSTestServer(t)
		wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws/tournaments/1"

		conn := dialWS(t, wsURL)
		defer conn.Close()

		time.Sleep(50 * time.Millisecond) // wait for registration
		hubManager.BroadcastTo("tournament:1", gin.H{"type": "test", "data": "hello"})

		msg, ok := readWithDeadline(conn, time.Second)
		require.True(t, ok, "expected to receive a message")
		assert.JSONEq(t, `{"data":"hello","type":"test"}`, msg)
	})

	t.Run("multiple clients on the same tournament all receive the broadcast", func(t *testing.T) {
		server, hubManager := newWSTestServer(t)
		wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws/tournaments/2"

		conn1 := dialWS(t, wsURL)
		defer conn1.Close()
		conn2 := dialWS(t, wsURL)
		defer conn2.Close()

		time.Sleep(50 * time.Millisecond)
		hubManager.BroadcastTo("tournament:2", gin.H{"type": "update", "score": 3})

		msg1, ok1 := readWithDeadline(conn1, time.Second)
		msg2, ok2 := readWithDeadline(conn2, time.Second)

		require.True(t, ok1, "client1 expected to receive a message")
		require.True(t, ok2, "client2 expected to receive a message")
		assert.JSONEq(t, `{"type":"update","score":3}`, msg1)
		assert.JSONEq(t, `{"type":"update","score":3}`, msg2)
	})

	t.Run("topic isolation - client on tournament:3 does not receive tournament:4 broadcast", func(t *testing.T) {
		server, hubManager := newWSTestServer(t)
		baseURL := "ws" + strings.TrimPrefix(server.URL, "http")

		conn3 := dialWS(t, baseURL+"/ws/tournaments/3")
		defer conn3.Close()
		conn4 := dialWS(t, baseURL+"/ws/tournaments/4")
		defer conn4.Close()

		time.Sleep(50 * time.Millisecond)
		hubManager.BroadcastTo("tournament:4", gin.H{"type": "goal"})

		// conn3 must NOT receive this message
		_, received := readWithDeadline(conn3, 200*time.Millisecond)
		assert.False(t, received, "client on tournament:3 should not receive tournament:4 message")

		// conn4 MUST receive the message
		msg, ok := readWithDeadline(conn4, time.Second)
		require.True(t, ok, "client on tournament:4 should receive the broadcast")
		assert.JSONEq(t, `{"type":"goal"}`, msg)
	})

	t.Run("broadcast to topic with no subscribers does not panic", func(t *testing.T) {
		_, hubManager := newWSTestServer(t)

		assert.NotPanics(t, func() {
			hubManager.BroadcastTo("tournament:99", gin.H{"type": "test"})
		})
	})

	t.Run("client disconnects and subsequent broadcast does not panic", func(t *testing.T) {
		server, hubManager := newWSTestServer(t)
		wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws/tournaments/5"

		conn := dialWS(t, wsURL)
		time.Sleep(50 * time.Millisecond)
		conn.Close() // disconnect
		time.Sleep(50 * time.Millisecond) // wait for unregister

		assert.NotPanics(t, func() {
			hubManager.BroadcastTo("tournament:5", gin.H{"type": "after_disconnect"})
		})
	})
}

func TestWebSocketHandler_ServeProgressWebSocket(t *testing.T) {
	t.Run("client receives broadcast on progress topic", func(t *testing.T) {
		server, hubManager := newWSTestServer(t)
		wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws/progress"

		conn := dialWS(t, wsURL)
		defer conn.Close()

		time.Sleep(50 * time.Millisecond)
		hubManager.BroadcastTo("progress", gin.H{"type": "progress_update", "class_id": 10})

		msg, ok := readWithDeadline(conn, time.Second)
		require.True(t, ok, "expected to receive a message on progress topic")
		assert.JSONEq(t, `{"type":"progress_update","class_id":10}`, msg)
	})

	t.Run("progress and tournament topics are isolated", func(t *testing.T) {
		server, hubManager := newWSTestServer(t)
		baseURL := "ws" + strings.TrimPrefix(server.URL, "http")

		progressConn := dialWS(t, baseURL+"/ws/progress")
		defer progressConn.Close()
		tournamentConn := dialWS(t, baseURL+"/ws/tournaments/6")
		defer tournamentConn.Close()

		time.Sleep(50 * time.Millisecond)
		hubManager.BroadcastTo("progress", gin.H{"type": "progress_update"})

		// tournament client must NOT receive progress message
		_, received := readWithDeadline(tournamentConn, 200*time.Millisecond)
		assert.False(t, received, "tournament client should not receive progress message")

		// progress client MUST receive it
		msg, ok := readWithDeadline(progressConn, time.Second)
		require.True(t, ok, "progress client should receive the broadcast")
		assert.JSONEq(t, `{"type":"progress_update"}`, msg)
	})
}
