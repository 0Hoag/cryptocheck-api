package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRequestIDPreservesSafeClientValue(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(RequestID())
	router.GET("/", func(c *gin.Context) { c.Status(http.StatusNoContent) })

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(RequestIDHeader, "deploy-2026_07.29")
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	if got := res.Header().Get(RequestIDHeader); got != "deploy-2026_07.29" {
		t.Fatalf("got request ID %q", got)
	}
}

func TestRequestIDReplacesUnsafeClientValue(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(RequestID())
	router.GET("/", func(c *gin.Context) { c.Status(http.StatusNoContent) })

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(RequestIDHeader, "bad value with spaces")
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	if got := res.Header().Get(RequestIDHeader); !validRequestID(got) || got == "bad value with spaces" {
		t.Fatalf("expected generated safe request ID, got %q", got)
	}
}
