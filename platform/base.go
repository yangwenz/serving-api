package platform

type InferRequest interface {
	GetModelName() string
	GetModelVersion() string
	GetInputs() map[string]interface{}
	GetNamespace() string
}

type InferResponse struct {
	Outputs map[string]interface{}
}

type Platform interface {
	Predict(request InferRequest, version string) (*InferResponse, error)
}
