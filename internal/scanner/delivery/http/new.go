package http

import (
	"github.com/0Hoag/cryptocheck-api/pkg/log"

	"github.com/0Hoag/cryptocheck-api/internal/scanner"
	"github.com/gin-gonic/gin"
)

type Handler interface {
	ScanToken(c *gin.Context)
	FindCandidates(c *gin.Context)
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
