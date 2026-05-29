//go:build ignore

package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

// AccessLog 只记录 method/path/status/latency/client_ip/trace_id，禁止记录请求体与响应体。
func AccessLog(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		startedAt := time.Now()
		c.Next()

		logger.Info("http access",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"latency_ms", time.Since(startedAt).Milliseconds(),
			"client_ip", c.ClientIP(),
			"trace_id", TraceIDFromContext(c),
		)
	}
}
