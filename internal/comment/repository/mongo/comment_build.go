package mongo

import (
	"context"

	"github.com/0Hoag/cryptocheck-api/internal/comment/repository"
	"github.com/0Hoag/cryptocheck-api/internal/models"
	"github.com/0Hoag/cryptocheck-api/pkg/mongo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (repo impleRepository) buildModels(ctx context.Context, sc models.Scope, opts repository.CreateOptions) (models.Comment, error) {
	now := repo.clock()

	postID, err := primitive.ObjectIDFromHex(opts.PostID)
	if err != nil {
		repo.l.Errorf(ctx, "reaction.repository.buildModels.ObjectIDFromHex: %v", err)
		return models.Comment{}, err
	}

	var attachments []models.Attachment
	if len(opts.Attach) > 0 {
		attachments = opts.Attach
	}

	comment := models.Comment{
		ID:          repo.db.NewObjectID(),
		PostID:      postID,
		AuthorID:    mongo.ObjectIDFromHexOrNil(sc.UserID),
		Content:     opts.Content,
		Attachments: attachments,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	return comment, nil
}

func (repo impleRepository) buildUpdateModels(ctx context.Context, sc models.Scope, opts repository.UpdateOptions) (models.Comment, bson.M, error) {
	now := repo.clock()

	set := bson.M{}

	if opts.Content != "" {
		set["content"] = opts.Content
	}

	if len(opts.Attach) > 0 {
		set["attachments"] = opts.Attach
	}

	set["updated_at"] = now
	opts.Comment.UpdatedAt = now
	if opts.Content != "" {
		opts.Comment.Content = opts.Content
	}
	if len(opts.Attach) > 0 {
		opts.Comment.Attachments = opts.Attach
	}

	return opts.Comment, bson.M{"$set": set}, nil
}
