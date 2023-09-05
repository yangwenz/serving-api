package api

import (
	"github.com/gin-gonic/gin"
	"github.com/hibiken/asynq"
	"github.com/yangwenz/model-serving/platform"
	"github.com/yangwenz/model-serving/worker"
	"net/http"
)

func (server *Server) predictV1(ctx *gin.Context) {
	var req platform.InferRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	response, err := server.platform.Predict(&req, "v1")
	if err != nil {
		switch err.StatusCode {
		case platform.MarshalError:
			ctx.JSON(http.StatusBadRequest, errorResponse(err))
		case platform.BuildRequestError:
			ctx.JSON(http.StatusBadRequest, errorResponse(err))
		case platform.SendRequestError:
			ctx.JSON(http.StatusForbidden, errorResponse(err))
		case platform.InvalidInputError:
			ctx.JSON(http.StatusBadRequest, errorResponse(err))
		default:
			ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		}
		return
	}
	ctx.JSON(http.StatusOK, response)
}

func (server *Server) asyncPredictV1(ctx *gin.Context) {
	var req platform.InferRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	opts := []asynq.Option{
		asynq.MaxRetry(5),
		asynq.Queue(worker.QueueCritical),
	}
	payload := &worker.PayloadRunPrediction{
		InferRequest: req,
		APIVersion:   "v1",
	}
	// Submit a new prediction task
	err := server.distributor.DistributeTaskRunPrediction(ctx, payload, opts...)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	// Add a prediction task record
	res, err := server.webhook.CreateNewTask(req.ModelName, req.ModelVersion)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
	}
	ctx.JSON(http.StatusOK, gin.H{"message": res})
}
