package api

import (
	"github.com/HyperGAI/serving-api/utils"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func newTestServer(
	t *testing.T,
	webhook Webhook,
) *Server {
	config := utils.Config{}
	server, err := NewServer(config, webhook)
	require.NoError(t, err)
	return server
}

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	os.Exit(m.Run())
}
