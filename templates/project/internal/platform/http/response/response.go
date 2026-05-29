//go:build ignore

package response

import (
	"time"

	"github.com/gin-gonic/gin"

	"{{.ModuleName}}/internal/platform/apierr"
	httpmiddleware "{{.ModuleName}}/internal/platform/http/middleware"
)

type BaseResponse struct {
	Code      int    `json:"code"`
	Msg       string `json:"msg"`
	Data      any    `json:"data,omitempty"`
	TraceID   string `json:"trace_id,omitempty"`
	Timestamp int64  `json:"timestamp"`
}

// Success 统一成功响应；trace_id 与 timestamp 为审计字段，必须始终写入。
func Success(c *gin.Context, status int, data any) {
	c.JSON(status, BaseResponse{
		Code:      apierr.CodeSuccess,
		Msg:       "ok",
		Data:      data,
		TraceID:   httpmiddleware.TraceIDFromContext(c),
		Timestamp: time.Now().Unix(),
	})
}

// Error 识别 apierr.Error，其余错误统一映射为 CodeInternal。
func Error(c *gin.Context, err error) {
	appErr := apierr.From(err)
	if appErr == nil {
		return
	}
	_ = c.Error(err)
	c.AbortWithStatusJSON(appErr.StatusCode(), BaseResponse{
		Code:      appErr.Code(),
		Msg:       appErr.Message(),
		Data:      appErr.Data(),
		TraceID:   httpmiddleware.TraceIDFromContext(c),
		Timestamp: time.Now().Unix(),
	})
}
