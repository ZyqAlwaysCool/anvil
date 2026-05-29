package server

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/ZyqAlwaysCool/anvil/internal/agent"
	"github.com/ZyqAlwaysCool/anvil/internal/platform/app"
	"github.com/ZyqAlwaysCool/anvil/internal/platform/http/response"
	"github.com/ZyqAlwaysCool/anvil/internal/platform/task"
)

// RegisterRoutes 只做路由聚合，不写业务规则。
func RegisterRoutes(a *app.ServerApp) error {
	a.Engine.GET("/health", func(c *gin.Context) {
		response.Success(c, http.StatusOK, gin.H{"status": "ok"})
	})

	// 任务路由仅在 Task Manager 装配成功时注册，避免未启用任务系统时暴露无效接口。
	if a.Task != nil {
		task.RegisterRoutes(a.Engine.Group("/api/v1/tasks"), a.Task)
	}
	return agent.RegisterRoutes(a)
}
