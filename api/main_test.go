package api

import (
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/yangwenz/model-serving/platform"
	"github.com/yangwenz/model-serving/utils"
	"github.com/yangwenz/model-serving/worker"
	"os"
	"testing"
)

func newTestServer(
	t *testing.T,
	platform platform.Platform,
	distributor worker.TaskDistributor,
	webhook platform.Webhook,
) *Server {
	config := utils.Config{}
	server, err := NewServer(config, platform, distributor, webhook)
	require.NoError(t, err)
	return server
}

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	os.Exit(m.Run())
}
