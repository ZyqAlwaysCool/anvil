//go:build ignore

package middleware

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type traceIDContextKey struct{}

const traceIDHeader = "X-Trace-ID"

// TraceID 从请求头读取或生成 trace id，并写回响应头与 context。
func TraceID() gin.HandlerFunc {
	return func(c *gin.Context) {
		traceID := c.GetHeader(traceIDHeader)
		if traceID == "" {
			traceID = uuid.NewString()
		}

		c.Header(traceIDHeader, traceID)
		c.Set("trace_id", traceID)
		ctx := context.WithValue(c.Request.Context(), traceIDContextKey{}, traceID)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

func TraceIDFromContext(c *gin.Context) string {
	if value, ok := c.Get("trace_id"); ok {
		if traceID, ok := value.(string); ok && traceID != "" {
			return traceID
		}
	}
	if traceID, ok := c.Request.Context().Value(traceIDContextKey{}).(string); ok {
		return traceID
	}
	return ""
}
