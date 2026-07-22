package http

import (
	"github.com/0Hoag/cryptocheck-api/pkg/log"

	"github.com/gin-gonic/gin"
	"github.com/0Hoag/cryptocheck-api/internal/scanner"
)

type Handler interface {
	ScanToken(c *gin.Context)
}

type handler struct {
	uc scanner.UseCase
	l  log.Logger
}

func New(l log.Logger, uc scanner.UseCase) Handler {
	return handler{
		uc: uc,
		l:  l,
	}
}
