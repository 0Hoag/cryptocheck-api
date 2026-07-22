package usecase

import (
	"github.com/0Hoag/cryptocheck-api/internal/users"
	"github.com/0Hoag/cryptocheck-api/internal/users/repository"
	"github.com/0Hoag/cryptocheck-api/pkg/log"
)

type impleUsecase struct {
	l    log.Logger
	repo repository.Repository
}

func New(
	l log.Logger,
	repo repository.Repository,
) users.UseCase {
	return &impleUsecase{
		l:    l,
		repo: repo,
	}
}
