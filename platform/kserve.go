package platform

type KServe struct {
	address string
}

func NewKServe(address string) Platform {
	return &KServe{address: address}
}

func (service *KServe) Predict(request InferRequest, version string) (*InferResponse, error) {
	return nil, nil
}
