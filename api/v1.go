package api

import (
	"github.com/gin-gonic/gin"
	"github.com/yangwenz/model-serving/platform"
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

func (server *Server) predictV1(ctx *gin.Context) {
	var req InferRequest
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
