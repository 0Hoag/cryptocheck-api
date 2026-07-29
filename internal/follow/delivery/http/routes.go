package http

import (
	"time"

	"github.com/0Hoag/cryptocheck-api/internal/middleware"
	"github.com/gin-gonic/gin"
)

func MapRoutes(r *gin.RouterGroup, h Handler, mw middleware.Middleware) {
	r.GET("/counts/:user_id", h.Counts)

	authenticated := r.Group("")
	authenticated.Use(mw.Auth())
	writes := authenticated.Group("")
	writes.Use(mw.RateLimit(30, time.Minute))
	writes.POST("", h.Create)
	authenticated.GET("/:id", h.Detail)
	authenticated.GET("", h.Get)
	writes.DELETE("/:id", h.Delete)
}
