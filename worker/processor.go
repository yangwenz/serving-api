package worker

import (
	"context"
	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
	"github.com/yangwenz/model-serving/platform"
	"github.com/yangwenz/model-serving/utils"
)

const (
	QueueCritical = "critical"
	QueueDefault  = "default"
)

type RedisTaskProcessor struct {
	config   utils.Config
	server   *asynq.Server
	platform platform.Platform
	webhook  platform.Webhook
}

func NewRedisTaskProcessor(config utils.Config, platform platform.Platform, webhook platform.Webhook) *RedisTaskProcessor {
	redisOpt := asynq.RedisClientOpt{
		Addr: config.RedisAddress,
	}
	logger := NewLogger()
	redis.SetLogger(logger)

	server := asynq.NewServer(
		redisOpt,
		asynq.Config{
			Concurrency: config.WorkerConcurrency,
			Queues: map[string]int{
				QueueCritical: 10,
				QueueDefault:  5,
			},
			ErrorHandler: asynq.ErrorHandlerFunc(func(ctx context.Context, task *asynq.Task, err error) {
				log.Error().Err(err).Str("type", task.Type()).
					Bytes("payload", task.Payload()).Msg("process task failed")
			}),
			Logger: logger,
		},
	)

	return &RedisTaskProcessor{
		config:   config,
		server:   server,
		platform: platform,
		webhook:  webhook,
	}
}

func (processor *RedisTaskProcessor) Start() error {
	mux := asynq.NewServeMux()
	mux.HandleFunc(TaskRunPrediction, processor.ProcessTaskRunPrediction)
	return processor.server.Start(mux)
}
