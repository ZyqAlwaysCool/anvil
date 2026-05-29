package example

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/ZyqAlwaysCool/anvil/internal/platform/task"
)

type Service struct {
	manager *task.Manager
}

func NewService(manager *task.Manager) *Service {
	return &Service{manager: manager}
}

// CreateTask 只负责提交异步任务；不在 HTTP 请求内执行业务长流程。
func (s *Service) CreateTask(ctx context.Context, req CreateTaskRequest, traceID string) (*CreateTaskResponse, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, invalidNameError()
	}

	record, err := s.manager.Submit(ctx, task.Submission{
		Agent:    AgentName,
		Type:     TaskType,
		Payload:  map[string]string{"name": name},
		TraceID:  traceID,
		IDPrefix: AgentName,
	})
	if err != nil {
		return nil, err
	}
	return &CreateTaskResponse{TaskID: record.ID}, nil
}

// QueryTask 读取任务当前状态与结果，供客户端轮询异步执行进度。
func (s *Service) QueryTask(ctx context.Context, taskID string) (*QueryTaskResponse, error) {
	record, found, err := s.manager.Get(ctx, taskID)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, taskNotFoundError()
	}

	resp := &QueryTaskResponse{
		TaskID: record.ID,
		Status: string(record.Status),
	}
	if len(record.Result) > 0 {
		var result GreetResult
		if err := json.Unmarshal(record.Result, &result); err == nil {
			resp.Result = &result
		}
	}
	return resp, nil
}

// HandleGreetTask 由 Worker 异步执行，模拟约 1 秒业务处理时延。
func HandleGreetTask(_ context.Context, message task.Message) (json.RawMessage, error) {
	var payload struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal(message.Payload, &payload); err != nil {
		return nil, err
	}
	time.Sleep(time.Second)
	result := GreetResult{Greeting: "Hello, " + strings.TrimSpace(payload.Name) + "!"}
	raw, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}
	return raw, nil
}
