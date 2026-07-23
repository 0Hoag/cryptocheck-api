package http

import (
	"github.com/0Hoag/cryptocheck-api/internal/middleware"
	"github.com/gin-gonic/gin"
)

func MapRoutes(r *gin.RouterGroup, h Handler, mw middleware.Middleware) {
	r.GET("/counts/:user_id", h.Counts)

	authenticated := r.Group("")
	authenticated.Use(mw.Auth())
	authenticated.POST("", h.Create)
	authenticated.GET("/:id", h.Detail)
	authenticated.GET("", h.Get)
	authenticated.DELETE("/:id", h.Delete)
}
