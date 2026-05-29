package test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	httpmiddleware "github.com/ZyqAlwaysCool/anvil/internal/platform/http/middleware"
)

type testClaims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

func TestJWTAuthMissingToken(t *testing.T) {
	engine := newJWTTestEngine(t)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/secure", nil)
	engine.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestJWTAuthInvalidToken(t *testing.T) {
	engine := newJWTTestEngine(t)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/secure", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	engine.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestJWTAuthValidToken(t *testing.T) {
	engine := newJWTTestEngine(t)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, testClaims{
		UserID: "user-1",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	})
	signed, err := token.SignedString([]byte("test-secret"))
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/secure", nil)
	req.Header.Set("Authorization", "Bearer "+signed)
	engine.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if rec.Body.String() != "user-1" {
		t.Fatalf("unexpected body: %s", rec.Body.String())
	}
}

func TestClaimsFromContextWrongType(t *testing.T) {
	engine := gin.New()
	engine.GET("/secure", func(c *gin.Context) {
		c.Set("jwt_claims", "not-a-claims-struct")
		_, ok := httpmiddleware.ClaimsFromContext[testClaims](c)
		if ok {
			c.String(http.StatusInternalServerError, "unexpected claims")
			return
		}
		c.String(http.StatusOK, "missing")
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/secure", nil)
	engine.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if rec.Body.String() != "missing" {
		t.Fatalf("unexpected body: %s", rec.Body.String())
	}
}

func newJWTTestEngine(t *testing.T) *gin.Engine {
	t.Helper()
	engine := gin.New()
	engine.Use(httpmiddleware.JWTAuth[testClaims](httpmiddleware.JWTOptions[testClaims]{
		Secret: []byte("test-secret"),
	}))
	engine.GET("/secure", func(c *gin.Context) {
		claims, ok := httpmiddleware.ClaimsFromContext[testClaims](c)
		if !ok {
			c.String(http.StatusInternalServerError, "missing claims")
			return
		}
		c.String(http.StatusOK, claims.UserID)
	})
	return engine
}
