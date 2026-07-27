package usecase

import (
	"github.com/0Hoag/cryptocheck-api/internal/follow"
	"github.com/0Hoag/cryptocheck-api/internal/follow/repository"
	"github.com/0Hoag/cryptocheck-api/internal/users"
	"github.com/0Hoag/cryptocheck-api/pkg/log"
	pkgMongo "github.com/0Hoag/cryptocheck-api/pkg/mongo"
)

type impleUsecase struct {
	l      log.Logger
	userUC users.UseCase
	repo   repository.Repository
	db     pkgMongo.Database
}

func New(
	l log.Logger,
	userUC users.UseCase,
	repo repository.Repository,
	db pkgMongo.Database,
) follow.UseCase {
	return &impleUsecase{
		l:      l,
		userUC: userUC,
		repo:   repo,
		db:     db,
	}
}
