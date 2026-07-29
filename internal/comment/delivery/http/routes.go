package http

import (
	"time"

	"github.com/0Hoag/cryptocheck-api/internal/middleware"
	"github.com/gin-gonic/gin"
)

func MapRoutes(r *gin.RouterGroup, h Handler, mw middleware.Middleware) {
	r.Use(mw.Auth())
	writes := r.Group("")
	writes.Use(mw.RateLimit(30, time.Minute))
	writes.POST("", h.Create)
	r.GET("/:id", h.Detail)
	r.GET("", h.Get)
	writes.PUT("", h.Update)
	writes.DELETE("/:id", h.Delete)
}
