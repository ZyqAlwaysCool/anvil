package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/zyq/anvil/internal/platform/task"
	"github.com/zyq/anvil/test/testkit"
)

func TestTaskHTTPCreateQueryReplayMetadata(t *testing.T) {
	mgr, repo, _ := testkit.NewTaskManager(t)
	engine := testkit.NewTaskRouter(t, mgr)

	createBody, _ := json.Marshal(map[string]any{
		"agent":     "example",
		"task_type": "agent.example.greet",
		"payload":   map[string]string{"name": "world"},
	})
	createRec := testkit.PerformJSON(t, engine, http.MethodPost, "/api/v1/tasks/create", bytes.NewReader(createBody))
	if createRec.Code != http.StatusAccepted {
		t.Fatalf("create status=%d body=%s", createRec.Code, createRec.Body.String())
	}

	var createResp struct {
		Data struct {
			TaskID string `json:"task_id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(createRec.Body.Bytes(), &createResp); err != nil {
		t.Fatalf("unmarshal create response: %v", err)
	}
	if createResp.Data.TaskID == "" {
		t.Fatal("task_id should not be empty")
	}

	queryRec := testkit.PerformJSON(t, engine, http.MethodGet, "/api/v1/tasks/query?task_id="+createResp.Data.TaskID, nil)
	if queryRec.Code != http.StatusOK {
		t.Fatalf("query status=%d body=%s", queryRec.Code, queryRec.Body.String())
	}

	metaBody, _ := json.Marshal(map[string]any{
		"task_id":  createResp.Data.TaskID,
		"metadata": map[string]string{"biz_key": "value"},
	})
	metaRec := testkit.PerformJSON(t, engine, http.MethodPatch, "/api/v1/tasks/metadata", bytes.NewReader(metaBody))
	if metaRec.Code != http.StatusOK {
		t.Fatalf("metadata status=%d body=%s", metaRec.Code, metaRec.Body.String())
	}

	record, found, err := mgr.Get(t.Context(), createResp.Data.TaskID)
	if err != nil || !found {
		t.Fatalf("get task: found=%v err=%v", found, err)
	}
	if record.Metadata["biz_key"] != "value" {
		t.Fatalf("unexpected metadata: %+v", record.Metadata)
	}

	failed := *record
	failed.Status = task.StatusFailed
	repo.Seed(failed)

	replayBody, _ := json.Marshal(map[string]string{"task_id": createResp.Data.TaskID})
	replayRec := testkit.PerformJSON(t, engine, http.MethodPost, "/api/v1/tasks/replay", bytes.NewReader(replayBody))
	if replayRec.Code != http.StatusAccepted {
		t.Fatalf("replay status=%d body=%s", replayRec.Code, replayRec.Body.String())
	}

	var replayResp struct {
		Data struct {
			TaskID string `json:"task_id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(replayRec.Body.Bytes(), &replayResp); err != nil {
		t.Fatalf("unmarshal replay response: %v", err)
	}
	if replayResp.Data.TaskID == "" || replayResp.Data.TaskID == createResp.Data.TaskID {
		t.Fatalf("expected new task_id, got %s", replayResp.Data.TaskID)
	}
}
