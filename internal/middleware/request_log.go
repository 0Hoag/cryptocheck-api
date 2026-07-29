package middleware

import (
	"time"

	pkgLog "github.com/0Hoag/cryptocheck-api/pkg/log"
	"github.com/gin-gonic/gin"
)

// RequestLog emits one completion record per request. The stable request ID is
// included so a user-facing error can be correlated with an API log entry
// without logging request bodies, credentials, or query parameters.
func RequestLog(logger pkgLog.Logger) gin.HandlerFunc {
	if logger == nil {
		return func(c *gin.Context) { c.Next() }
	}
	return func(c *gin.Context) {
		started := time.Now()
		c.Next()
		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}
		requestID, _ := c.Get(RequestIDHeader)
		logger.Infof(c.Request.Context(), "http_request method=%s path=%s status=%d latency_ms=%d request_id=%v", c.Request.Method, path, c.Writer.Status(), time.Since(started).Milliseconds(), requestID)
	}
}
