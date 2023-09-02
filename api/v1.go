package api

import (
	"github.com/gin-gonic/gin"
	"github.com/yangwenz/model-serving/platform"
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
