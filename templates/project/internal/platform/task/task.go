//go:build ignore

package task

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"
)

type Status string

const (
	StatusPending Status = "pending"
	StatusQueued  Status = "queued"
	StatusRunning Status = "running"
	StatusSuccess Status = "success"
	StatusFailed  Status = "failed"
)

const (
	MetadataTokenInputKey  = "token_input_total"
	MetadataTokenOutputKey = "token_output_total"
)

type Record struct {
	ID             string          `json:"task_id" bson:"_id"`
	Agent          string          `json:"agent" bson:"agent"`
	Type           string          `json:"type" bson:"type"`
	Status         Status          `json:"status" bson:"status"`
	Payload        json.RawMessage `json:"payload,omitempty" bson:"payload,omitempty"`
	Result         json.RawMessage `json:"result,omitempty" bson:"result,omitempty"`
	Metadata       map[string]any  `json:"metadata,omitempty" bson:"metadata,omitempty"`
	ErrorMessage   string          `json:"error_message,omitempty" bson:"error_message,omitempty"`
	QueueMessageID string          `json:"queue_message_id,omitempty" bson:"queue_message_id,omitempty"`
	ReplayedFrom   string          `json:"replayed_from,omitempty" bson:"replayed_from,omitempty"`
	TraceID        string          `json:"trace_id,omitempty" bson:"trace_id,omitempty"`
	CreatedAt      time.Time       `json:"created_at" bson:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at" bson:"updated_at"`
	StartedAt      *time.Time      `json:"started_at,omitempty" bson:"started_at,omitempty"`
	FinishedAt     *time.Time      `json:"finished_at,omitempty" bson:"finished_at,omitempty"`
}

type Submission struct {
	Agent        string
	Type         string
	Payload      any
	ReplayedFrom string
	IDPrefix     string
	TaskID       string
	TraceID      string
}

type Message struct {
	TaskID             string          `json:"task_id"`
	Agent              string          `json:"agent"`
	Type               string          `json:"type"`
	Payload            json.RawMessage `json:"payload"`
	ReplayedFrom       string          `json:"replayed_from,omitempty"`
	TraceID            string          `json:"trace_id,omitempty"`
	EnqueuedAt         time.Time       `json:"enqueued_at"`
	WorkerConsumerName string          `json:"worker_consumer_name"`
	Attempt            int             `json:"attempt,omitempty"`
}

type QueuedMessage struct {
	QueueMessageID string
	Message        Message
}

type HealthSnapshot struct {
	StorageBackend string `json:"storage_backend"`
	QueueBackend   string `json:"queue_backend"`
	TaskCount      int64  `json:"task_count"`
}

func NewTaskID(prefix string) string {
	randomBytes := make([]byte, 4)
	if _, err := rand.Read(randomBytes); err != nil {
		now := time.Now().UTC()
		return fmt.Sprintf("%s_%s_%d", prefix, now.Format("20060102_150405"), now.UnixNano())
	}
	now := time.Now().UTC()
	return fmt.Sprintf("%s_%s_%s", prefix, now.Format("20060102_150405"), hex.EncodeToString(randomBytes))
}

func newRecord(taskID, agent, taskType string, payload json.RawMessage, replayedFrom, traceID string, now time.Time) Record {
	return Record{
		ID:           taskID,
		Agent:        agent,
		Type:         taskType,
		Status:       StatusPending,
		Payload:      payload,
		ReplayedFrom: replayedFrom,
		TraceID:      traceID,
		Metadata:     map[string]any{},
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}
