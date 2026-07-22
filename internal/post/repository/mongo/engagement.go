package mongo

import (
	"context"

	"github.com/0Hoag/cryptocheck-api/internal/post/repository"
	pkgMongo "github.com/0Hoag/cryptocheck-api/pkg/mongo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type engagementAggregate struct {
	PostID primitive.ObjectID `bson:"_id"`
	Count  int64              `bson:"count"`
}

func (repo impleRepository) GetEngagementCounts(ctx context.Context, postIDs []primitive.ObjectID) (map[primitive.ObjectID]repository.EngagementCounts, error) {
	counts := make(map[primitive.ObjectID]repository.EngagementCounts, len(postIDs))
	if len(postIDs) == 0 {
		return counts, nil
	}

	reactionCounts, err := repo.aggregateCounts(ctx, repo.getReactionCollection(), postIDs)
	if err != nil {
		return nil, err
	}
	commentCounts, err := repo.aggregateCounts(ctx, repo.getCommentCollection(), postIDs)
	if err != nil {
		return nil, err
	}

	for _, result := range reactionCounts {
		count := counts[result.PostID]
		count.ReactionCount = result.Count
		counts[result.PostID] = count
	}
	for _, result := range commentCounts {
		count := counts[result.PostID]
		count.CommentCount = result.Count
		counts[result.PostID] = count
	}

	return counts, nil
}

func (repo impleRepository) aggregateCounts(ctx context.Context, collection pkgMongo.Collection, postIDs []primitive.ObjectID) ([]engagementAggregate, error) {
	pipeline := bson.A{
		bson.M{"$match": bson.M{"post_id": bson.M{"$in": postIDs}, "deleted_at": nil}},
		bson.M{"$group": bson.M{"_id": "$post_id", "count": bson.M{"$sum": 1}}},
	}
	cur, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var results []engagementAggregate
	if err := cur.All(ctx, &results); err != nil {
		return nil, err
	}
	return results, nil
}
