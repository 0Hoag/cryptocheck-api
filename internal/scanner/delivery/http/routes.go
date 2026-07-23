package http

import (
	"github.com/0Hoag/cryptocheck-api/internal/middleware"
	"github.com/gin-gonic/gin"
)

func MapRoutes(r *gin.RouterGroup, h Handler, mw middleware.Middleware) {
	r.GET("", mw.OptionalAuth(), h.ScanToken)
	r.GET("/candidates", h.FindCandidates)

	authenticated := r.Group("")
	authenticated.Use(mw.Auth())
	authenticated.GET("/history", h.History)
}
