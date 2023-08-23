package api

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type InferRequest struct {
	ModelName    string                 `json:"model_name" binding:"required"`
	ModelVersion string                 `json:"model_version"`
	Inputs       map[string]interface{} `json:"inputs" binding:"required"`
}

func (request *InferRequest) GetModelName() string {
	return request.ModelName
}

func (request *InferRequest) GetModelVersion() string {
	return request.ModelVersion
}

func (request *InferRequest) GetInputs() map[string]interface{} {
	return request.Inputs
}

func (request *InferRequest) GetNamespace() string {
	return "default"
}

func (server *Server) predictV1(ctx *gin.Context) {
	var req InferRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	response, err := server.platform.Predict(&req, "v1")
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	ctx.JSON(http.StatusOK, response)
}
