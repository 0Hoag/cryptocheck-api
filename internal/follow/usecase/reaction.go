package usecase

import (
	"context"

	"github.com/0Hoag/cryptocheck-api/internal/follow"
	"github.com/0Hoag/cryptocheck-api/internal/follow/repository"
	"github.com/0Hoag/cryptocheck-api/internal/models"
	"github.com/0Hoag/cryptocheck-api/internal/users"
)

func (uc impleUsecase) Create(ctx context.Context, sc models.Scope, input follow.CreateInput) (models.Follow, error) {
	if input.FolloweeID == sc.UserID {
		return models.Follow{}, follow.ErrCannotFollowSelf
	}

	_, err := uc.userUC.GetOne(ctx, users.Filter{ID: input.FolloweeID})
	if err != nil {
		uc.l.Errorf(ctx, "follow.usecase.Create.GetOne: %v", err)
		return models.Follow{}, err
	}

	existing, err := uc.repo.List(ctx, sc, repository.ListOptions{Filter: repository.Filter{
		AuthorID:   sc.UserID,
		FolloweeID: input.FolloweeID,
	}})
	if err != nil {
		uc.l.Errorf(ctx, "follow.usecase.Create.List: %v", err)
		return models.Follow{}, err
	}
	if len(existing) > 0 {
		return models.Follow{}, follow.ErrAlreadyFollowing
	}

	created, err := uc.repo.Create(ctx, sc, repository.CreateOptions{FolloweeID: input.FolloweeID})
	if err != nil {
		uc.l.Errorf(ctx, "follow.usecase.Create.Create: %v", err)
		return models.Follow{}, err
	}

	return created, nil
}

func (uc impleUsecase) Detail(ctx context.Context, sc models.Scope, id string) (models.Follow, error) {
	follow, err := uc.repo.Detail(ctx, sc, id)
	if err != nil {
		uc.l.Errorf(ctx, "follow.usecase.Detail.Detail: %v", err)
		return models.Follow{}, err
	}
	return follow, nil
}

func (uc impleUsecase) List(ctx context.Context, sc models.Scope, input follow.ListInput) ([]models.Follow, error) {
	follows, err := uc.repo.List(ctx, sc, repository.ListOptions{
		Filter: repository.Filter{
			ID:         input.ID,
			IDs:        input.IDs,
			AuthorID:   input.AuthorID,
			FolloweeID: input.FolloweeID,
		},
	})
	if err != nil {
		uc.l.Errorf(ctx, "follow.usecase.List.List: %v", err)
		return []models.Follow{}, err
	}
	return follows, nil
}

func (uc impleUsecase) Get(ctx context.Context, sc models.Scope, input follow.GetInput) (follow.GetOutput, error) {
	follows, paginator, err := uc.repo.Get(ctx, sc, repository.GetOptions{
		Filter: repository.Filter{
			ID:         input.ID,
			IDs:        input.IDs,
			AuthorID:   input.AuthorID,
			FolloweeID: input.FolloweeID,
		},
		PagQuery: input.PagQuery,
	})
	if err != nil {
		uc.l.Errorf(ctx, "follow.usecase.Get.Get: %v", err)
		return follow.GetOutput{}, err
	}
	return follow.GetOutput{
		Follows:   follows,
		Paginator: paginator,
	}, nil
}

func (uc impleUsecase) Delete(ctx context.Context, sc models.Scope, id string) error {
	item, err := uc.repo.Detail(ctx, models.Scope{}, id)
	if err != nil {
		uc.l.Errorf(ctx, "follow.usecase.Delete.Detail: %v", err)
		return err
	}
	if item.AuthorID.Hex() != sc.UserID {
		return follow.ErrPermissionDenied
	}

	err = uc.repo.Delete(ctx, models.Scope{}, id)
	if err != nil {
		uc.l.Errorf(ctx, "follow.usecase.Delete.Delete: %v", err)
		return err
	}
	return nil
}
