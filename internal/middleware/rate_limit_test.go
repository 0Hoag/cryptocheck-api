package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/0Hoag/cryptocheck-api/internal/models"
	"github.com/0Hoag/cryptocheck-api/pkg/jwt"
	"github.com/gin-gonic/gin"
)

func TestRateLimitUsesAuthenticatedUserAndReturnsRetryAfter(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Request = c.Request.WithContext(jwt.SetScopeToContext(c.Request.Context(), models.Scope{UserID: "user-a"}))
		c.Next()
	})
	router.Use((Middleware{}).RateLimit(2, time.Minute))
	router.GET("/", func(c *gin.Context) { c.Status(http.StatusNoContent) })

	for i := 0; i < 2; i++ {
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))
		if recorder.Code != http.StatusNoContent {
			t.Fatalf("request %d: got status %d, want %d", i+1, recorder.Code, http.StatusNoContent)
		}
	}

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))
	if recorder.Code != http.StatusTooManyRequests {
		t.Fatalf("got status %d, want %d", recorder.Code, http.StatusTooManyRequests)
	}
	if recorder.Header().Get("Retry-After") == "" {
		t.Fatal("expected Retry-After header")
	}
}
