package http

import (
	"github.com/0Hoag/cryptocheck-api/internal/follow"
	"github.com/0Hoag/cryptocheck-api/pkg/log"
	"github.com/gin-gonic/gin"
)

type Handler interface {
	Create(c *gin.Context)
	Detail(c *gin.Context)
	Get(c *gin.Context)
	Counts(c *gin.Context)
	Delete(c *gin.Context)
}

type handler struct {
	l  log.Logger
	uc follow.UseCase
}

func New(l log.Logger, uc follow.UseCase) Handler {
	return handler{
		l:  l,
		uc: uc,
	}
}
