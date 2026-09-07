package router

import (
	"backapp/internal/config"
	"backapp/internal/websocket"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestSetupRouterDoesNotPanic(t *testing.T) {
	gin.SetMode(gin.TestMode)

	require.NotPanics(t, func() {
		SetupRouter(nil, &config.Config{
			TrustedProxyCIDRs: []string{"127.0.0.1/32"},
		}, websocket.NewHubManager())
	})
}
