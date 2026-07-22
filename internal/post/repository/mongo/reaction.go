package mongo

import (
	"context"

	"github.com/0Hoag/cryptocheck-api/internal/models"
	"github.com/0Hoag/cryptocheck-api/internal/post"
	"github.com/0Hoag/cryptocheck-api/internal/post/repository"
	"github.com/0Hoag/cryptocheck-api/pkg/mongo"
	"github.com/0Hoag/cryptocheck-api/pkg/paginator"
	driverMongo "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	ReactionCollection = "reaction"
)

func (repo impleRepository) getReactionCollection() mongo.Collection {
	return repo.db.Collection(ReactionCollection)
}

func (repo *impleRepository) CreateReaction(ctx context.Context, sc models.Scope, opts repository.CreateReactionOptions) (models.Reaction, error) {
	col := repo.getReactionCollection()
	if err := repo.ensureReactionIndex(ctx); err != nil {
		repo.l.Errorf(ctx, "post.mongo.CreateReaction.ensureReactionIndex: %v", err)
		return models.Reaction{}, err
	}

	m, err := repo.buildReactionModels(ctx, sc, opts)
	if err != nil {
		repo.l.Errorf(ctx, "post.mongo.CreateReaction.buildReactionModels: %v", err)
		return models.Reaction{}, err
	}

	_, err = col.InsertOne(ctx, m)
	if err != nil {
		if driverMongo.IsDuplicateKeyError(err) {
			return models.Reaction{}, post.ErrReactionAlreadyExists
		}
		repo.l.Errorf(ctx, "Reactions.mogno.CreateReaction.InsertOne: %v", err)
		return models.Reaction{}, err
	}

	return m, nil
}

func (repo impleRepository) DetailReaction(ctx context.Context, sc models.Scope, id string) (models.Reaction, error) {
	col := repo.getReactionCollection()

	filter, err := repo.buildDetailReactionQuery(ctx, sc, id)
	if err != nil {
		repo.l.Errorf(ctx, "post.mongo.DetailReaction.buildDetailQuery: %v", err)
		return models.Reaction{}, err
	}

	var m models.Reaction
	err = col.FindOne(ctx, filter).Decode(&m)
	if err != nil {
		repo.l.Errorf(ctx, "post.mongo.DetailReaction.FindOne: %v", err)
		return models.Reaction{}, err
	}

	return m, nil
}

func (repo impleRepository) ListReaction(ctx context.Context, sc models.Scope, opts repository.ListReactionOptions) ([]models.Reaction, error) {
	col := repo.getReactionCollection()

	filter, err := repo.buildListReactionQuery(ctx, sc, opts)
	if err != nil {
		repo.l.Errorf(ctx, "post.mongo.ListReaction.buildListReactionQuery: %v", err)
		return []models.Reaction{}, err
	}

	cur, err := col.Find(ctx, filter)
	if err != nil {
		repo.l.Errorf(ctx, "post.mongo.ListReaction.Find: %v", err)
		return []models.Reaction{}, err
	}

	var ms []models.Reaction
	err = cur.All(ctx, &ms)
	if err != nil {
		repo.l.Errorf(ctx, "post.mongo.ListReaction.All: %v", err)
		return []models.Reaction{}, err
	}

	return ms, nil
}

func (repo impleRepository) GetReaction(ctx context.Context, sc models.Scope, opts repository.GetReactionOptions) ([]models.Reaction, paginator.Paginator, error) {
	col := repo.getReactionCollection()

	filter, err := repo.buildGetReactionQuery(ctx, sc, opts)
	if err != nil {
		repo.l.Errorf(ctx, "post.mongo.GetReaction.buildGetReactionQuery: %v", err)
		return []models.Reaction{}, paginator.Paginator{}, err
	}

	cur, err := col.Find(ctx, filter, options.Find().
		SetLimit(opts.PagQuery.Limit).
		SetSkip(opts.PagQuery.Offset()))
	if err != nil {
		repo.l.Errorf(ctx, "post.mongo.GetReaction.Find: %v", err)
		return []models.Reaction{}, paginator.Paginator{}, err
	}

	var ms []models.Reaction
	err = cur.All(ctx, &ms)
	if err != nil {
		repo.l.Errorf(ctx, "post.mongo.GetReaction.All: %v", err)
		return []models.Reaction{}, paginator.Paginator{}, err
	}

	total, err := col.CountDocuments(ctx, filter)
	if err != nil {
		repo.l.Errorf(ctx, "post.mongo.GetReaction.CountDocuments: %v", err)
		return []models.Reaction{}, paginator.Paginator{}, err
	}

	return ms, paginator.Paginator{
		Total:       total,
		Count:       int64(len(ms)),
		PerPage:     opts.PagQuery.Limit,
		CurrentPage: opts.PagQuery.Page,
	}, nil
}

func (repo impleRepository) DeleteReaction(ctx context.Context, sc models.Scope, id string) error {
	col := repo.getReactionCollection()

	filter, err := repo.buildDetailReactionQuery(ctx, sc, id)
	if err != nil {
		repo.l.Errorf(ctx, "post.mongo.DeleteReaction.buildDetailQuery: %v", err)
		return err
	}

	_, err = col.DeleteOne(ctx, filter)
	if err != nil {
		repo.l.Errorf(ctx, "post.mongo.DeleteReaction.DeleteOne: %v", err)
		return err
	}

	return nil
}
