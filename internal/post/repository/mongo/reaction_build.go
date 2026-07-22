package mongo

import (
	"context"

	"github.com/0Hoag/cryptocheck-api/internal/models"
	"github.com/0Hoag/cryptocheck-api/internal/post/repository"
	"github.com/0Hoag/cryptocheck-api/pkg/mongo"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (repo impleRepository) buildReactionModels(ctx context.Context, sc models.Scope, opts repository.CreateReactionOptions) (models.Reaction, error) {
	now := repo.clock()

	postID, err := primitive.ObjectIDFromHex(opts.PostID)
	if err != nil {
		repo.l.Errorf(ctx, "reaction.repository.buildModels.ObjectIDFromHex: %v", err)
		return models.Reaction{}, err
	}

	reaction := models.Reaction{
		ID:        repo.db.NewObjectID(),
		PostID:    postID,
		AuthorID:  mongo.ObjectIDFromHexOrNil(sc.UserID),
		Type:      opts.Type,
		CreatedAt: now,
	}

	return reaction, nil
}
