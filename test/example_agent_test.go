package test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/zyq/anvil/internal/agent/biz/example"
	"github.com/zyq/anvil/internal/platform/apierr"
	httpmiddleware "github.com/zyq/anvil/internal/platform/http/middleware"
	"github.com/zyq/anvil/internal/platform/http/response"
	"github.com/zyq/anvil/internal/platform/task"
	"github.com/zyq/anvil/test/testkit"
)

func TestExampleAgentCreateAndWorkerComplete(t *testing.T) {
	repo := testkit.NewMemoryRepository()
	queue := testkit.NewMemoryQueue()
	mgr := task.NewManager(repo, queue, testkit.NewTestLogger(), testkit.DefaultTaskConfig(), "memory", "memory")
	registry := task.NewRegistry()
	_ = registry.Register(example.TaskType, task.HandlerFunc(example.HandleGreetTask))

	engine := gin.New()
	svc := example.NewService(mgr)
	registerExampleRoutes(engine.Group("/api/v1/agents/example"), svc)

	createBody, _ := json.Marshal(example.CreateTaskRequest{Name: "world"})
	createRec := testkit.PerformJSON(t, engine, http.MethodPost, "/api/v1/agents/example/task/create", bytes.NewReader(createBody))
	if createRec.Code != http.StatusAccepted {
		t.Fatalf("create status=%d body=%s", createRec.Code, createRec.Body.String())
	}

	var createResp struct {
		Data example.CreateTaskResponse `json:"data"`
	}
	if err := json.Unmarshal(createRec.Body.Bytes(), &createResp); err != nil {
		t.Fatalf("unmarshal create response: %v", err)
	}

	worker := task.NewWorker(repo, queue, registry, testkit.NewTestLogger(), testkit.DefaultTaskConfig())
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	go func() {
		_ = worker.Run(ctx)
	}()

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		queryRec := testkit.PerformJSON(t, engine, http.MethodGet, "/api/v1/agents/example/task/query?task_id="+createResp.Data.TaskID, nil)
		var queryResp struct {
			Data example.QueryTaskResponse `json:"data"`
		}
		if err := json.Unmarshal(queryRec.Body.Bytes(), &queryResp); err != nil {
			t.Fatalf("unmarshal query response: %v", err)
		}
		if queryResp.Data.Result != nil && queryResp.Data.Result.Greeting == "Hello, world!" {
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatal("timeout waiting example agent result")
}

func registerExampleRoutes(rg *gin.RouterGroup, svc *example.Service) {
	rg.POST("/task/create", func(c *gin.Context) {
		var req example.CreateTaskRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			response.Error(c, apierr.BadRequest("invalid request body"))
			return
		}
		resp, err := svc.CreateTask(c.Request.Context(), req, httpmiddleware.TraceIDFromContext(c))
		if err != nil {
			response.Error(c, err)
			return
		}
		response.Success(c, http.StatusAccepted, resp)
	})

	rg.GET("/task/query", func(c *gin.Context) {
		taskID := c.Query("task_id")
		if taskID == "" {
			response.Error(c, apierr.BadRequest("task_id is required"))
			return
		}
		resp, err := svc.QueryTask(c.Request.Context(), taskID)
		if err != nil {
			response.Error(c, err)
			return
		}
		response.Success(c, http.StatusOK, resp)
	})
}
