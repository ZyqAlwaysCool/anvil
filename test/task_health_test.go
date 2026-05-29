package test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/ZyqAlwaysCool/anvil/internal/platform/task"
	"github.com/ZyqAlwaysCool/anvil/test/testkit"
)

func TestTaskManagerHealthSnapshot(t *testing.T) {
	mgr, _, _ := testkit.NewTaskManager(t)
	_, err := mgr.Submit(context.Background(), task.Submission{
		Agent:   "example",
		Type:    "agent.example.greet",
		Payload: map[string]string{"name": "world"},
	})
	if err != nil {
		t.Fatalf("submit: %v", err)
	}

	snapshot, err := mgr.Health(context.Background())
	if err != nil {
		t.Fatalf("health: %v", err)
	}
	if snapshot.StorageBackend != "memory" {
		t.Fatalf("unexpected storage backend: %s", snapshot.StorageBackend)
	}
	if snapshot.QueueBackend != "memory" {
		t.Fatalf("unexpected queue backend: %s", snapshot.QueueBackend)
	}
	if snapshot.TaskCount != 1 {
		t.Fatalf("unexpected task count: %d", snapshot.TaskCount)
	}
}

func TestTaskHTTPHealthContract(t *testing.T) {
	mgr, _, _ := testkit.NewTaskManager(t)
	engine := testkit.NewTaskRouter(t, mgr)
	rec := testkit.PerformJSON(t, engine, http.MethodGet, "/api/v1/tasks/health", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("health status=%d body=%s", rec.Code, rec.Body.String())
	}

	var body struct {
		Data task.HealthSnapshot `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal health: %v", err)
	}
	if body.Data.StorageBackend == "" || body.Data.QueueBackend == "" {
		t.Fatalf("health snapshot missing backend info: %+v", body.Data)
	}
}
