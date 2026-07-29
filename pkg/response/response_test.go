package response

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestResponsesIncludeRequestID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	for _, tt := range []struct {
		name    string
		respond func(*gin.Context)
	}{
		{name: "success", respond: func(c *gin.Context) { OK(c, map[string]string{"status": "ok"}) }},
		{name: "error", respond: func(c *gin.Context) { Error(c, errors.New("unexpected")) }},
		{name: "unauthorized", respond: Unauthorized},
		{name: "explicit status", respond: func(c *gin.Context) { StatusError(c, http.StatusTooManyRequests, 429, "slow down") }},
	} {
		t.Run(tt.name, func(t *testing.T) {
			writer := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(writer)
			ctx.Request = httptest.NewRequest(http.MethodGet, "/", nil)
			ctx.Request.Header.Set("X-Request-ID", "trace-12345678")
			tt.respond(ctx)

			if got := writer.Body.String(); !strings.Contains(got, `"request_id":"trace-12345678"`) {
				t.Fatalf("response missing request ID: %s", got)
			}
		})
	}
}
