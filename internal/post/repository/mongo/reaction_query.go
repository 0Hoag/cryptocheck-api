package mongo

import (
	"context"

	"github.com/0Hoag/cryptocheck-api/internal/models"
	"github.com/0Hoag/cryptocheck-api/internal/post/repository"
	"github.com/0Hoag/cryptocheck-api/pkg/mongo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (repo impleRepository) buildDetailReactionQuery(ctx context.Context, sc models.Scope, id string) (bson.M, error) {
	filter := bson.M{}
	var err error

	filter = mongo.BuildQueryWithSoftDelete(filter)

	filter["_id"], err = primitive.ObjectIDFromHex(id)
	if err != nil {
		repo.l.Errorf(ctx, "post.mongo.buildDetailReactionQuery.BuildQueryWithSoftDelete: %v", err)
		return bson.M{}, err
	}
	if sc.UserID != "" {
		filter["author_id"], err = primitive.ObjectIDFromHex(sc.UserID)
		if err != nil {
			return bson.M{}, err
		}
	}

	return filter, nil
}

func (repo impleRepository) buildGetOneReactionQuery(ctx context.Context, sc models.Scope, f repository.FilterReaction) (bson.M, error) {
	filter := bson.M{}
	var err error

	filter = mongo.BuildQueryWithSoftDelete(filter)

	if f.ID != "" {
		filter["_id"], err = primitive.ObjectIDFromHex(f.ID)
		if err != nil {
			repo.l.Errorf(ctx, "post.mongo.buildGetOneReactionQuery.BuildQueryWithSoftDelete: %v", err)
			return bson.M{}, err
		}
	}

	if f.UserID != "" {
		filter["user_id"], err = primitive.ObjectIDFromHex(f.UserID)
		if err != nil {
			repo.l.Errorf(ctx, "reaction.mongo.buildGetOneReactionQuery.ObjectIDFromHex: %v", err)
			return bson.M{}, err
		}
	}
	if f.PostID != "" {
		filter["post_id"], err = primitive.ObjectIDFromHex(f.PostID)
		if err != nil {
			return bson.M{}, err
		}
	}

	if f.Type != "" {
		filter["type"] = f.Type
	}

	return filter, nil
}

func (repo impleRepository) buildListReactionQuery(ctx context.Context, sc models.Scope, opts repository.ListReactionOptions) (bson.M, error) {
	filter := bson.M{}
	var err error

	filter = mongo.BuildQueryWithSoftDelete(filter)

	if opts.ID != "" {
		filter["_id"], err = primitive.ObjectIDFromHex(opts.ID)
		if err != nil {
			repo.l.Errorf(ctx, "reaction.mongo.buildListReactionQuery.ObjectIDFromHex: %v", err)
			return bson.M{}, err
		}
	}

	mIDs := make([]primitive.ObjectID, len(opts.IDs))
	if len(opts.IDs) > 0 {
		for _, id := range opts.IDs {
			mID, err := primitive.ObjectIDFromHex(id)
			if err != nil {
				repo.l.Errorf(ctx, "reaction.mongo.buildListReactionQuery.ObjectIDFromHex: %v", err)
				return bson.M{}, err
			}
			mIDs = append(mIDs, mID)
		}
		filter["_id"] = bson.M{"$in": mIDs}
	}

	if opts.UserID != "" {
		filter["user_id"], err = primitive.ObjectIDFromHex(opts.UserID)
		if err != nil {
			repo.l.Errorf(ctx, "reaction.mongo.buildListReactionQuery.ObjectIDFromHex: %v", err)
			return bson.M{}, err
		}
	}
	if opts.PostID != "" {
		filter["post_id"], err = primitive.ObjectIDFromHex(opts.PostID)
		if err != nil {
			return bson.M{}, err
		}
	}

	if opts.Type != "" {
		filter["type"] = opts.Type
	}

	return filter, nil
}

func (repo impleRepository) buildGetReactionQuery(ctx context.Context, sc models.Scope, opts repository.GetReactionOptions) (bson.M, error) {
	filter := bson.M{}
	var err error

	filter = mongo.BuildQueryWithSoftDelete(filter)

	if opts.ID != "" {
		filter["_id"], err = primitive.ObjectIDFromHex(opts.ID)
		if err != nil {
			repo.l.Errorf(ctx, "reaction.mongo.buildGetReactionQuery.ObjectIDFromHex: %v", err)
			return bson.M{}, err
		}
	}

	mIDs := make([]primitive.ObjectID, len(opts.IDs))
	if len(opts.IDs) > 0 {
		for _, id := range opts.IDs {
			mID, err := primitive.ObjectIDFromHex(id)
			if err != nil {
				repo.l.Errorf(ctx, "reaction.mongo.buildGetReactionQuery.ObjectIDFromHex: %v", err)
				return bson.M{}, err
			}
			mIDs = append(mIDs, mID)
		}
		filter["_id"] = bson.M{"$in": mIDs}
	}

	if opts.UserID != "" {
		filter["user_id"], err = primitive.ObjectIDFromHex(opts.UserID)
		if err != nil {
			repo.l.Errorf(ctx, "reaction.mongo.buildGetReactionQuery.ObjectIDFromHex: %v", err)
			return bson.M{}, err
		}
	}
	if opts.PostID != "" {
		filter["post_id"], err = primitive.ObjectIDFromHex(opts.PostID)
		if err != nil {
			return bson.M{}, err
		}
	}

	if opts.Type != "" {
		filter["type"] = opts.Type
	}

	return filter, nil
}
