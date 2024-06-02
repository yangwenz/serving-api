package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/HyperGAI/serving-api/utils"
	"io"
	"net/http"
	"time"
)

type Webhook interface {
	GetTaskInfo(taskID string) (interface{}, error)
}

type InternalWebhook struct {
	config utils.Config
	url    string
}

func NewInternalWebhook(config utils.Config) Webhook {
	webhook := InternalWebhook{
		config: config,
		url:    fmt.Sprintf("http://%s/task", config.WebhookServerAddress),
	}
	return &webhook
}

func (webhook *InternalWebhook) GetTaskInfo(taskID string) (interface{}, error) {
	url := fmt.Sprintf("%s/%s", webhook.url, taskID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, errors.New("failed to build request")
	}
	req.Header.Set("apikey", webhook.config.WebhookAPIKey)

	client := http.Client{Timeout: 10 * time.Second}
	res, err := client.Do(req)
	if err != nil {
		return nil, errors.New("failed to get task info")
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return nil, errors.New("failed to get task info")
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}
	var outputs interface{}
	err = json.Unmarshal(body, &outputs)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal data: %w", err)
	}
	return outputs, nil
}
