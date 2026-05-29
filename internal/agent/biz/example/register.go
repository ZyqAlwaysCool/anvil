package example

import (
	"github.com/gin-gonic/gin"
	"github.com/ZyqAlwaysCool/anvil/internal/platform/app"
)

// RegisterRoutes 注册示例 Agent HTTP 路由；要求 Server 已装配 Task Manager。
func RegisterRoutes(rg *gin.RouterGroup, a *app.ServerApp) error {
	registerHTTPRoutes(rg, a)
	return nil
}

// RegisterTaskHandlers 注册 Worker 侧任务处理器。
func RegisterTaskHandlers(a *app.WorkerApp) error {
	return a.Registry.Register(TaskType, taskHandler())
}
