package mongo

import (
	"context"

	"github.com/0Hoag/cryptocheck-api/internal/models"
	"github.com/0Hoag/cryptocheck-api/internal/post/repository"
	"github.com/0Hoag/cryptocheck-api/pkg/mongo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const followCollection = "follow"

func (repo impleRepository) buildDetailQuery(ctx context.Context, sc models.Scope, id string) (bson.M, error) {
	filter := bson.M{}

	filter = mongo.BuildQueryWithSoftDelete(filter)

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		repo.l.Errorf(ctx, "post.mongo.buildDetailQuery.ObjectIDFromHex: %v", err)
		return bson.M{}, err
	}

	filter["_id"] = objectID
	if err := repo.applyPostVisibility(ctx, filter, sc); err != nil {
		return bson.M{}, err
	}

	return filter, nil
}

func (repo impleRepository) buildListQuery(ctx context.Context, sc models.Scope, opts repository.ListOptions) (bson.M, error) {
	filter := bson.M{}

	filter = mongo.BuildQueryWithSoftDelete(filter)

	if opts.ID != "" {
		objectID, err := primitive.ObjectIDFromHex(opts.ID)
		if err != nil {
			repo.l.Errorf(ctx, "post.mongo.buildListQuery.ObjectIDFromHex: %v", err)
			return bson.M{}, err
		}
		filter["_id"] = objectID
	}

	mIDs := make([]primitive.ObjectID, 0, len(opts.IDs))
	if len(opts.IDs) > 0 {
		for _, id := range opts.IDs {
			mID, err := primitive.ObjectIDFromHex(id)
			if err != nil {
				repo.l.Errorf(ctx, "post.mongo.buildListQuery.ObjectIDFromHex: %v", err)
				return bson.M{}, err
			}
			mIDs = append(mIDs, mID)
		}
		filter["_id"] = bson.M{"$in": mIDs}
	}

	if opts.AuthorID != "" {
		objectID, err := primitive.ObjectIDFromHex(opts.AuthorID)
		if err != nil {
			repo.l.Errorf(ctx, "post.mongo.buildListQuery.ObjectIDFromHex: %v", err)
			return bson.M{}, err
		}
		filter["author_id"] = objectID
	}

	filter["pin"] = opts.Pin
	if err := repo.applyPostVisibility(ctx, filter, sc); err != nil {
		return bson.M{}, err
	}

	return filter, nil
}

func (repo impleRepository) buildGetQuery(ctx context.Context, sc models.Scope, opts repository.GetOptions) (bson.M, error) {
	filter := bson.M{}

	filter = mongo.BuildQueryWithSoftDelete(filter)

	if opts.ID != "" {
		objectID, err := primitive.ObjectIDFromHex(opts.ID)
		if err != nil {
			repo.l.Errorf(ctx, "post.mongo.buildListQuery.ObjectIDFromHex: %v", err)
			return bson.M{}, err
		}
		filter["_id"] = objectID
	}

	mIDs := make([]primitive.ObjectID, 0, len(opts.IDs))
	if len(opts.IDs) > 0 {
		for _, id := range opts.IDs {
			mID, err := primitive.ObjectIDFromHex(id)
			if err != nil {
				repo.l.Errorf(ctx, "post.mongo.buildListQuery.ObjectIDFromHex: %v", err)
				return bson.M{}, err
			}
			mIDs = append(mIDs, mID)
		}
		filter["_id"] = bson.M{"$in": mIDs}
	}

	if opts.AuthorID != "" {
		objectID, err := primitive.ObjectIDFromHex(opts.AuthorID)
		if err != nil {
			repo.l.Errorf(ctx, "post.mongo.buildListQuery.ObjectIDFromHex: %v", err)
			return bson.M{}, err
		}
		filter["author_id"] = objectID
	}

	filter["pin"] = opts.Pin
	if err := repo.applyPostVisibility(ctx, filter, sc); err != nil {
		return bson.M{}, err
	}

	return filter, nil
}

func (repo impleRepository) applyPostVisibility(ctx context.Context, filter bson.M, sc models.Scope) error {
	publicFilter := bson.M{"permission": bson.M{"$in": bson.A{models.PrivacyTypePublic, ""}}}
	if sc.UserID == "" {
		for key, value := range publicFilter {
			filter[key] = value
		}
		return nil
	}

	authorID, err := primitive.ObjectIDFromHex(sc.UserID)
	if err != nil {
		return err
	}

	followeeIDs, err := repo.followeeIDs(ctx, authorID)
	if err != nil {
		return err
	}

	filter["$or"] = visiblePostFilters(publicFilter, authorID, followeeIDs)
	return nil
}

func visiblePostFilters(publicFilter bson.M, viewerID primitive.ObjectID, followeeIDs []primitive.ObjectID) bson.A {
	visibleFilters := bson.A{publicFilter, bson.M{"author_id": viewerID}}
	if len(followeeIDs) > 0 {
		visibleFilters = append(visibleFilters, bson.M{
			"permission": models.PrivacyTypeFollowers,
			"author_id":  bson.M{"$in": followeeIDs},
		})
	}
	return visibleFilters
}

func (repo impleRepository) followeeIDs(ctx context.Context, followerID primitive.ObjectID) ([]primitive.ObjectID, error) {
	filter := mongo.BuildQueryWithSoftDelete(bson.M{"author_id": followerID})
	cur, err := repo.db.Collection(followCollection).Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var follows []models.Follow
	if err := cur.All(ctx, &follows); err != nil {
		return nil, err
	}

	followeeIDs := make([]primitive.ObjectID, 0, len(follows))
	for _, follow := range follows {
		followeeIDs = append(followeeIDs, follow.FolloweeID)
	}
	return followeeIDs, nil
}

func (repo impleRepository) buildGetOneQuery(ctx context.Context, sc models.Scope, opts repository.GetOneOptions) (bson.M, error) {
	filter := bson.M{}

	filter = mongo.BuildQueryWithSoftDelete(filter)

	if opts.ID != "" {
		objectID, err := primitive.ObjectIDFromHex(opts.ID)
		if err != nil {
			repo.l.Errorf(ctx, "post.mongo.buildGetOneQuery.ObjectIDFromHex: %v", err)
			return bson.M{}, err
		}
		filter["_id"] = objectID
		if err != nil {
			repo.l.Errorf(ctx, "post.mongo.buildGetOneQuery.ObjectIDFromHex: %v", err)
			return bson.M{}, err
		}
	}

	if opts.AuthorID != "" {
		objectID, err := primitive.ObjectIDFromHex(opts.AuthorID)
		if err != nil {
			repo.l.Errorf(ctx, "post.mongo.buildGetOneQuery.ObjectIDFromHex: %v", err)
			return bson.M{}, err
		}
		filter["author_id"] = objectID
	}

	if opts.SourceURL != "" {
		filter["source_url"] = opts.SourceURL
	}

	return filter, nil
}
