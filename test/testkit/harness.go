package testkit

import (
	"io"
	"log/slog"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ZyqAlwaysCool/anvil/internal/platform/config"
	"github.com/ZyqAlwaysCool/anvil/internal/platform/task"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func NewTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func NewTaskManager(t *testing.T) (*task.Manager, *MemoryRepository, *MemoryQueue) {
	t.Helper()
	repo := NewMemoryRepository()
	queue := NewMemoryQueue()
	mgr := task.NewManager(repo, queue, NewTestLogger(), DefaultTaskConfig(), "memory", "memory")
	return mgr, repo, queue
}

func NewTaskRouter(t *testing.T, mgr *task.Manager) *gin.Engine {
	t.Helper()
	engine := gin.New()
	task.RegisterRoutes(engine.Group("/api/v1/tasks"), mgr)
	return engine
}

func PerformJSON(t *testing.T, engine *gin.Engine, method, target string, body io.Reader) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, target, body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	return rec
}

func DefaultTaskConfig() config.TaskConfig {
	return config.TaskConfig{
		StreamKey:         "test-stream",
		ConsumerGroup:     "test-group",
		ConsumerName:      "test-worker",
		ReadBlock:         time.Second,
		MaxLen:            1000,
		WorkerConcurrency: 1,
		RetryJobs:         true,
		MaxTries:          3,
		JobTimeout:        time.Minute,
	}
}
