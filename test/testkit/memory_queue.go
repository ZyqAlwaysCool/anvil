package testkit

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/zyq/anvil/internal/platform/task"
)

type MemoryQueue struct {
	mu          sync.Mutex
	pending     []task.QueuedMessage
	failNext    bool
	failNextAck bool
}

func NewMemoryQueue() *MemoryQueue {
	return &MemoryQueue{pending: make([]task.QueuedMessage, 0)}
}

func (q *MemoryQueue) FailNextEnqueue() {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.failNext = true
}

func (q *MemoryQueue) FailNextAck() {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.failNextAck = true
}

func (q *MemoryQueue) Enqueue(_ context.Context, msg task.Message) (string, error) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.failNext {
		q.failNext = false
		return "", fmt.Errorf("enqueue failed")
	}
	messageID := fmt.Sprintf("msg-%s-%d", msg.TaskID, time.Now().UnixNano())
	q.pending = append(q.pending, task.QueuedMessage{
		QueueMessageID: messageID,
		Message:        msg,
	})
	return messageID, nil
}

func (q *MemoryQueue) Read(ctx context.Context, _ string, count int) ([]task.QueuedMessage, error) {
	if count <= 0 {
		count = 1
	}
	for {
		q.mu.Lock()
		if len(q.pending) > 0 {
			n := count
			if n > len(q.pending) {
				n = len(q.pending)
			}
			out := append([]task.QueuedMessage(nil), q.pending[:n]...)
			q.mu.Unlock()
			return out, nil
		}
		q.mu.Unlock()
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(5 * time.Millisecond):
		}
	}
}

func (q *MemoryQueue) Ack(_ context.Context, queueMessageID string) error {
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.failNextAck {
		q.failNextAck = false
		return fmt.Errorf("ack failed")
	}
	remaining := q.pending[:0]
	for _, item := range q.pending {
		if item.QueueMessageID == queueMessageID {
			continue
		}
		remaining = append(remaining, item)
	}
	q.pending = remaining
	return nil
}

func (q *MemoryQueue) Retry(_ context.Context, queueMessageID string, msg task.Message) (string, error) {
	newID, err := q.Enqueue(context.Background(), msg)
	if err != nil {
		return "", err
	}
	if err := q.Ack(context.Background(), queueMessageID); err != nil {
		return newID, fmt.Errorf("retry ack old message failed: %w", err)
	}
	return newID, nil
}

var _ task.Queue = (*MemoryQueue)(nil)
