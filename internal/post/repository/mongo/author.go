package mongo

import (
	"context"

	"github.com/0Hoag/cryptocheck-api/internal/models"
	"github.com/0Hoag/cryptocheck-api/internal/post/repository"
	"github.com/0Hoag/cryptocheck-api/pkg/mongo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const socialUserCollection = "social_users"

func (repo impleRepository) GetAuthorSummaries(ctx context.Context, authorIDs []primitive.ObjectID) (map[primitive.ObjectID]repository.AuthorSummary, error) {
	summaries := make(map[primitive.ObjectID]repository.AuthorSummary, len(authorIDs))
	if len(authorIDs) == 0 {
		return summaries, nil
	}

	cur, err := repo.db.Collection(socialUserCollection).Find(ctx, mongo.BuildQueryWithSoftDelete(bson.M{"_id": bson.M{"$in": authorIDs}}))
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var users []models.User
	if err := cur.All(ctx, &users); err != nil {
		return nil, err
	}
	for _, user := range users {
		summaries[user.ID] = repository.AuthorSummary{Username: user.Username, AvatarURL: user.AvatarURL}
	}
	return summaries, nil
}
