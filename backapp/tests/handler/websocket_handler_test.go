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
)

func TestWebSocketHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	hubManager := websocket.NewHubManager()
	wsHandler := handler.NewWebSocketHandler(hubManager)

	router := gin.New()
	router.GET("/ws/tournaments/:tournament_id", wsHandler.ServeTournamentWebSocket)

	server := httptest.NewServer(router)
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws/tournaments/1"

	// Create a WebSocket client
	conn, _, err := gowebsocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Failed to connect to WebSocket: %v", err)
	}
	defer conn.Close()

	// Broadcast a message
	go func() {
		time.Sleep(100 * time.Millisecond) // wait for client to connect
		hubManager.BroadcastTo("tournament:1", gin.H{"type": "test", "data": "hello"})
	}()

	// Read message from the client
	conn.SetReadDeadline(time.Now().Add(1 * time.Second))
	_, msg, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("Failed to read message: %v", err)
	}

	expectedMsg := `{"data":"hello","type":"test"}`
	assert.JSONEq(t, expectedMsg, string(msg))
}
