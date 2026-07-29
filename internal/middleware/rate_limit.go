package middleware

import (
	"strconv"
	"sync"
	"time"

	pkgErrors "github.com/0Hoag/cryptocheck-api/pkg/errors"
	"github.com/0Hoag/cryptocheck-api/pkg/jwt"
	"github.com/0Hoag/cryptocheck-api/pkg/response"
	"github.com/gin-gonic/gin"
)

type rateLimitEntry struct {
	startedAt time.Time
	count     int
}

// RateLimit limits requests per authenticated account. It is intentionally
// process-local: production should additionally enforce edge/distributed rate
// limits when the API is scaled to multiple instances.
func (m Middleware) RateLimit(limit int, window time.Duration) gin.HandlerFunc {
	var mu sync.Mutex
	entries := make(map[string]rateLimitEntry)

	return func(c *gin.Context) {
		key := c.ClientIP()
		if scope, ok := jwt.GetScopeFromContext(c.Request.Context()); ok && scope.UserID != "" {
			key = "user:" + scope.UserID
		}

		now := time.Now()
		mu.Lock()
		entry := entries[key]
		if entry.startedAt.IsZero() || now.Sub(entry.startedAt) >= window {
			entry = rateLimitEntry{startedAt: now}
		}
		entry.count++
		entries[key] = entry
		exceeded := entry.count > limit
		retryAfter := int((window - now.Sub(entry.startedAt)).Seconds())
		mu.Unlock()

		if exceeded {
			if retryAfter < 1 {
				retryAfter = 1
			}
			c.Header("Retry-After", strconv.Itoa(retryAfter))
			response.Error(c, pkgErrors.NewTooManyRequestsHTTPError())
			c.Abort()
			return
		}

		c.Next()
	}
}
