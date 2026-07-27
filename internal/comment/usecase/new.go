package usecase

import (
	"github.com/0Hoag/cryptocheck-api/internal/comment"
	"github.com/0Hoag/cryptocheck-api/internal/comment/repository"
	"github.com/0Hoag/cryptocheck-api/internal/post"
	"github.com/0Hoag/cryptocheck-api/pkg/log"
	pkgMongo "github.com/0Hoag/cryptocheck-api/pkg/mongo"
)

type impleUsecase struct {
	l      log.Logger
	postUC post.UseCase
	repo   repository.Repository
	db     pkgMongo.Database
}

func New(
	l log.Logger,
	postUC post.UseCase,
	repo repository.Repository,
	db pkgMongo.Database,
) comment.UseCase {
	return &impleUsecase{
		l:      l,
		postUC: postUC,
		repo:   repo,
		db:     db,
	}
}
