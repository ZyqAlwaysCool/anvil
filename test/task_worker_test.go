package test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/ZyqAlwaysCool/anvil/internal/platform/task"
	"github.com/ZyqAlwaysCool/anvil/test/testkit"
)

func TestTaskWorkerSuccess(t *testing.T) {
	repo := testkit.NewMemoryRepository()
	queue := testkit.NewMemoryQueue()
	registry := task.NewRegistry()
	_ = registry.Register("agent.example.greet", task.HandlerFunc(func(_ context.Context, message task.Message) (json.RawMessage, error) {
		raw, _ := json.Marshal(map[string]string{"greeting": "Hello, world!"})
		return raw, nil
	}))

	mgr := task.NewManager(repo, queue, testkit.NewTestLogger(), testkit.DefaultTaskConfig(), "memory", "memory")
	record, err := mgr.Submit(context.Background(), task.Submission{
		Agent:   "example",
		Type:    "agent.example.greet",
		Payload: map[string]string{"name": "world"},
	})
	if err != nil {
		t.Fatalf("submit: %v", err)
	}

	worker := task.NewWorker(repo, queue, registry, testkit.NewTestLogger(), testkit.DefaultTaskConfig())
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	go func() {
		_ = worker.Run(ctx)
	}()

	waitForStatus(t, repo, record.ID, task.StatusSuccess)
}

func TestTaskWorkerRetryThenSuccess(t *testing.T) {
	repo := testkit.NewMemoryRepository()
	queue := testkit.NewMemoryQueue()
	registry := task.NewRegistry()
	attempts := 0
	_ = registry.Register("agent.example.retry", task.HandlerFunc(func(_ context.Context, _ task.Message) (json.RawMessage, error) {
		attempts++
		if attempts < 2 {
			return nil, errors.New("temporary failure")
		}
		return json.RawMessage(`{"ok":true}`), nil
	}))

	cfg := testkit.DefaultTaskConfig()
	cfg.RetryJobs = true
	cfg.MaxTries = 3
	mgr := task.NewManager(repo, queue, testkit.NewTestLogger(), cfg, "memory", "memory")
	record, err := mgr.Submit(context.Background(), task.Submission{
		Agent: "example",
		Type:  "agent.example.retry",
	})
	if err != nil {
		t.Fatalf("submit: %v", err)
	}

	worker := task.NewWorker(repo, queue, registry, testkit.NewTestLogger(), cfg)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	go func() {
		_ = worker.Run(ctx)
	}()

	waitForStatus(t, repo, record.ID, task.StatusSuccess)
}

func TestTaskWorkerExceedMaxRetries(t *testing.T) {
	repo := testkit.NewMemoryRepository()
	queue := testkit.NewMemoryQueue()
	registry := task.NewRegistry()
	_ = registry.Register("agent.example.always_fail", task.HandlerFunc(func(_ context.Context, _ task.Message) (json.RawMessage, error) {
		return nil, errors.New("always fail")
	}))

	cfg := testkit.DefaultTaskConfig()
	cfg.RetryJobs = true
	cfg.MaxTries = 2
	mgr := task.NewManager(repo, queue, testkit.NewTestLogger(), cfg, "memory", "memory")
	record, err := mgr.Submit(context.Background(), task.Submission{
		Agent: "example",
		Type:  "agent.example.always_fail",
	})
	if err != nil {
		t.Fatalf("submit: %v", err)
	}

	worker := task.NewWorker(repo, queue, registry, testkit.NewTestLogger(), cfg)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	go func() {
		_ = worker.Run(ctx)
	}()

	waitForStatus(t, repo, record.ID, task.StatusFailed)
}

func waitForStatus(t *testing.T, repo *testkit.MemoryRepository, taskID string, expected task.Status) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		record, found, err := repo.Get(context.Background(), taskID)
		if err == nil && found && record.Status == expected {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	record, found, _ := repo.Get(context.Background(), taskID)
	t.Fatalf("timeout waiting status=%s found=%v current=%v", expected, found, record.Status)
}
