package task

import (
	"context"
	"encoding/json"
	"time"
)

type Repository interface {
	Create(ctx context.Context, record Record) error
	Get(ctx context.Context, taskID string) (Record, bool, error)
	SetQueued(ctx context.Context, taskID, queueMessageID string) error
	SetRunning(ctx context.Context, taskID string, startedAt time.Time) error
	SetSuccess(ctx context.Context, taskID string, result json.RawMessage, finishedAt time.Time) error
	SetFailed(ctx context.Context, taskID string, errMsg string, finishedAt time.Time) error
	MergeMetadata(ctx context.Context, taskID string, fields map[string]any) error
	IncrementTokenUsage(ctx context.Context, taskID string, inputDelta, outputDelta int) error
	Count(ctx context.Context) (int64, error)
}
