package api

import (
	"github.com/gin-gonic/gin"
	"github.com/yangwenz/model-serving/utils"
	"net/http"
)

type Server struct {
	config utils.Config
	router *gin.Engine
}

func NewServer(config utils.Config) (*Server, error) {
	server := Server{
		config: config,
		router: nil,
	}
	server.setupRouter()
	return &server, nil
}

func (server *Server) setupRouter() {
	router := gin.Default()

	router.GET("/live", server.checkHealth)
	router.GET("/ready", server.checkHealth)

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
