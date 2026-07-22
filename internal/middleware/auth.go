package middleware

import (
	"strings"

	"github.com/0Hoag/cryptocheck-api/pkg/jwt"
	"github.com/0Hoag/cryptocheck-api/pkg/response"
	"github.com/gin-gonic/gin"
)

func (m Middleware) Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := strings.TrimSpace(strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer "))
		if tokenString == "" {
			response.Unauthorized(c)
			c.Abort()
			return
		}

		payload, err := m.jwtManager.Verify(tokenString)
		if err != nil {
			response.Unauthorized(c)
			c.Abort()
			return
		}

		ctx := c.Request.Context()
		ctx = jwt.SetPayloadToContext(ctx, payload)

		// Set scope to context
		scope := jwt.NewScope(payload)
		ctx = jwt.SetScopeToContext(ctx, scope)

		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

// OptionalAuth attaches a verified session when a bearer token is present,
// while allowing anonymous requests to continue for public resources.
func (m Middleware) OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := strings.TrimSpace(strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer "))
		if tokenString == "" {
			c.Next()
			return
		}

		payload, err := m.jwtManager.Verify(tokenString)
		if err != nil {
			response.Unauthorized(c)
			c.Abort()
			return
		}

		ctx := c.Request.Context()
		ctx = jwt.SetPayloadToContext(ctx, payload)
		ctx = jwt.SetScopeToContext(ctx, jwt.NewScope(payload))
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
