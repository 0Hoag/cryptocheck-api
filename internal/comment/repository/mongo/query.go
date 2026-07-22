package mongo

import (
	"context"

	"github.com/0Hoag/cryptocheck-api/internal/comment/repository"
	"github.com/0Hoag/cryptocheck-api/internal/models"
	"github.com/0Hoag/cryptocheck-api/pkg/mongo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (repo impleRepository) buildDetailQuery(ctx context.Context, sc models.Scope, id string) (bson.M, error) {
	filter := bson.M{}
	var err error

	filter = mongo.BuildQueryWithSoftDelete(filter)

	filter["_id"], err = primitive.ObjectIDFromHex(id)
	if err != nil {
		repo.l.Errorf(ctx, "comment.mongo.buildDetailQuery.BuildQueryWithSoftDelete: %v", err)
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

func (repo impleRepository) buildGetOneQuery(ctx context.Context, sc models.Scope, f repository.Filter) (bson.M, error) {
	filter := bson.M{}
	var err error

	filter = mongo.BuildQueryWithSoftDelete(filter)

	if f.ID != "" {
		filter["_id"], err = primitive.ObjectIDFromHex(f.ID)
		if err != nil {
			repo.l.Errorf(ctx, "comment.mongo.buildListQuery.ObjectIDFromHex: %v", err)
			return bson.M{}, err
		}
	}

	mIDs := make([]primitive.ObjectID, 0, len(f.IDs))
	if len(f.IDs) > 0 {
		for _, id := range f.IDs {
			mID, err := primitive.ObjectIDFromHex(id)
			if err != nil {
				repo.l.Errorf(ctx, "comment.mongo.buildListQuery.ObjectIDFromHex: %v", err)
				return bson.M{}, err
			}
			mIDs = append(mIDs, mID)
		}
		filter["_id"] = bson.M{"$in": mIDs}
	}

	if f.PostID != "" {
		filter["post_id"], err = primitive.ObjectIDFromHex(f.PostID)
		if err != nil {
			repo.l.Errorf(ctx, "comment.mongo.buildListQuery.ObjectIDFromHex: %v", err)
			return bson.M{}, err
		}
	}

	return filter, nil
}

func (repo impleRepository) buildListQuery(ctx context.Context, sc models.Scope, opts repository.ListOptions) (bson.M, error) {
	filter := bson.M{}
	var err error

	filter = mongo.BuildQueryWithSoftDelete(filter)

	if opts.ID != "" {
		filter["_id"], err = primitive.ObjectIDFromHex(opts.ID)
		if err != nil {
			repo.l.Errorf(ctx, "comment.mongo.buildListQuery.ObjectIDFromHex: %v", err)
			return bson.M{}, err
		}
	}

	mIDs := make([]primitive.ObjectID, 0, len(opts.IDs))
	if len(opts.IDs) > 0 {
		for _, id := range opts.IDs {
			mID, err := primitive.ObjectIDFromHex(id)
			if err != nil {
				repo.l.Errorf(ctx, "comment.mongo.buildListQuery.ObjectIDFromHex: %v", err)
				return bson.M{}, err
			}
			mIDs = append(mIDs, mID)
		}
		filter["_id"] = bson.M{"$in": mIDs}
	}
	if opts.PostID != "" {
		filter["post_id"], err = primitive.ObjectIDFromHex(opts.PostID)
		if err != nil {
			return bson.M{}, err
		}
	}

	return filter, nil
}

func (repo impleRepository) buildGetQuery(ctx context.Context, sc models.Scope, opts repository.GetOptions) (bson.M, error) {
	filter := bson.M{}
	var err error

	filter = mongo.BuildQueryWithSoftDelete(filter)

	if opts.ID != "" {
		filter["_id"], err = primitive.ObjectIDFromHex(opts.ID)
		if err != nil {
			repo.l.Errorf(ctx, "comment.mongo.buildGetQuery.ObjectIDFromHex: %v", err)
			return bson.M{}, err
		}
	}

	mIDs := make([]primitive.ObjectID, 0, len(opts.IDs))
	if len(opts.IDs) > 0 {
		for _, id := range opts.IDs {
			mID, err := primitive.ObjectIDFromHex(id)
			if err != nil {
				repo.l.Errorf(ctx, "comment.mongo.buildGetQuery.ObjectIDFromHex: %v", err)
				return bson.M{}, err
			}
			mIDs = append(mIDs, mID)
		}
		filter["_id"] = bson.M{"$in": mIDs}
	}
	if opts.PostID != "" {
		filter["post_id"], err = primitive.ObjectIDFromHex(opts.PostID)
		if err != nil {
			return bson.M{}, err
		}
	}

	return filter, nil
}
