package test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	httpmiddleware "github.com/ZyqAlwaysCool/anvil/internal/platform/http/middleware"
)

func TestTraceIDMiddlewarePassThrough(t *testing.T) {
	engine := gin.New()
	engine.Use(httpmiddleware.TraceID())
	engine.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, httpmiddleware.TraceIDFromContext(c))
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	req.Header.Set("X-Trace-ID", "incoming-trace")
	engine.ServeHTTP(rec, req)

	if rec.Header().Get("X-Trace-ID") != "incoming-trace" {
		t.Fatalf("unexpected response trace header: %s", rec.Header().Get("X-Trace-ID"))
	}
	if rec.Body.String() != "incoming-trace" {
		t.Fatalf("unexpected body: %s", rec.Body.String())
	}
}

func TestTraceIDMiddlewareGenerate(t *testing.T) {
	engine := gin.New()
	engine.Use(httpmiddleware.TraceID())
	engine.GET("/ping", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	engine.ServeHTTP(rec, req)

	if rec.Header().Get("X-Trace-ID") == "" {
		t.Fatal("expected generated trace id")
	}
}
