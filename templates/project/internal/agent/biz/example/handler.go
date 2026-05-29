//go:build ignore

package example

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"{{.ModuleName}}/internal/platform/apierr"
	"{{.ModuleName}}/internal/platform/app"
	httpmiddleware "{{.ModuleName}}/internal/platform/http/middleware"
	"{{.ModuleName}}/internal/platform/http/response"
	"{{.ModuleName}}/internal/platform/task"
)

func invalidNameError() *apierr.Error {
	return apierr.New(http.StatusBadRequest, CodeExampleInvalidName, "name is required")
}

func taskNotFoundError() *apierr.Error {
	return apierr.New(http.StatusNotFound, CodeExampleTaskNotFound, "task not found")
}

// registerHTTPRoutes handler 保持薄层：解析参数、调用 service、序列化响应。
func registerHTTPRoutes(rg *gin.RouterGroup, a *app.ServerApp) {
	svc := NewService(a.Task)

	rg.POST("/task/create", func(c *gin.Context) {
		var req CreateTaskRequest
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

func taskHandler() task.Handler {
	return task.HandlerFunc(HandleGreetTask)
}
