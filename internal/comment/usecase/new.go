package usecase

import (
	"github.com/0Hoag/cryptocheck-api/internal/comment"
	"github.com/0Hoag/cryptocheck-api/internal/comment/repository"
	"github.com/0Hoag/cryptocheck-api/internal/post"
	"github.com/0Hoag/cryptocheck-api/pkg/log"
)

type impleUsecase struct {
	l      log.Logger
	postUC post.UseCase
	repo   repository.Repository
}

func New(
	l log.Logger,
	postUC post.UseCase,
	repo repository.Repository,
) comment.UseCase {
	return &impleUsecase{
		l:      l,
		postUC: postUC,
		repo:   repo,
	}
}
