package mongo

import (
	"context"
	"sync"
	"time"

	"github.com/0Hoag/cryptocheck-api/internal/post/repository"
	"github.com/0Hoag/cryptocheck-api/pkg/log"
	"github.com/0Hoag/cryptocheck-api/pkg/mongo"
	"github.com/0Hoag/cryptocheck-api/pkg/util"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type impleRepository struct {
	l                         log.Logger
	db                        mongo.Database
	clock                     func() time.Time
	reactionIndexOnce         sync.Once
	reactionIndexErr          error
	crawlerRetentionIndexOnce sync.Once
	crawlerRetentionIndexErr  error
}

func (repo *impleRepository) ensureReactionIndex(ctx context.Context) error {
	repo.reactionIndexOnce.Do(func() {
		_, repo.reactionIndexErr = repo.getReactionCollection().CreateIndex(ctx,
			bson.D{{Key: "post_id", Value: 1}, {Key: "author_id", Value: 1}, {Key: "type", Value: 1}},
			options.Index().SetName("reaction_post_author_type_unique").SetUnique(true),
		)
	})
	return repo.reactionIndexErr
}

// ensureCrawlerRetentionIndex keeps the periodic crawler cleanup bounded as
// source-backed posts grow. The query is scoped to one bot author and sorted
// newest-first, so this compound index avoids scanning unrelated social posts.
func (repo *impleRepository) ensureCrawlerRetentionIndex(ctx context.Context) error {
	repo.crawlerRetentionIndexOnce.Do(func() {
		_, repo.crawlerRetentionIndexErr = repo.getPostCollection().CreateIndex(ctx,
			bson.D{{Key: "author_id", Value: 1}, {Key: "created_at", Value: -1}},
			options.Index().SetName("crawler_retention_author_created_at"),
		)
	})
	return repo.crawlerRetentionIndexErr
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
