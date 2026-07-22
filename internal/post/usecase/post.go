package usecase

import (
	"context"
	"strings"

	"golang.org/x/sync/errgroup"

	"github.com/0Hoag/cryptocheck-api/internal/models"
	"github.com/0Hoag/cryptocheck-api/internal/post"
	"github.com/0Hoag/cryptocheck-api/internal/post/repository"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (uc impleUsecase) Create(ctx context.Context, sc models.Scope, input post.CreateInput) (models.Post, error) {
	_, ctx = errgroup.WithContext(ctx)

	post, err := uc.repo.Create(ctx, sc, repository.CreateOptions{
		Pin:           input.Pin,
		Title:         input.Title,
		TitleEn:       input.TitleEn,
		Content:       input.Content,
		FullContent:   input.FullContent,
		FullContentEn: input.FullContentEn,
		FileIDs:       input.FileIDs,
		TaggedTarget:  input.TaggedTarget,
		Permission:    input.Permission,
		SourceURL:     input.SourceURL,
	})
	if err != nil {
		uc.l.Errorf(ctx, "post.usecase.Create.Create: %v", err)
		return models.Post{}, err
	}

	if len(post.TaggedTarget) > 0 {
		err = uc.handleCreatePostNotification(ctx, sc, post)
		if err != nil {
			uc.l.Errorf(ctx, "post.usecase.Create.handleCreatePostNotification : %v", err)
			return models.Post{}, nil
		}
	}

	return post, nil
}

func (uc impleUsecase) Detail(ctx context.Context, sc models.Scope, id string) (models.Post, error) {
	post, err := uc.repo.Detail(ctx, sc, id)
	if err != nil {
		uc.l.Errorf(ctx, "post.usecase.Detail.Detail: %v", err)
		return models.Post{}, err
	}
	return post, nil
}

func (uc impleUsecase) List(ctx context.Context, sc models.Scope, input post.ListInput) ([]models.Post, error) {
	posts, err := uc.repo.List(ctx, sc, repository.ListOptions{
		Filter: repository.Filter{
			ID:       input.ID,
			IDs:      input.IDs,
			Pin:      input.Pin,
			AuthorID: input.AuthorID,
		},
	})
	if err != nil {
		uc.l.Errorf(ctx, "post.usecase.List.List: %v", err)
		return []models.Post{}, err
	}
	return posts, nil
}

func (uc impleUsecase) GetOne(ctx context.Context, sc models.Scope, input post.GetOneInput) (models.Post, error) {
	p, err := uc.repo.GetOne(ctx, sc, repository.GetOneOptions{
		Filter: repository.Filter{
			ID:        input.ID,
			IDs:       input.IDs,
			Pin:       input.Pin,
			AuthorID:  input.AuthorID,
			SourceURL: input.SourceURL,
		},
	})
	if err != nil {
		if !strings.Contains(err.Error(), "no documents") {
			uc.l.Errorf(ctx, "post.usecase.GetOne.GetOne: %v", err)
		}
		return models.Post{}, err
	}
	return p, nil
}

func (uc impleUsecase) Get(ctx context.Context, sc models.Scope, input post.GetInput) (post.GetOutput, error) {
	posts, paginator, err := uc.repo.Get(ctx, sc, repository.GetOptions{
		Filter: repository.Filter{
			ID:       input.ID,
			IDs:      input.IDs,
			Pin:      input.Pin,
			AuthorID: input.AuthorID,
		},
		PagQuery: input.PagQuery,
	})
	if err != nil {
		uc.l.Errorf(ctx, "post.usecase.Get.Get: %v", err)
		return post.GetOutput{}, err
	}
	postIDs := make([]primitive.ObjectID, 0, len(posts))
	for _, item := range posts {
		postIDs = append(postIDs, item.ID)
	}
	counts, err := uc.repo.GetEngagementCounts(ctx, postIDs)
	if err != nil {
		uc.l.Errorf(ctx, "post.usecase.Get.GetEngagementCounts: %v", err)
		return post.GetOutput{}, err
	}
	for index := range posts {
		if count, ok := counts[posts[index].ID]; ok {
			posts[index].ReactionCount = count.ReactionCount
			posts[index].CommentCount = count.CommentCount
		}
	}
	return post.GetOutput{
		Posts:     posts,
		Paginator: paginator,
	}, nil
}

func (uc impleUsecase) Update(ctx context.Context, sc models.Scope, input post.UpdateInput) error {
	post, err := uc.repo.Detail(ctx, sc, input.ID)
	if err != nil {
		uc.l.Errorf(ctx, "post.usecase.Update.Detail: %v", err)
		return err
	}
	if post.AuthorID.Hex() != sc.UserID {
		return postDomainPermissionDenied()
	}

	err = uc.repo.Update(ctx, sc, repository.UpdateOptions{
		Post:         post,
		Pin:          input.Pin,
		Content:      input.Content,
		FileIDs:      input.FileIDs,
		TaggedTarget: input.TaggedTarget,
		Permission:   input.Permission,
	})
	if err != nil {
		uc.l.Errorf(ctx, "post.usecase.Update.Update: %v", err)
		return err
	}
	return nil
}

func (uc impleUsecase) Delete(ctx context.Context, sc models.Scope, id string) error {
	p, err := uc.repo.Detail(ctx, sc, id)
	if err != nil {
		uc.l.Errorf(ctx, "post.usecase.Delete.Detail: %v", err)
		return err
	}
	if p.AuthorID.Hex() != sc.UserID {
		return postDomainPermissionDenied()
	}
	err = uc.repo.Delete(ctx, sc, id)
	if err != nil {
		uc.l.Errorf(ctx, "post.usecase.Delete.Delete: %v", err)
		return err
	}
	return nil
}

func postDomainPermissionDenied() error { return post.ErrPermissionDenied }
