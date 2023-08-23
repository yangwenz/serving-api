package platform

type InferRequest interface {
	GetModelName() string
	GetModelVersion() string
	GetInputs() map[string]interface{}
}

type InferResponse interface {
	GetOutputs() map[string]interface{}
}

type Platform interface {
	Predict(request InferRequest, version string) (InferResponse, error)
}
