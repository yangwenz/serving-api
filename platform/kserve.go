package platform

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/yangwenz/model-serving/utils"
	"io"
	"net/http"
	"time"
)

type KServe struct {
	address      string
	customDomain string
	timeout      int
}

func NewKServe(config utils.Config) Platform {
	return &KServe{
		address:      config.KServeAddress,
		customDomain: config.KServeCustomDomain,
		timeout:      config.KServeRequestTimeout,
	}
}

func (service *KServe) Predict(request InferRequest, version string) (*InferResponse, error) {
	modelName := request.GetModelName()
	inputs := request.GetInputs()

	if version == "v1" {
		// Marshal the input data
		data, err := json.Marshal(inputs)
		if err != nil {
			return nil, fmt.Errorf("failed to marshall request: %s", err)
		}

		// Build a new prediction request
		url := fmt.Sprintf("http://%s/v1/models/%s:predict", service.address, modelName)
		req, err := http.NewRequest("POST", url, bytes.NewReader(data))
		if err != nil {
			return nil, fmt.Errorf("failed to build request: %s", err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Host = fmt.Sprintf("%s.%s.%s",
			modelName, request.GetNamespace(), service.customDomain)

		// Send the prediction request
		client := http.Client{Timeout: time.Duration(service.timeout) * time.Second}
		res, err := client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("falied to send request: %s", err)
		}
		if res.StatusCode != 200 {
			return nil, fmt.Errorf("kserve service error")
		}

		// Parse the response
		defer res.Body.Close()
		body, err := io.ReadAll(res.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read response body: %s", err)
		}
		var outputs map[string]interface{}
		err = json.Unmarshal(body, &outputs)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal response body: %s", err)
		}
		response := InferResponse{Outputs: outputs}
		return &response, nil
	}
	return nil, fmt.Errorf("unknown kserve predict version: %s", version)
}
