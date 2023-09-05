package platform

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/yangwenz/model-serving/utils"
	"io"
	"net/http"
	"time"
)

type InternalWebhook struct {
	config utils.Config
	url    string
}

type TaskInfo struct {
	ID          string      `json:"id"`
	Status      string      `json:"status"`
	RunningTime string      `json:"running_time"`
	Outputs     interface{} `json:"outputs"`
}

func NewInternalWebhook(config utils.Config) Webhook {
	webhook := InternalWebhook{
		config: config,
		url:    fmt.Sprintf("http://%s/task", config.WebhookServerAddress),
	}
	return &webhook
}

func (webhook *InternalWebhook) CreateNewTask(taskID string, modelName string, modelVersion string) (string, error) {
	info := map[string]string{
		"id":            taskID,
		"model_name":    modelName,
		"model_version": modelVersion,
	}
	data, err := json.Marshal(info)
	if err != nil {
		return "", errors.New("failed to marshal task info")
	}
	req, err := http.NewRequest("POST", webhook.url, bytes.NewReader(data))
	if err != nil {
		return "", errors.New("failed to build request")
	}
	req.Header.Set("Content-Type", "application/json")

	client := http.Client{Timeout: 10 * time.Second}
	res, err := client.Do(req)
	if err != nil || res.StatusCode != 200 {
		return "", fmt.Errorf("http post request /task failed: %w", err)
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}
	return string(body), nil
}

func (webhook *InternalWebhook) UpdateTaskInfo(info TaskInfo) error {
	data, err := json.Marshal(info)
	if err != nil {
		return errors.New("failed to marshal task info")
	}
	req, err := http.NewRequest("PUT", webhook.url, bytes.NewReader(data))
	if err != nil {
		return errors.New("failed to build request")
	}
	req.Header.Set("Content-Type", "application/json")

	client := http.Client{Timeout: 10 * time.Second}
	res, err := client.Do(req)
	if err != nil || res.StatusCode != 200 {
		return errors.New("failed to update task info")
	}
	return nil
}
