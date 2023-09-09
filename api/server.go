package api

import (
	"fmt"
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
	webhook     platform.Webhook
}

func NewServer(
	config utils.Config,
	platform platform.Platform,
	distributor worker.TaskDistributor,
	webhook platform.Webhook,
) (*Server, error) {
	server := Server{
		config:      config,
		router:      nil,
		platform:    platform,
		distributor: distributor,
		webhook:     webhook,
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

	taskRoutes := router.Group("/task")
	taskRoutes.GET("/:id", server.getTask)

	server.router = router
}

// Start runs the HTTP server on a specific address.
func (server *Server) Start(address string) error {
	return server.router.Run(address)
}

func (server *Server) checkHealth(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"message": "API OK"})
}

func (server *Server) getTask(ctx *gin.Context) {
	ctx.Redirect(http.StatusFound,
		fmt.Sprintf("http://%s%s", server.config.WebhookServerAddress, ctx.Request.RequestURI))
	/*
		url := fmt.Sprintf("http://%s%s", server.config.WebhookServerAddress, ctx.Request.RequestURI)
		res, err := http.Get(url)
		if err != nil || res.StatusCode != 200 {
			ctx.JSON(http.StatusInternalServerError, errorResponse(err))
			return
		}
		fmt.Println(res)
		ctx.JSON(http.StatusOK, nil)
	*/
}

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
