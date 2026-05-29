//go:build ignore

package task

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"{{.ModuleName}}/internal/platform/apierr"
	"{{.ModuleName}}/internal/platform/config"
)

type Manager struct {
	repo           Repository
	queue          Queue
	logger         *slog.Logger
	cfg            config.TaskConfig
	storageBackend string
	queueBackend   string
}

func NewManager(repo Repository, queue Queue, logger *slog.Logger, cfg config.TaskConfig, storageBackend, queueBackend string) *Manager {
	return &Manager{
		repo:           repo,
		queue:          queue,
		logger:         logger,
		cfg:            cfg,
		storageBackend: storageBackend,
		queueBackend:   queueBackend,
	}
}

// Submit 先持久化再入队；入队失败时将任务标记为 failed 并返回错误。
func (m *Manager) Submit(ctx context.Context, sub Submission) (*Record, error) {
	payload, err := json.Marshal(sub.Payload)
	if err != nil {
		return nil, apierr.BadRequest("invalid task payload")
	}

	now := time.Now().UTC()
	taskID := strings.TrimSpace(sub.TaskID)
	if taskID == "" {
		prefix := strings.TrimSpace(sub.IDPrefix)
		if prefix == "" {
			prefix = sub.Agent
		}
		taskID = NewTaskID(prefix)
	}

	record := newRecord(taskID, sub.Agent, sub.Type, payload, sub.ReplayedFrom, sub.TraceID, now)
	if err := m.repo.Create(ctx, record); err != nil {
		m.logger.Error("create task record failed", "agent", sub.Agent, "type", sub.Type, "error", err)
		return nil, apierr.Internal("create task record failed")
	}

	message := Message{
		TaskID:       record.ID,
		Agent:        record.Agent,
		Type:         record.Type,
		Payload:      record.Payload,
		ReplayedFrom: record.ReplayedFrom,
		TraceID:      record.TraceID,
		EnqueuedAt:   now,
		Attempt:      1,
	}

	messageID, err := m.queue.Enqueue(ctx, message)
	if err != nil {
		finishedAt := time.Now().UTC()
		if updateErr := m.repo.SetFailed(ctx, record.ID, err.Error(), finishedAt); updateErr != nil {
			m.logger.Error("mark enqueue failure failed", "task_id", record.ID, "error", updateErr)
		}
		m.logger.Error("enqueue task failed", "task_id", record.ID, "error", err)
		return nil, apierr.Internal("enqueue task failed")
	}

	if err := m.repo.SetQueued(ctx, record.ID, messageID); err != nil {
		m.logger.Error("update queued state failed", "task_id", record.ID, "error", err)
		return nil, apierr.Internal("update task queue state failed")
	}

	record.Status = StatusQueued
	record.QueueMessageID = messageID
	record.UpdatedAt = time.Now().UTC()
	return &record, nil
}

func (m *Manager) Get(ctx context.Context, taskID string) (*Record, bool, error) {
	record, found, err := m.repo.Get(ctx, taskID)
	if err != nil {
		m.logger.Error("query task failed", "task_id", taskID, "error", err)
		return nil, false, apierr.Internal("query task failed")
	}
	if !found {
		return nil, false, nil
	}
	return &record, true, nil
}

// Replay 仅允许重放 failed 任务，并生成新的 task_id。
func (m *Manager) Replay(ctx context.Context, taskID string) (*Record, error) {
	record, found, err := m.repo.Get(ctx, taskID)
	if err != nil {
		m.logger.Error("query task failed", "task_id", taskID, "error", err)
		return nil, apierr.Internal("query task failed")
	}
	if !found {
		return nil, apierr.NotFound("task not found")
	}
	if record.Status != StatusFailed {
		return nil, apierr.New(http.StatusBadRequest, apierr.CodeTaskNotAllowed, "only failed tasks can be replayed")
	}

	var payload any
	if len(record.Payload) > 0 {
		payload = json.RawMessage(record.Payload)
	}
	return m.Submit(ctx, Submission{
		Agent:        record.Agent,
		Type:         record.Type,
		Payload:      payload,
		ReplayedFrom: record.ID,
		TraceID:      record.TraceID,
	})
}

func (m *Manager) MergeMetadata(ctx context.Context, taskID string, fields map[string]any) error {
	if err := m.repo.MergeMetadata(ctx, taskID, fields); err != nil {
		m.logger.Error("merge metadata failed", "task_id", taskID, "error", err)
		return apierr.Internal(fmt.Sprintf("merge metadata failed: %v", err))
	}
	return nil
}

func (m *Manager) IncrementTokenUsage(ctx context.Context, taskID string, inputDelta, outputDelta int) error {
	if inputDelta <= 0 && outputDelta <= 0 {
		return nil
	}
	if err := m.repo.IncrementTokenUsage(ctx, taskID, inputDelta, outputDelta); err != nil {
		m.logger.Error("increment token usage failed", "task_id", taskID, "error", err)
		return apierr.Internal("increment token usage failed")
	}
	return nil
}

func (m *Manager) Health(ctx context.Context) (*HealthSnapshot, error) {
	count, err := m.repo.Count(ctx)
	if err != nil {
		m.logger.Error("task health query failed", "error", err)
		return nil, apierr.Internal("task health query failed")
	}
	return &HealthSnapshot{
		StorageBackend: m.storageBackend,
		QueueBackend:   m.queueBackend,
		TaskCount:      count,
	}, nil
}
