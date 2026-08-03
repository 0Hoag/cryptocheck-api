package mongo

import (
	"context"
	"fmt"

	"github.com/0Hoag/cryptocheck-api/internal/models"
	"github.com/0Hoag/cryptocheck-api/internal/post/repository"
	"github.com/0Hoag/cryptocheck-api/pkg/mongo"
	"github.com/0Hoag/cryptocheck-api/pkg/paginator"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	driverMongo "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	postCollection = "social_post"
)

func (repo impleRepository) getPostCollection() mongo.Collection {
	collName := fmt.Sprintf("%s", postCollection)
	return repo.db.Collection(collName)
}

func (repo impleRepository) Create(ctx context.Context, sc models.Scope, opts repository.CreateOptions) (models.Post, error) {
	col := repo.getPostCollection()

	m, err := repo.buildModels(ctx, sc, opts)
	if err != nil {
		repo.l.Errorf(ctx, "post.mongo.Create.buildModels: %v", err)
		return models.Post{}, err
	}

	_, err = col.InsertOne(ctx, m)
	if err != nil {
		repo.l.Errorf(ctx, "post.mogno.Create.InsertOne: %v", err)
		return models.Post{}, err
	}

	return m, nil
}

func (repo impleRepository) Detail(ctx context.Context, sc models.Scope, id string) (models.Post, error) {
	col := repo.getPostCollection()

	filter, err := repo.buildDetailQuery(ctx, sc, id)
	if err != nil {
		repo.l.Errorf(ctx, "post.mongo.Detail.buildDetailQuery: %v", err)
		return models.Post{}, err
	}

	var m models.Post
	err = col.FindOne(ctx, filter).Decode(&m)
	if err != nil {
		repo.l.Errorf(ctx, "post.mongo.Detail.FindOne: %v", err)
		return models.Post{}, err
	}

	return m, nil
}

func (repo impleRepository) List(ctx context.Context, sc models.Scope, opts repository.ListOptions) ([]models.Post, error) {
	col := repo.getPostCollection()

	filter, err := repo.buildListQuery(ctx, sc, opts)
	if err != nil {
		repo.l.Errorf(ctx, "post.mongo.List.buildListQuery: %v", err)
		return []models.Post{}, err
	}

	cur, err := col.Find(ctx, filter)
	if err != nil {
		repo.l.Errorf(ctx, "post.mongo.List.buildListQuery: %v", err)
		return []models.Post{}, err
	}

	var ms []models.Post
	err = cur.All(ctx, ms)
	if err != nil {
		repo.l.Errorf(ctx, "post.mongo.List.All: %v", err)
		return []models.Post{}, err
	}

	return ms, nil
}

func (repo *impleRepository) Get(ctx context.Context, sc models.Scope, opts repository.GetOptions) ([]models.Post, paginator.Paginator, error) {
	col := repo.getPostCollection()
	if err := repo.ensureFeedIndexes(ctx); err != nil {
		return []models.Post{}, paginator.Paginator{}, fmt.Errorf("ensure post feed indexes: %w", err)
	}

	filter, err := repo.buildGetQuery(ctx, sc, opts)
	if err != nil {
		repo.l.Errorf(ctx, "post.mongo.Get.buildGetQuery: %v", err)
		return []models.Post{}, paginator.Paginator{}, err
	}

	cur, err := col.Find(ctx, filter, options.Find().
		SetLimit(opts.PagQuery.Limit).
		SetSkip(opts.PagQuery.Offset()).
		SetSort(bson.D{{"created_at", -1}})) // Newest first
	if err != nil {
		repo.l.Errorf(ctx, "post.mongo.Get.Find: %v", err)
		return []models.Post{}, paginator.Paginator{}, err
	}

	var ms []models.Post
	err = cur.All(ctx, &ms)
	if err != nil {
		repo.l.Errorf(ctx, "post.mongo.Get.All: %v", err)
		return []models.Post{}, paginator.Paginator{}, err
	}

	total, err := col.CountDocuments(ctx, filter)
	if err != nil {
		repo.l.Errorf(ctx, "post.mongo.Get.CountDocuments: %v", err)
		return []models.Post{}, paginator.Paginator{}, err
	}

	return ms, paginator.Paginator{
		Total:       total,
		Count:       int64(len(ms)),
		PerPage:     opts.PagQuery.Limit,
		CurrentPage: opts.PagQuery.Page,
	}, nil
}

func (repo impleRepository) Update(ctx context.Context, sc models.Scope, opts repository.UpdateOptions) error {
	col := repo.getPostCollection()

	filter, err := repo.buildDetailQuery(ctx, sc, opts.Post.ID.Hex())
	if err != nil {
		repo.l.Errorf(ctx, "post.mongo.Update.buildUpdateModels: %v", err)
		return err
	}

	update, err := repo.buildUpdateModels(ctx, opts)
	if err != nil {
		repo.l.Errorf(ctx, "post.mongo.Update.buildUpdateModels: %v", err)
		return err
	}

	_, err = col.UpdateOne(ctx, filter, update)
	if err != nil {
		repo.l.Errorf(ctx, "post.mongo.Update.UpdateOne: %v", err)
		return err
	}

	return nil
}

func (repo impleRepository) GetOne(ctx context.Context, sc models.Scope, opts repository.GetOneOptions) (models.Post, error) {
	col := repo.getPostCollection()

	filter, err := repo.buildGetOneQuery(ctx, sc, opts)
	if err != nil {
		repo.l.Errorf(ctx, "post.mongo.GetOne.buildGetOneQuery: %v", err)
		return models.Post{}, err
	}

	var m models.Post
	err = col.FindOne(ctx, filter).Decode(&m)
	if err != nil {
		if err != driverMongo.ErrNoDocuments {
			repo.l.Errorf(ctx, "post.mongo.GetOne.FindOne: %v", err)
		}
		return models.Post{}, err
	}

	return m, nil
}

func (repo impleRepository) Delete(ctx context.Context, sc models.Scope, id string) error {
	col := repo.getPostCollection()

	filter, err := repo.buildDetailQuery(ctx, sc, id)
	if err != nil {
		repo.l.Errorf(ctx, "post.mongo.Delete.buildDetailQuery: %v", err)
		return err
	}

	_, err = col.DeleteOne(ctx, filter)
	if err != nil {
		repo.l.Errorf(ctx, "post.mongo.Delete.DeleteOne: %v", err)
		return err
	}

	return nil
}

// crawlerPostIDsBeyondRetention returns the IDs to soft-delete after posts are
// already sorted newest-first by the database query.
func crawlerPostIDsBeyondRetention(posts []models.Post, keep int) []primitive.ObjectID {
	if keep < 0 {
		keep = 0
	}
	if len(posts) <= keep {
		return nil
	}
	ids := make([]primitive.ObjectID, 0, len(posts)-keep)
	for _, item := range posts[keep:] {
		ids = append(ids, item.ID)
	}
	return ids
}

// PruneCrawlerPosts keeps the newest crawl-generated posts for one dedicated
// bot author. It intentionally uses soft deletion and a non-empty source URL
// so community/group posts are never included in crawler retention.
func (repo *impleRepository) PruneCrawlerPosts(ctx context.Context, authorID string, keep int) (int64, error) {
	ownerID, err := primitive.ObjectIDFromHex(authorID)
	if err != nil {
		return 0, fmt.Errorf("invalid crawler author id: %w", err)
	}
	if err := repo.ensureCrawlerRetentionIndex(ctx); err != nil {
		return 0, fmt.Errorf("ensure crawler retention index: %w", err)
	}

	filter := mongo.BuildQueryWithSoftDelete(bson.M{
		"author_id":  ownerID,
		"source_url": bson.M{"$exists": true, "$ne": ""},
	})
	cur, err := repo.getPostCollection().Find(ctx, filter, options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetProjection(bson.M{"_id": 1}).
		SetSkip(int64(max(keep, 0))))
	if err != nil {
		return 0, err
	}
	defer cur.Close(ctx)

	var posts []models.Post
	if err := cur.All(ctx, &posts); err != nil {
		return 0, err
	}
	ids := crawlerPostIDsBeyondRetention(posts, 0)
	if len(ids) == 0 {
		return 0, nil
	}

	return repo.getPostCollection().DeleteSoftMany(ctx, bson.M{"_id": bson.M{"$in": ids}})
}
