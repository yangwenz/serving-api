package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/hibiken/asynq"
	"github.com/rs/zerolog/log"
	"github.com/yangwenz/model-serving/platform"
)

const TaskRunPrediction = "task:run_prediction"

type PayloadRunPrediction struct {
	platform.InferRequest
	ID         string `json:"id"`
	APIVersion string `json:"api_version" default:"v1"`
}

func (distributor *RedisTaskDistributor) DistributeTaskRunPrediction(
	ctx context.Context,
	payload *PayloadRunPrediction,
	opts ...asynq.Option,
) error {
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal task payload: %w", err)
	}

	task := asynq.NewTask(TaskRunPrediction, jsonPayload, opts...)
	info, err := distributor.client.EnqueueContext(ctx, task)
	if err != nil {
		return fmt.Errorf("failed to enqueue task: %w", err)
	}

	log.Info().Str("type", task.Type()).Str("queue", info.Queue).
		Int("max_retry", info.MaxRetry).Msg("enqueued task")
	return nil
}

func (processor *RedisTaskProcessor) ProcessTaskRunPrediction(
	ctx context.Context,
	task *asynq.Task,
) error {
	var payload PayloadRunPrediction
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", asynq.SkipRetry)
	}

	info := platform.TaskInfo{ID: payload.ID}
	response, err := processor.platform.Predict(&payload.InferRequest, payload.APIVersion)
	if err != nil {
		info.Status = "failed"
		err := processor.webhook.UpdateTaskInfo(info)
		if err != nil {
			return fmt.Errorf("failed to update task %s: %w", payload.ID, err)
		}
		return fmt.Errorf("failed to run prediction: %w", err)
	}

	output, e := json.Marshal(response)
	if e != nil {
		info.Status = "failed"
		err := processor.webhook.UpdateTaskInfo(info)
		if err != nil {
			return fmt.Errorf("failed to update task %s: %w", payload.ID, err)
		}
		return fmt.Errorf("failed to marshal output: %w", err)
	}
	// TODO: Update running time
	info.Status = "succeeded"
	info.Outputs = output
	if err := processor.webhook.UpdateTaskInfo(info); err != nil {
		return fmt.Errorf("failed to update task %s: %w", payload.ID, err)
	}
	return nil
}
