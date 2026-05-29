package redisstream

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/zyq/anvil/internal/platform/config"
	"github.com/zyq/anvil/internal/platform/task"
)

type Queue struct {
	client        redis.UniversalClient
	streamKey     string
	consumerGroup string
	readBlock     time.Duration
	maxLen        int64
	logger        *slog.Logger

	groupMu    sync.Mutex
	groupReady bool
}

// NewQueue 构造 Redis Stream 队列；consumer group 在首次读写前懒初始化，不泄漏到 task.Queue 公共接口。
func NewQueue(client redis.UniversalClient, cfg config.TaskConfig, logger *slog.Logger) *Queue {
	return &Queue{
		client:        client,
		streamKey:     cfg.StreamKey,
		consumerGroup: cfg.ConsumerGroup,
		readBlock:     cfg.ReadBlock,
		maxLen:        cfg.MaxLen,
		logger:        logger,
	}
}

func (q *Queue) ensureConsumerGroup(ctx context.Context) error {
	q.groupMu.Lock()
	defer q.groupMu.Unlock()

	if q.groupReady {
		return nil
	}

	err := q.client.XGroupCreateMkStream(ctx, q.streamKey, q.consumerGroup, "0").Err()
	if err == nil || strings.Contains(err.Error(), "BUSYGROUP") {
		q.groupReady = true
		return nil
	}
	// 失败不缓存，下次调用仍可重试，避免 sync.Once 把瞬时错误永久化。
	return fmt.Errorf("create redis consumer group: %w", err)
}

func (q *Queue) Enqueue(ctx context.Context, msg task.Message) (string, error) {
	if err := q.ensureConsumerGroup(ctx); err != nil {
		return "", err
	}

	body, err := json.Marshal(msg)
	if err != nil {
		return "", fmt.Errorf("marshal task message: %w", err)
	}
	messageID, err := q.client.XAdd(ctx, &redis.XAddArgs{
		Stream: q.streamKey,
		MaxLen: q.maxLen,
		Approx: true,
		Values: map[string]any{"data": string(body)},
	}).Result()
	if err != nil {
		return "", fmt.Errorf("enqueue redis stream message: %w", err)
	}
	return messageID, nil
}

func (q *Queue) Read(ctx context.Context, consumer string, count int) ([]task.QueuedMessage, error) {
	if err := q.ensureConsumerGroup(ctx); err != nil {
		return nil, err
	}

	if count <= 0 {
		count = 1
	}
	streams, err := q.client.XReadGroup(ctx, &redis.XReadGroupArgs{
		Group:    q.consumerGroup,
		Consumer: consumer,
		Streams:  []string{q.streamKey, ">"},
		Count:    int64(count),
		Block:    q.readBlock,
	}).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, fmt.Errorf("read redis stream: %w", err)
	}

	result := make([]task.QueuedMessage, 0)
	for _, stream := range streams {
		for _, item := range stream.Messages {
			body, err := extractPayload(item.Values)
			if err != nil {
				return nil, err
			}
			var message task.Message
			if err := json.Unmarshal(body, &message); err != nil {
				return nil, fmt.Errorf("unmarshal task message: %w", err)
			}
			result = append(result, task.QueuedMessage{
				QueueMessageID: item.ID,
				Message:        message,
			})
		}
	}
	return result, nil
}

func (q *Queue) Ack(ctx context.Context, queueMessageID string) error {
	if queueMessageID == "" {
		return nil
	}
	if err := q.client.XAck(ctx, q.streamKey, q.consumerGroup, queueMessageID).Err(); err != nil {
		return fmt.Errorf("ack redis stream message: %w", err)
	}
	if err := q.client.XDel(ctx, q.streamKey, queueMessageID).Err(); err != nil {
		return fmt.Errorf("delete redis stream message: %w", err)
	}
	return nil
}

func (q *Queue) Retry(ctx context.Context, queueMessageID string, msg task.Message) (string, error) {
	newID, err := q.Enqueue(ctx, msg)
	if err != nil {
		return "", err
	}
	if err := q.Ack(ctx, queueMessageID); err != nil {
		// 部分成功：重试消息已入队，但旧 delivery 未 Ack，旧消息可能被再次投递；handler 必须幂等。
		q.logger.Warn("retry partial success",
			"task_id", msg.TaskID,
			"old_queue_message_id", queueMessageID,
			"new_queue_message_id", newID,
			"error", err,
		)
		return newID, fmt.Errorf("retry ack old message failed: %w", err)
	}
	return newID, nil
}

func extractPayload(values map[string]any) ([]byte, error) {
	value, ok := values["data"]
	if !ok {
		return nil, fmt.Errorf("redis stream message missing data field")
	}
	switch body := value.(type) {
	case string:
		return []byte(body), nil
	case []byte:
		return body, nil
	default:
		return nil, fmt.Errorf("unsupported redis stream payload type: %T", value)
	}
}

var _ task.Queue = (*Queue)(nil)
