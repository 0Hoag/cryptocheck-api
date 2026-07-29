package middleware

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

type recordedLogger struct{ messages []string }

func (l *recordedLogger) Debug(context.Context, ...any)          {}
func (l *recordedLogger) Debugf(context.Context, string, ...any) {}
func (l *recordedLogger) Info(context.Context, ...any)           {}
func (l *recordedLogger) Warn(context.Context, ...any)           {}
func (l *recordedLogger) Warnf(context.Context, string, ...any)  {}
func (l *recordedLogger) Error(context.Context, ...any)          {}
func (l *recordedLogger) Errorf(context.Context, string, ...any) {}
func (l *recordedLogger) Fatal(context.Context, ...any)          {}
func (l *recordedLogger) Fatalf(context.Context, string, ...any) {}
func (l *recordedLogger) Infof(_ context.Context, template string, args ...any) {
	l.messages = append(l.messages, fmt.Sprintf(template, args...))
}

func TestRequestLogIncludesRequestIDAndRoute(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := &recordedLogger{}
	router := gin.New()
	router.Use(RequestID(), RequestLog(logger))
	router.GET("/healthz", func(c *gin.Context) { c.Status(http.StatusNoContent) })
	request := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	request.Header.Set(RequestIDHeader, "trace-1234")
	router.ServeHTTP(httptest.NewRecorder(), request)
	if len(logger.messages) != 1 || !strings.Contains(logger.messages[0], "request_id=trace-1234") || !strings.Contains(logger.messages[0], "path=/healthz") {
		t.Fatalf("unexpected request log: %#v", logger.messages)
	}
}
