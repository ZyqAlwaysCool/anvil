package test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/zyq/anvil/internal/platform/apierr"
	httpmiddleware "github.com/zyq/anvil/internal/platform/http/middleware"
	"github.com/zyq/anvil/internal/platform/http/response"
)

func TestResponseSuccessContract(t *testing.T) {
	engine := gin.New()
	engine.Use(httpmiddleware.TraceID())
	engine.GET("/ok", func(c *gin.Context) {
		response.Success(c, http.StatusOK, gin.H{"value": 1})
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ok", nil)
	req.Header.Set("X-Trace-ID", "trace-test")
	engine.ServeHTTP(rec, req)

	var body response.BaseResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if body.Code != apierr.CodeSuccess {
		t.Fatalf("unexpected code: %d", body.Code)
	}
	if body.TraceID != "trace-test" {
		t.Fatalf("unexpected trace_id: %s", body.TraceID)
	}
	if body.Timestamp == 0 {
		t.Fatal("timestamp should not be zero")
	}
}

func TestResponseErrorContract(t *testing.T) {
	engine := gin.New()
	engine.Use(httpmiddleware.TraceID())
	engine.GET("/fail", func(c *gin.Context) {
		response.Error(c, apierr.BadRequest("bad input"))
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/fail", nil)
	engine.ServeHTTP(rec, req)

	var body response.BaseResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if body.Code != apierr.CodeBadRequest {
		t.Fatalf("unexpected code: %d", body.Code)
	}
	if body.TraceID == "" {
		t.Fatal("trace_id should exist")
	}
	if body.Timestamp == 0 {
		t.Fatal("timestamp should not be zero")
	}
}
