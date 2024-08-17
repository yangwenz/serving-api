package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
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

type StreamingMessage struct {
	Id   int    `json:"id"`
	Data string `json:"data"`
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
	client := http.Client{Timeout: 60 * time.Second}
	return client.Do(req)
}

func (server *Server) sendStreamingRequest(
	ctx context.Context,
	userID string,
	method string,
	url string,
	body io.Reader,
	encoder *json.Encoder,
	flusher http.Flusher,
) error {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("UID", userID)

	client := http.Client{Timeout: 60 * time.Second}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("status-code: %d", res.StatusCode)
	}
	decoder := json.NewDecoder(res.Body)

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("client stopped listening")
		default:
			var m StreamingMessage
			if err := decoder.Decode(&m); err != nil {
				if err == io.EOF {
					return nil
				}
				return fmt.Errorf("failed to decode request: %v", err)
			}
			if err := encoder.Encode(m); err != nil {
				return fmt.Errorf("failed to encode request: %v", err)
			}
			flusher.Flush()
		}
	}
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

func (server *Server) callServingAgentStreaming(
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

	w, r := ctx.Writer, ctx.Request
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Transfer-Encoding", "chunked")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	encoder := json.NewEncoder(w)

	err = server.sendStreamingRequest(r.Context(), userID, method, requestURL, requestBody, encoder, flusher)
	if err != nil {
		log.Error().Msgf("failed to call %s, user-id: %s, model-name: %s, error: %v",
			requestURL, userID, modelName, err)
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
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

func (server *Server) generate(ctx *gin.Context) {
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
	server.callServingAgentStreaming(userID, "POST", "v1/generate", req.ModelName, data, ctx)
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
