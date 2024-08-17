package api

import (
	"github.com/HyperGAI/serving-api/utils"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
)

type Server struct {
	config  utils.Config
	webhook Webhook
	router  *gin.Engine
}

func NewServer(
	config utils.Config,
	webhook Webhook,
) (*Server, error) {
	server := Server{
		config:  config,
		webhook: webhook,
	}
	server.setupRouter()
	return &server, nil
}

func (server *Server) setupRouter() {
	router := gin.Default()

	router.GET("/live", server.checkHealth)
	router.GET("/ready", server.checkHealth)
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))
	router.POST("/pause/:model", server.pauseTaskQueue)
	router.POST("/unpause/:model", server.unpauseTaskQueue)

	syncRoutes := router.Group("/v1")
	syncRoutes.Use(traceRequest())
	syncRoutes.Use(authenticateRequest())
	syncRoutes.Use(rateLimitByUser(server.config, server.config.FormattedRateSync, "sync_predict"))
	syncRoutes.Use(prometheusMiddleware())
	syncRoutes.POST("/predict", server.predict)
	syncRoutes.POST("/generate", server.generate)

	asyncRoutes := router.Group("/async/v1")
	asyncRoutes.Use(traceRequest())
	asyncRoutes.Use(authenticateRequest())
	asyncRoutes.Use(rateLimitByUser(server.config, server.config.FormattedRateAsync, "async_predict"))
	asyncRoutes.Use(prometheusMiddleware())
	asyncRoutes.POST("/predict", server.asyncPredict)

	taskRoutes := router.Group("/task")
	taskRoutes.Use(traceRequest())
	taskRoutes.Use(authenticateRequest())
	taskRoutes.Use(rateLimitByUser(server.config, server.config.FormattedRateTask, "task"))
	taskRoutes.Use(prometheusMiddleware())
	taskRoutes.GET("/:id", server.getTask)
	taskRoutes.POST("/batch", server.getTasks)

	queueRoutes := router.Group("/queue_size")
	queueRoutes.Use(traceRequest())
	queueRoutes.Use(authenticateRequest())
	queueRoutes.Use(rateLimitByUser(server.config, server.config.FormattedRateTask, "queue"))
	queueRoutes.Use(prometheusMiddleware())
	queueRoutes.GET("/:model", server.getTaskQueueSize)

	server.router = router
}

func (server *Server) Start(address string) error {
	return server.router.Run(address)
}

func (server *Server) Handler() http.Handler {
	return server.router.Handler()
}

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}

func (server *Server) checkHealth(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"message": "API OK"})
}
