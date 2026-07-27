package usecase

import (
	"context"

	"github.com/0Hoag/cryptocheck-api/internal/models"
	appNotification "github.com/0Hoag/cryptocheck-api/internal/notification"
	"github.com/0Hoag/cryptocheck-api/internal/post"
	"github.com/0Hoag/cryptocheck-api/internal/post/repository"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (uc impleUsecase) CreateReaction(ctx context.Context, sc models.Scope, input post.CreateReactionInput) (models.Reaction, error) {
	if input.Type != models.LikeReaction && input.Type != models.LoveReaction {
		return models.Reaction{}, post.ErrInvalidReactionType
	}

	p, err := uc.repo.Detail(ctx, sc, input.PostID)
	if err != nil {
		uc.l.Errorf(ctx, "post.usecase.CreateReaction.Detail: %v", err)
		return models.Reaction{}, err
	}

	existing, err := uc.repo.ListReaction(ctx, sc, repository.ListReactionOptions{FilterReaction: repository.FilterReaction{
		PostID: input.PostID,
		UserID: sc.UserID,
		Type:   input.Type,
	}})
	if err != nil {
		uc.l.Errorf(ctx, "post.usecase.CreateReaction.ListReaction: %v", err)
		return models.Reaction{}, err
	}
	if len(existing) > 0 {
		return models.Reaction{}, post.ErrReactionAlreadyExists
	}

	reaction, err := uc.repo.CreateReaction(ctx, sc, repository.CreateReactionOptions{
		PostID: input.PostID,
		Type:   input.Type,
	})
	if err != nil {
		uc.l.Errorf(ctx, "post.usecase.CreateReaction.CreateReaction: %v", err)
		return models.Reaction{}, err
	}
	if uc.db != nil {
		actorID, parseErr := primitive.ObjectIDFromHex(sc.UserID)
		if parseErr == nil {
			_ = appNotification.Create(ctx, uc.db, p.AuthorID, actorID, p.ID, "post.reaction_created", "Someone reacted to your post")
		}
	}

	return reaction, nil
}

func (uc impleUsecase) DetailReaction(ctx context.Context, sc models.Scope, id string) (models.Reaction, error) {
	reaction, err := uc.repo.DetailReaction(ctx, sc, id)
	if err != nil {
		uc.l.Errorf(ctx, "post.usecase.DetailReaction.DetailReaction: %v", err)
		return models.Reaction{}, err
	}
	return reaction, nil
}

func (uc impleUsecase) ListReaction(ctx context.Context, sc models.Scope, input post.ListReactionInput) ([]models.Reaction, error) {
	reactions, err := uc.repo.ListReaction(ctx, sc, repository.ListReactionOptions{
		FilterReaction: repository.FilterReaction{
			ID:     input.ID,
			IDs:    input.IDs,
			PostID: input.PostID,
			UserID: input.UserID,
			Type:   input.Type,
		},
	})
	if err != nil {
		uc.l.Errorf(ctx, "post.usecase.ListReaction.ListReaction: %v", err)
		return []models.Reaction{}, err
	}

	return reactions, nil
}

func (uc impleUsecase) GetReaction(ctx context.Context, sc models.Scope, input post.GetReactionInput) (post.GetReactionOutput, error) {
	reactions, paginator, err := uc.repo.GetReaction(ctx, sc, repository.GetReactionOptions{
		FilterReaction: repository.FilterReaction{
			ID:     input.ID,
			IDs:    input.IDs,
			PostID: input.PostID,
			UserID: input.UserID,
			Type:   input.Type,
		},
		PagQuery: input.PagQuery,
	})
	if err != nil {
		uc.l.Errorf(ctx, "post.usecase.GetReaction.GetReaction: %v", err)
		return post.GetReactionOutput{}, err
	}
	return post.GetReactionOutput{
		Reactions: reactions,
		Paginator: paginator,
	}, nil
}

func (uc impleUsecase) DeleteReaction(ctx context.Context, sc models.Scope, id string) error {
	reaction, err := uc.repo.DetailReaction(ctx, models.Scope{}, id)
	if err != nil {
		uc.l.Errorf(ctx, "reaction.usecase.DeleteReaction.DetailReaction: %v", err)
		return err
	}
	if reaction.AuthorID.Hex() != sc.UserID {
		return postDomainPermissionDenied()
	}

	err = uc.repo.DeleteReaction(ctx, sc, id)
	if err != nil {
		uc.l.Errorf(ctx, "reaction.usecase.DeleteReaction.DeleteReaction: %v", err)
		return err
	}
	return nil
}
