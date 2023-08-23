package platform

import (
	"fmt"
	"github.com/yangwenz/model-serving/utils"
)

type KServe struct {
	address      string
	customDomain string
}

func NewKServe(config utils.Config) Platform {
	return &KServe{
		address:      config.KServeAddress,
		customDomain: config.KServeCustomDomain,
	}
}

func (service *KServe) Predict(request InferRequest, version string) (*InferResponse, error) {
	modelName := request.GetModelName()
	if version == "v1" {
		host := fmt.Sprintf("%s.%s", modelName, service.customDomain)
		url := fmt.Sprintf("%s/v1/models/%s:predict", service.address, modelName)
		fmt.Println(host)
		fmt.Println(url)
	}
	return nil, nil
}
