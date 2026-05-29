package test

import (
	"context"
	"testing"

	"github.com/ZyqAlwaysCool/anvil/internal/platform/task"
	"github.com/ZyqAlwaysCool/anvil/test/testkit"
)

func TestTaskManagerSubmit(t *testing.T) {
	mgr, repo, _ := testkit.NewTaskManager(t)
	record, err := mgr.Submit(context.Background(), task.Submission{
		Agent:   "example",
		Type:    "agent.example.greet",
		Payload: map[string]string{"name": "world"},
	})
	if err != nil {
		t.Fatalf("submit task: %v", err)
	}
	stored, found, err := repo.Get(context.Background(), record.ID)
	if err != nil || !found {
		t.Fatalf("get task: found=%v err=%v", found, err)
	}
	if stored.Status != task.StatusQueued {
		t.Fatalf("expected queued status, got %s", stored.Status)
	}
}

func TestTaskManagerSubmitEnqueueFailure(t *testing.T) {
	mgr, repo, queue := testkit.NewTaskManager(t)
	queue.FailNextEnqueue()
	_, err := mgr.Submit(context.Background(), task.Submission{
		Agent:   "example",
		Type:    "agent.example.greet",
		Payload: map[string]string{"name": "world"},
	})
	if err == nil {
		t.Fatal("expected enqueue failure")
	}
	stored, found := repo.First()
	if !found {
		t.Fatal("expected stored task record")
	}
	if stored.Status != task.StatusFailed {
		t.Fatalf("expected failed status, got %s", stored.Status)
	}
}

func TestTaskManagerReplayOnlyFailed(t *testing.T) {
	mgr, repo, _ := testkit.NewTaskManager(t)
	record, err := mgr.Submit(context.Background(), task.Submission{
		Agent:   "example",
		Type:    "agent.example.greet",
		Payload: map[string]string{"name": "world"},
	})
	if err != nil {
		t.Fatalf("submit task: %v", err)
	}
	if _, err := mgr.Replay(context.Background(), record.ID); err == nil {
		t.Fatal("expected replay not allowed for non-failed task")
	}

	finished := *record
	finished.Status = task.StatusFailed
	repo.Seed(finished)

	replayed, err := mgr.Replay(context.Background(), record.ID)
	if err != nil {
		t.Fatalf("replay task: %v", err)
	}
	if replayed.ReplayedFrom != record.ID {
		t.Fatalf("expected replayed_from=%s got=%s", record.ID, replayed.ReplayedFrom)
	}
}

func TestTaskManagerMetadataAndTokenUsage(t *testing.T) {
	mgr, repo, _ := testkit.NewTaskManager(t)
	record, err := mgr.Submit(context.Background(), task.Submission{
		Agent:   "example",
		Type:    "agent.example.greet",
		Payload: map[string]string{"name": "world"},
	})
	if err != nil {
		t.Fatalf("submit task: %v", err)
	}
	if err := mgr.MergeMetadata(context.Background(), record.ID, map[string]any{"biz_key": "value"}); err != nil {
		t.Fatalf("merge metadata: %v", err)
	}
	if err := mgr.IncrementTokenUsage(context.Background(), record.ID, 10, 5); err != nil {
		t.Fatalf("increment token usage: %v", err)
	}
	stored, found, err := repo.Get(context.Background(), record.ID)
	if err != nil || !found {
		t.Fatalf("get task: found=%v err=%v", found, err)
	}
	if stored.Metadata["biz_key"] != "value" {
		t.Fatalf("unexpected metadata: %+v", stored.Metadata)
	}
}
