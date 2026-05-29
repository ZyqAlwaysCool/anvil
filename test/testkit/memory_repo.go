package testkit

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/zyq/anvil/internal/platform/task"
)

type MemoryRepository struct {
	mu      sync.RWMutex
	records map[string]task.Record
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{records: make(map[string]task.Record)}
}

func (r *MemoryRepository) Create(_ context.Context, record task.Record) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.records[record.ID]; exists {
		return fmt.Errorf("task already exists: %s", record.ID)
	}
	if record.Metadata == nil {
		record.Metadata = map[string]any{}
	}
	r.records[record.ID] = record
	return nil
}

func (r *MemoryRepository) Get(_ context.Context, taskID string) (task.Record, bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	record, ok := r.records[taskID]
	return record, ok, nil
}

func (r *MemoryRepository) SetQueued(_ context.Context, taskID, queueMessageID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	record, ok := r.records[taskID]
	if !ok {
		return fmt.Errorf("task not found: %s", taskID)
	}
	record.Status = task.StatusQueued
	record.QueueMessageID = queueMessageID
	record.UpdatedAt = time.Now().UTC()
	r.records[taskID] = record
	return nil
}

func (r *MemoryRepository) SetRunning(_ context.Context, taskID string, startedAt time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	record, ok := r.records[taskID]
	if !ok {
		return fmt.Errorf("task not found: %s", taskID)
	}
	record.Status = task.StatusRunning
	record.StartedAt = &startedAt
	record.UpdatedAt = startedAt
	r.records[taskID] = record
	return nil
}

func (r *MemoryRepository) SetSuccess(_ context.Context, taskID string, result json.RawMessage, finishedAt time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	record, ok := r.records[taskID]
	if !ok {
		return fmt.Errorf("task not found: %s", taskID)
	}
	record.Status = task.StatusSuccess
	record.Result = append(json.RawMessage(nil), result...)
	record.FinishedAt = &finishedAt
	record.UpdatedAt = finishedAt
	r.records[taskID] = record
	return nil
}

func (r *MemoryRepository) SetFailed(_ context.Context, taskID, errMsg string, finishedAt time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	record, ok := r.records[taskID]
	if !ok {
		return fmt.Errorf("task not found: %s", taskID)
	}
	record.Status = task.StatusFailed
	record.ErrorMessage = errMsg
	record.FinishedAt = &finishedAt
	record.UpdatedAt = finishedAt
	r.records[taskID] = record
	return nil
}

func (r *MemoryRepository) MergeMetadata(_ context.Context, taskID string, fields map[string]any) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	record, ok := r.records[taskID]
	if !ok {
		return fmt.Errorf("task not found: %s", taskID)
	}
	if record.Metadata == nil {
		record.Metadata = map[string]any{}
	}
	for key, value := range fields {
		record.Metadata[key] = value
	}
	record.UpdatedAt = time.Now().UTC()
	r.records[taskID] = record
	return nil
}

func (r *MemoryRepository) IncrementTokenUsage(_ context.Context, taskID string, inputDelta, outputDelta int) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	record, ok := r.records[taskID]
	if !ok {
		return fmt.Errorf("task not found: %s", taskID)
	}
	if record.Metadata == nil {
		record.Metadata = map[string]any{}
	}
	if inputDelta > 0 {
		prev, _ := record.Metadata[task.MetadataTokenInputKey].(int)
		record.Metadata[task.MetadataTokenInputKey] = prev + inputDelta
	}
	if outputDelta > 0 {
		prev, _ := record.Metadata[task.MetadataTokenOutputKey].(int)
		record.Metadata[task.MetadataTokenOutputKey] = prev + outputDelta
	}
	record.UpdatedAt = time.Now().UTC()
	r.records[taskID] = record
	return nil
}

func (r *MemoryRepository) Count(context.Context) (int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return int64(len(r.records)), nil
}

func (r *MemoryRepository) Seed(record task.Record) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if record.Metadata == nil {
		record.Metadata = map[string]any{}
	}
	r.records[record.ID] = record
}

func (r *MemoryRepository) First() (task.Record, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, record := range r.records {
		return record, true
	}
	return task.Record{}, false
}

var _ task.Repository = (*MemoryRepository)(nil)
