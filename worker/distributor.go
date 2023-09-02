package worker

import (
	"context"
	"github.com/hibiken/asynq"
	"github.com/yangwenz/model-serving/utils"
)

type TaskDistributor interface {
	DistributeTaskRunPrediction(
		ctx context.Context,
		payload *PayloadRunPrediction,
		opts ...asynq.Option,
	) error
}

type RedisTaskDistributor struct {
	client *asynq.Client
}

func NewRedisTaskDistributor(config utils.Config) TaskDistributor {
	if config.RedisAddress == "" {
		return nil
	}
	redisOpt := asynq.RedisClientOpt{
		Addr: config.RedisAddress,
	}
	client := asynq.NewClient(redisOpt)
	return &RedisTaskDistributor{
		client: client,
	}
}
