package test

import (
	"context"
	"os"
	"testing"

	"github.com/ZyqAlwaysCool/anvil/internal/platform/app"
	"github.com/ZyqAlwaysCool/anvil/internal/platform/config"
)

func TestConfigLoadMissingRequired(t *testing.T) {
	t.Setenv("APP_NAME", "")
	t.Setenv("HTTP_PORT", "")
	_, err := config.Load()
	if err == nil {
		t.Fatal("expected error for missing required config")
	}
}

func TestConfigLoadLLMRequiredWhenEnabled(t *testing.T) {
	setValidBaseConfig(t)
	t.Setenv("LLM_ENABLED", "true")
	t.Setenv("LLM_API_KEY", "")
	t.Setenv("LLM_MODEL", "")
	_, err := config.Load()
	if err == nil {
		t.Fatal("expected llm validation error")
	}
}

func TestConfigLoadRedisRequiredWhenEnabled(t *testing.T) {
	setValidBaseConfig(t)
	t.Setenv("REDIS_ENABLED", "true")
	t.Setenv("REDIS_ADDRS", "")
	_, err := config.Load()
	if err == nil {
		t.Fatal("expected redis validation error")
	}
}

func TestConfigLoadMongoRequiredWhenEnabled(t *testing.T) {
	setValidBaseConfig(t)
	t.Setenv("MONGO_ENABLED", "true")
	t.Setenv("MONGO_URI", "")
	_, err := config.Load()
	if err == nil {
		t.Fatal("expected mongo validation error")
	}
}

func TestConfigLoadTaskRequiredWhenEnabled(t *testing.T) {
	setValidBaseConfig(t)
	t.Setenv("TASK_ENABLED", "true")
	t.Setenv("TASK_STREAM_KEY", "")
	_, err := config.Load()
	if err == nil {
		t.Fatal("expected task validation error")
	}
}

func TestServerStartsWithoutTaskAndLLM(t *testing.T) {
	setValidBaseConfig(t)
	t.Setenv("TASK_ENABLED", "false")
	t.Setenv("LLM_ENABLED", "false")

	serverApp, err := app.NewServer(context.Background(), app.ServerOptions{
		RegisterRoutes: func(a *app.ServerApp) error {
			if a.Task != nil {
				t.Fatal("task manager should be nil when task disabled")
			}
			if a.LLM != nil {
				t.Fatal("llm client should be nil when llm disabled")
			}
			return nil
		},
	})
	if err != nil {
		t.Fatalf("new server: %v", err)
	}
	defer func() {
		_ = serverApp.Close(context.Background())
	}()
}

func setValidBaseConfig(t *testing.T) {
	t.Helper()
	t.Setenv("APP_NAME", "test-app")
	t.Setenv("HTTP_PORT", "8080")
	t.Setenv("LOG_DIR", t.TempDir())
	t.Setenv("TASK_ENABLED", "false")
	t.Setenv("LLM_ENABLED", "false")
	t.Setenv("REDIS_ENABLED", "false")
	t.Setenv("MONGO_ENABLED", "false")
	_ = os.Unsetenv(".env")
}
