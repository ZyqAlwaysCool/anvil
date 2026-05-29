package test

import (
	"context"
	"strings"
	"testing"

	"github.com/zyq/anvil/internal/platform/task"
	"github.com/zyq/anvil/test/testkit"
)

func TestQueueRetryPartialSuccess(t *testing.T) {
	queue := testkit.NewMemoryQueue()
	queue.FailNextAck()

	newID, err := queue.Retry(context.Background(), "old-msg", task.Message{
		TaskID: "task-1",
		Type:   "agent.example.greet",
	})
	if err == nil {
		t.Fatal("expected partial success error")
	}
	if newID == "" {
		t.Fatal("expected new queue message id on partial success")
	}
	if !strings.Contains(err.Error(), "retry ack old message failed") {
		t.Fatalf("unexpected error: %v", err)
	}
}
