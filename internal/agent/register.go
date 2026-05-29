package agent

import (
	"github.com/ZyqAlwaysCool/anvil/internal/agent/biz/example"
	"github.com/ZyqAlwaysCool/anvil/internal/platform/app"
)

func RegisterRoutes(a *app.ServerApp) error {
	if a.Task == nil {
		return nil
	}
	return example.RegisterRoutes(a.Engine.Group("/api/v1/agents/example"), a)
}

func RegisterTaskHandlers(a *app.WorkerApp) error {
	return example.RegisterTaskHandlers(a)
}
