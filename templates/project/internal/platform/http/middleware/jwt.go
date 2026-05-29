//go:build ignore

package middleware

import (
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"anvil-scaffold-template/internal/platform/apierr"
)

const jwtClaimsKey = "jwt_claims"

type JWTOptions[T any] struct {
	Secret       []byte
	ExcludePaths []string
}

// JWTAuth 仅支持 Authorization: Bearer <token>；校验失败返回 401。
// T 必须是 struct，且 *T 需实现 jwt.Claims（通常嵌入 jwt.RegisteredClaims）。
// 示例：
//
//	type MyClaims struct {
//	    UserID string `json:"user_id"`
//	    jwt.RegisteredClaims
//	}
//	engine.Use(JWTAuth[MyClaims](JWTOptions[MyClaims]{Secret: secret}))
func JWTAuth[T any, PT interface {
	*T
	jwt.Claims
}](opts JWTOptions[T]) gin.HandlerFunc {
	excludes := make(map[string]struct{}, len(opts.ExcludePaths))
	for _, path := range opts.ExcludePaths {
		excludes[path] = struct{}{}
	}

	return func(c *gin.Context) {
		if _, skip := excludes[c.FullPath()]; skip {
			c.Next()
			return
		}

		authHeader := c.GetHeader("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			abortUnauthorized(c, apierr.Unauthorized("missing bearer token"))
			return
		}

		tokenString := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
		claims := PT(new(T))
		if _, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
			return opts.Secret, nil
		}); err != nil {
			abortUnauthorized(c, apierr.Unauthorized("invalid token"))
			return
		}

		c.Set(jwtClaimsKey, *claims)
		c.Next()
	}
}

func ClaimsFromContext[T jwt.Claims](c *gin.Context) (T, bool) {
	var zero T
	value, ok := c.Get(jwtClaimsKey)
	if !ok {
		return zero, false
	}
	claims, ok := value.(T)
	return claims, ok
}

func abortUnauthorized(c *gin.Context, err *apierr.Error) {
	c.AbortWithStatusJSON(err.StatusCode(), gin.H{
		"code":      err.Code(),
		"msg":       err.Message(),
		"trace_id":  TraceIDFromContext(c),
		"timestamp": time.Now().Unix(),
	})
}
