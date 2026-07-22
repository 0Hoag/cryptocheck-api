package http

import (
	"github.com/gin-gonic/gin"
	"github.com/0Hoag/cryptocheck-api/internal/auth"
	"github.com/0Hoag/cryptocheck-api/pkg/log"
)

type Handler interface {
	Login(c *gin.Context)
}

type handler struct {
	l  log.Logger
	uc auth.UseCase
}

func New(l log.Logger, uc auth.UseCase) Handler {
	return handler{
		l:  l,
		uc: uc,
	}
}
