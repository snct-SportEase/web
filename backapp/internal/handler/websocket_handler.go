package handler

import (
	"backapp/internal/websocket"

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
	hub := h.hubManager.GetHub("tournament:" + tournamentID)
	websocket.ServeWs(hub, c.Writer, c.Request, h.allowedOrigin)
}

func (h *WebSocketHandler) ServeProgressWebSocket(c *gin.Context) {
	hub := h.hubManager.GetHub("progress")
	websocket.ServeWs(hub, c.Writer, c.Request, h.allowedOrigin)
}
