package http

import (
	"time"

	"github.com/0Hoag/cryptocheck-api/internal/middleware"
	"github.com/gin-gonic/gin"
)

func MapRoutes(r *gin.RouterGroup, h Handler, mw middleware.Middleware) {
	// Public routes (no auth required)
	r.GET("/:id", mw.OptionalAuth(), h.Detail)
	r.GET("", mw.OptionalAuth(), h.Get)

	// Protected routes (auth required)
	authenticated := r.Group("")
	authenticated.Use(mw.Auth(), mw.RateLimit(30, time.Minute))
	authenticated.POST("", h.Create)
	authenticated.PUT("", h.Update)
	authenticated.DELETE("/:id", h.Delete)

	mapReactionRoutes(r, h, mw)
}

func mapReactionRoutes(r *gin.RouterGroup, h Handler, mw middleware.Middleware) {
	reaction := r.Group("/reaction")
	reaction.Use(mw.Auth())
	reactionWrites := reaction.Group("")
	reactionWrites.Use(mw.RateLimit(30, time.Minute))
	reactionWrites.POST("", h.CreateReaction)
	reaction.GET("/:id", h.DetailReaction)
	reaction.GET("", h.GetReaction)
	reactionWrites.DELETE("/:id", h.DeleteReaction)
}
