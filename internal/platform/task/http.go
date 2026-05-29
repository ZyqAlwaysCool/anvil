package task

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/zyq/anvil/internal/platform/apierr"
	httpmiddleware "github.com/zyq/anvil/internal/platform/http/middleware"
	"github.com/zyq/anvil/internal/platform/http/response"
)

type createTaskRequest struct {
	Agent    string          `json:"agent"`
	TaskType string          `json:"task_type"`
	Payload  json.RawMessage `json:"payload"`
}

type createTaskResponse struct {
	TaskID string `json:"task_id"`
}

type replayTaskRequest struct {
	TaskID string `json:"task_id"`
}

type metadataRequest struct {
	TaskID   string         `json:"task_id"`
	Metadata map[string]any `json:"metadata"`
}

// RegisterRoutes 平台默认任务路由；业务可包一层，但不得破坏 task_id 立即返回契约。
func RegisterRoutes(rg *gin.RouterGroup, mgr *Manager) {
	rg.POST("/create", func(c *gin.Context) {
		var req createTaskRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			response.Error(c, apierr.BadRequest("invalid request body"))
			return
		}
		if req.Agent == "" || req.TaskType == "" {
			response.Error(c, apierr.BadRequest("agent and task_type are required"))
			return
		}

		var payload any
		if len(req.Payload) > 0 {
			payload = json.RawMessage(req.Payload)
		}

		record, err := mgr.Submit(c.Request.Context(), Submission{
			Agent:   req.Agent,
			Type:    req.TaskType,
			Payload: payload,
			TraceID: httpmiddleware.TraceIDFromContext(c),
		})
		if err != nil {
			response.Error(c, err)
			return
		}

		response.Success(c, http.StatusAccepted, createTaskResponse{TaskID: record.ID})
	})

	rg.GET("/query", func(c *gin.Context) {
		taskID := c.Query("task_id")
		if taskID == "" {
			response.Error(c, apierr.BadRequest("task_id is required"))
			return
		}

		record, found, err := mgr.Get(c.Request.Context(), taskID)
		if err != nil {
			response.Error(c, err)
			return
		}
		if !found {
			response.Error(c, apierr.NotFound("task not found"))
			return
		}
		response.Success(c, http.StatusOK, record)
	})

	rg.POST("/replay", func(c *gin.Context) {
		var req replayTaskRequest
		if err := c.ShouldBindJSON(&req); err != nil || req.TaskID == "" {
			response.Error(c, apierr.BadRequest("task_id is required"))
			return
		}

		record, err := mgr.Replay(c.Request.Context(), req.TaskID)
		if err != nil {
			response.Error(c, err)
			return
		}
		response.Success(c, http.StatusAccepted, createTaskResponse{TaskID: record.ID})
	})

	rg.PATCH("/metadata", func(c *gin.Context) {
		var req metadataRequest
		if err := c.ShouldBindJSON(&req); err != nil || req.TaskID == "" {
			response.Error(c, apierr.BadRequest("task_id and metadata are required"))
			return
		}
		if len(req.Metadata) == 0 {
			response.Error(c, apierr.BadRequest("metadata is required"))
			return
		}

		if err := mgr.MergeMetadata(c.Request.Context(), req.TaskID, req.Metadata); err != nil {
			response.Error(c, err)
			return
		}
		response.Success(c, http.StatusOK, gin.H{"task_id": req.TaskID})
	})

	rg.GET("/health", func(c *gin.Context) {
		snapshot, err := mgr.Health(c.Request.Context())
		if err != nil {
			response.Error(c, err)
			return
		}
		response.Success(c, http.StatusOK, snapshot)
	})
}
