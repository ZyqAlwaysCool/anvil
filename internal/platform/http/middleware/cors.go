package middleware

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type CORSOptions struct {
	AllowOrigins []string
	AllowMethods []string
	AllowHeaders []string
	MaxAge       int
}

// CORS 使用白名单策略；AllowOrigins 为空时不放开任意来源，仅允许同源请求。
func CORS(opts CORSOptions) gin.HandlerFunc {
	methods := opts.AllowMethods
	if len(methods) == 0 {
		methods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}
	}
	headers := opts.AllowHeaders
	if len(headers) == 0 {
		headers = []string{"Content-Type", "Authorization", traceIDHeader}
	}

	originSet := make(map[string]struct{}, len(opts.AllowOrigins))
	for _, origin := range opts.AllowOrigins {
		origin = strings.TrimSpace(origin)
		if origin != "" {
			originSet[origin] = struct{}{}
		}
	}

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin != "" {
			if _, ok := originSet[origin]; ok {
				c.Header("Access-Control-Allow-Origin", origin)
				c.Header("Vary", "Origin")
			}
		}

		if c.Request.Method == http.MethodOptions {
			c.Header("Access-Control-Allow-Methods", strings.Join(methods, ","))
			c.Header("Access-Control-Allow-Headers", strings.Join(headers, ","))
			if opts.MaxAge > 0 {
				c.Header("Access-Control-Max-Age", strconv.Itoa(opts.MaxAge))
			}
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
