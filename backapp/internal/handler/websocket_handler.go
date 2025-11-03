package handler

import (
	"backapp/internal/websocket"

	"github.com/gin-gonic/gin"
)

type WebSocketHandler struct {
	hubManager *websocket.HubManager
}

func NewWebSocketHandler(hubManager *websocket.HubManager) *WebSocketHandler {
	return &WebSocketHandler{hubManager: hubManager}
}

func (h *WebSocketHandler) ServeTournamentWebSocket(c *gin.Context) {
	tournamentID := c.Param("tournament_id")
	hub := h.hubManager.GetHub("tournament:" + tournamentID)
	websocket.ServeWs(hub, c.Writer, c.Request)
}
