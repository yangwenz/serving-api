package api

import (
	"fmt"
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

type InferResponse struct {
	Outputs map[string]interface{} `json:"outputs"`
}

func (request *InferResponse) GetOutputs() map[string]interface{} {
	return request.Outputs
}

func (server *Server) predictV1(ctx *gin.Context) {
	var req InferRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	fmt.Println(req)
}
