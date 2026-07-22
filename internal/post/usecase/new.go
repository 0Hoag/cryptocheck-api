package usecase

import (
	"github.com/0Hoag/cryptocheck-api/internal/delivery/rabbitmq/producer"
	"github.com/0Hoag/cryptocheck-api/internal/post"
	"github.com/0Hoag/cryptocheck-api/internal/post/repository"
	"github.com/0Hoag/cryptocheck-api/internal/users"
	"github.com/0Hoag/cryptocheck-api/pkg/log"
)

type impleUsecase struct {
	l      log.Logger
	prod   producer.Producer
	userUC users.UseCase
	repo   repository.Repository
}

func New(
	l log.Logger,
	prod producer.Producer,
	userUC users.UseCase,
	repo repository.Repository,
) post.UseCase {
	return &impleUsecase{
		l:      l,
		prod:   prod,
		userUC: userUC,
		repo:   repo,
	}
}
