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

func (service *KServe) Predict(request InferRequest, version string) (*InferResponse, *RequestError) {
	if version == "v1" {
		return service.predictV1(request, version)
	}
	return nil, NewRequestError(UnknownAPIVersion,
		errors.New("prediction API version is not supported"))
}

func (service *KServe) predictV1(request InferRequest, version string) (*InferResponse, *RequestError) {
	modelName := request.GetModelName()
	inputs := request.GetInputs()

	// Marshal the input data
	data, err := json.Marshal(inputs)
	if err != nil {
		return nil, NewRequestError(MarshalError,
			errors.New("failed to marshal request"))
	}

	// Build a new prediction request
	url := fmt.Sprintf("http://%s/v1/models/%s:predict", service.address, modelName)
	req, err := http.NewRequest("POST", url, bytes.NewReader(data))
	if err != nil {
		return nil, NewRequestError(BuildRequestError,
			errors.New("failed to build request"))
	}
	req.Header.Set("Content-Type", "application/json")
	req.Host = fmt.Sprintf("%s.%s.%s",
		modelName, request.GetNamespace(), service.customDomain)

	// Send the prediction request
	client := http.Client{Timeout: time.Duration(service.timeout) * time.Second}
	res, err := client.Do(req)
	if err != nil {
		return nil, NewRequestError(SendRequestError,
			errors.New("failed to send request, model not ready"))
	}
	if res.StatusCode != 200 {
		return nil, NewRequestError(InvalidInputError,
			errors.New("invalid model or inputs"))
	}

	// Parse the response
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, NewRequestError(ReadResponseError,
			errors.New("failed to read response body"))
	}
	var outputs map[string]interface{}
	err = json.Unmarshal(body, &outputs)
	if err != nil {
		return nil, NewRequestError(UnmarshalResponseError,
			errors.New("failed to unmarshal response body"))
	}
	response := InferResponse{Outputs: outputs}
	return &response, nil
}
