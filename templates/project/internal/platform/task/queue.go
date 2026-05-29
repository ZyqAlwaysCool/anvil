//go:build ignore

package task

import "context"

// Queue 只暴露任务流转所需能力；consumer group 初始化属于 Redis Stream 实现细节。
type Queue interface {
	Enqueue(ctx context.Context, msg Message) (string, error)
	Read(ctx context.Context, consumer string, count int) ([]QueuedMessage, error)
	Ack(ctx context.Context, queueMessageID string) error
	Retry(ctx context.Context, queueMessageID string, msg Message) (string, error)
}
