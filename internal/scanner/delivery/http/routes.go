package http

import (
	"github.com/0Hoag/cryptocheck-api/internal/middleware"
	"github.com/gin-gonic/gin"
)

func MapRoutes(r *gin.RouterGroup, h Handler, mw middleware.Middleware) {
	r.GET("", h.ScanToken)
	r.GET("/candidates", h.FindCandidates)
}
