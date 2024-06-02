package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"
)

type InferRequest struct {
	ModelName string                 `json:"model_name" binding:"required"`
	Inputs    map[string]interface{} `json:"inputs" binding:"required"`
}

type TaskRequest struct {
	IDs []string `json:"ids"`
}

type QueueRequest struct {
	ModelName string `uri:"model" binding:"required"`
}

func (server *Server) sendRequest(
	userID string,
	method string,
	url string,
	body io.Reader,
) (*http.Response, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, errors.New("failed to build request")
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("UID", userID)
	client := http.Client{Timeout: 30 * time.Second}
	return client.Do(req)
}

func (server *Server) callServingAgent(
	userID string,
	method string,
	path string,
	modelName string,
	data []byte,
	ctx *gin.Context,
) {
	agentURL := strings.Replace(server.config.ServingAgentAddress, "{MODEL-NAME}", modelName, 1)
	requestURL, err := url.JoinPath(agentURL, path)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	var requestBody io.Reader = nil
	if data != nil {
		requestBody = bytes.NewReader(data)
	}
	res, err := server.sendRequest(userID, method, requestURL, requestBody)
	if err != nil {
		log.Error().Msgf("failed to call %s, user-id: %s, model-name: %s, error: %v",
			requestURL, userID, modelName, err)
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Error().Msgf("failed to call %s, user-id: %s, model-name: %s, error: %v",
			requestURL, userID, modelName, err)
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	var outputs map[string]interface{}
	err = json.Unmarshal(body, &outputs)
	if err != nil {
		log.Error().Msgf("failed to call %s, user-id: %s, model-name: %s, error: %v",
			requestURL, userID, modelName, err)
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	if res.StatusCode >= 300 {
		log.Error().Msgf("failed to call %s, user-id: %s, model-name: %s, outputs: %v",
			requestURL, userID, modelName, outputs)
	}
	ctx.JSON(res.StatusCode, outputs)
}

func (server *Server) predict(ctx *gin.Context) {
	var req InferRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	data, err := json.Marshal(req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	userID := ctx.Request.Header.Get("UID")
	server.callServingAgent(userID, "POST", "v1/predict", req.ModelName, data, ctx)
}

func (server *Server) asyncPredict(ctx *gin.Context) {
	var req InferRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	data, err := json.Marshal(req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	userID := ctx.Request.Header.Get("UID")
	server.callServingAgent(userID, "POST", "async/v1/predict", req.ModelName, data, ctx)
}

func (server *Server) getTask(ctx *gin.Context) {
	taskID := path.Base(ctx.Request.RequestURI)
	outputs, err := server.webhook.GetTaskInfo(taskID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	ctx.JSON(http.StatusOK, outputs)
}

func (server *Server) getTasks(ctx *gin.Context) {
	var req TaskRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	if len(req.IDs) > 20 {
		ctx.JSON(http.StatusBadRequest, errorResponse(
			errors.New("the number of ids cannot be > 20")))
		return
	}
	outputs := make([]interface{}, 0)
	for _, taskID := range req.IDs {
		result, err := server.webhook.GetTaskInfo(taskID)
		if err != nil {
			continue
		}
		outputs = append(outputs, result)
	}
	ctx.JSON(http.StatusOK, outputs)
}

func (server *Server) getTaskQueueSize(ctx *gin.Context) {
	var req QueueRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	userID := ctx.Request.Header.Get("UID")
	server.callServingAgent(userID, "GET", "v1/queue_size", req.ModelName, nil, ctx)
}

func (server *Server) pauseTaskQueue(ctx *gin.Context) {
	var req QueueRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	server.callServingAgent("", "POST", "pause", req.ModelName, nil, ctx)
}

func (server *Server) unpauseTaskQueue(ctx *gin.Context) {
	var req QueueRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	server.callServingAgent("", "POST", "unpause", req.ModelName, nil, ctx)
}
