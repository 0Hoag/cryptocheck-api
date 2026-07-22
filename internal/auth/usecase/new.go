package usecase

import (
	"github.com/0Hoag/cryptocheck-api/config"
	"github.com/0Hoag/cryptocheck-api/internal/auth"
	"github.com/0Hoag/cryptocheck-api/internal/users"
	"github.com/0Hoag/cryptocheck-api/pkg/log"
)

type impleUsecase struct {
	l      log.Logger
	cfg    *config.Config
	userUC users.UseCase
}

func New(
	l log.Logger,
	cfg *config.Config,
	userUC users.UseCase,
) auth.UseCase {
	return &impleUsecase{
		l:      l,
		cfg:    cfg,
		userUC: userUC,
	}
}
