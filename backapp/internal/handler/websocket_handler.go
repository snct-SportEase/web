package handler

import (
	"backapp/internal/websocket"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type WebSocketHandler struct {
	hubManager    *websocket.HubManager
	allowedOrigin string
}

func NewWebSocketHandler(hubManager *websocket.HubManager, allowedOrigin ...string) *WebSocketHandler {
	origin := ""
	if len(allowedOrigin) > 0 {
		origin = allowedOrigin[0]
	}
	return &WebSocketHandler{hubManager: hubManager, allowedOrigin: origin}
}

func (h *WebSocketHandler) ServeTournamentWebSocket(c *gin.Context) {
	tournamentID := c.Param("tournament_id")
	parsedID, err := strconv.Atoi(tournamentID)
	if err != nil || parsedID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tournament ID"})
		return
	}
	topic := "tournament:" + strconv.Itoa(parsedID)
	websocket.ServeWs(func() *websocket.Hub { return h.hubManager.GetHub(topic) }, c.Writer, c.Request, h.allowedOrigin)
}

func (h *WebSocketHandler) ServeProgressWebSocket(c *gin.Context) {
	websocket.ServeWs(func() *websocket.Hub { return h.hubManager.GetHub("progress") }, c.Writer, c.Request, h.allowedOrigin)
}
