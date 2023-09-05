package platform

type InferRequest struct {
	ModelName    string                 `json:"model_name" binding:"required"`
	ModelVersion string                 `json:"model_version"`
	Inputs       map[string]interface{} `json:"inputs" binding:"required"`
}

type InferResponse struct {
	Outputs map[string]interface{}
}

type Platform interface {
	Predict(request *InferRequest, version string) (*InferResponse, *RequestError)
}

type Webhook interface {
	CreateNewTask(taskID string, modelName string, modelVersion string) (string, error)
	UpdateTaskInfo(info TaskInfo) error
}
