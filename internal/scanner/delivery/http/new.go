package http

import (
	"github.com/0Hoag/cryptocheck-api/pkg/log"
	pkgMongo "github.com/0Hoag/cryptocheck-api/pkg/mongo"

	"github.com/0Hoag/cryptocheck-api/internal/scanner"
	"github.com/gin-gonic/gin"
)

type Handler interface {
	ScanToken(c *gin.Context)
	FindCandidates(c *gin.Context)
	History(c *gin.Context)
}

type handler struct {
	uc scanner.UseCase
	l  log.Logger
	db pkgMongo.Database
}

func New(l log.Logger, uc scanner.UseCase, db pkgMongo.Database) Handler {
	return handler{
		uc: uc,
		l:  l,
		db: db,
	}
}
