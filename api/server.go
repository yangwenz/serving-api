package api

import (
	"github.com/gin-gonic/gin"
	"github.com/yangwenz/model-serving/platform"
	"github.com/yangwenz/model-serving/utils"
	"github.com/yangwenz/model-serving/worker"
	"net/http"
)

type Server struct {
	config      utils.Config
	router      *gin.Engine
	platform    platform.Platform
	distributor worker.TaskDistributor
}

func NewServer(config utils.Config, platform platform.Platform, distributor worker.TaskDistributor) (*Server, error) {
	server := Server{
		config:      config,
		router:      nil,
		platform:    platform,
		distributor: distributor,
	}
	server.setupRouter()
	return &server, nil
}

func (server *Server) setupRouter() {
	router := gin.Default()

	router.GET("/live", server.checkHealth)
	router.GET("/ready", server.checkHealth)

	v1Routes := router.Group("/v1")
	v1Routes.POST("/predict", server.predictV1)

	asyncV1Routes := router.Group("/async/v1")
	asyncV1Routes.POST("/predict", server.asyncPredictV1)

	server.router = router
}

// Start runs the HTTP server on a specific address.
func (server *Server) Start(address string) error {
	return server.router.Run(address)
}

func (server *Server) checkHealth(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"message": "API OK"})
}

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
