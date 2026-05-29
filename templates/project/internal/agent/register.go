//go:build ignore

package agent

import (
	"anvil-scaffold-template/internal/agent/biz/example"
	"anvil-scaffold-template/internal/platform/app"
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
