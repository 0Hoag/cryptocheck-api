package mongo

import (
	"time"

	"github.com/0Hoag/cryptocheck-api/internal/post/repository"
	"github.com/0Hoag/cryptocheck-api/pkg/log"
	"github.com/0Hoag/cryptocheck-api/pkg/mongo"
	"github.com/0Hoag/cryptocheck-api/pkg/util"
)

type impleRepository struct {
	l     log.Logger
	db    mongo.Database
	clock func() time.Time
}

func New(
	l log.Logger,
	db mongo.Database,
) repository.Repository {
	now := util.Now
	return &impleRepository{
		l:     l,
		db:    db,
		clock: now,
	}
}
