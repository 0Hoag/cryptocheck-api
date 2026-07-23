package usecase

import (
	"context"
	"testing"

	"github.com/0Hoag/cryptocheck-api/internal/models"
	"github.com/0Hoag/cryptocheck-api/internal/post"
	"github.com/0Hoag/cryptocheck-api/internal/post/repository"
	"github.com/0Hoag/cryptocheck-api/pkg/paginator"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type postRepositoryStub struct {
	post                models.Post
	updateCalls         int
	deleteCalls         int
	updatedWith         repository.UpdateOptions
	reaction            models.Reaction
	reactionDeleteCalls int
}

func (s *postRepositoryStub) Create(context.Context, models.Scope, repository.CreateOptions) (models.Post, error) {
	return models.Post{}, nil
}
func (s *postRepositoryStub) Detail(context.Context, models.Scope, string) (models.Post, error) {
	return s.post, nil
}
func (s *postRepositoryStub) List(context.Context, models.Scope, repository.ListOptions) ([]models.Post, error) {
	return nil, nil
}
func (s *postRepositoryStub) Get(context.Context, models.Scope, repository.GetOptions) ([]models.Post, paginator.Paginator, error) {
	return nil, paginator.Paginator{}, nil
}
func (s *postRepositoryStub) GetOne(context.Context, models.Scope, repository.GetOneOptions) (models.Post, error) {
	return models.Post{}, nil
}
func (s *postRepositoryStub) Update(_ context.Context, _ models.Scope, opts repository.UpdateOptions) error {
	s.updateCalls++
	s.updatedWith = opts
	return nil
}
func (s *postRepositoryStub) Delete(context.Context, models.Scope, string) error {
	s.deleteCalls++
	return nil
}
func (s *postRepositoryStub) GetEngagementCounts(context.Context, []primitive.ObjectID) (map[primitive.ObjectID]repository.EngagementCounts, error) {
	return nil, nil
}
func (s *postRepositoryStub) GetAuthorSummaries(context.Context, []primitive.ObjectID) (map[primitive.ObjectID]repository.AuthorSummary, error) {
	return nil, nil
}
func (s *postRepositoryStub) CreateReaction(context.Context, models.Scope, repository.CreateReactionOptions) (models.Reaction, error) {
	return models.Reaction{}, nil
}
func (s *postRepositoryStub) DetailReaction(context.Context, models.Scope, string) (models.Reaction, error) {
	return s.reaction, nil
}
func (s *postRepositoryStub) ListReaction(context.Context, models.Scope, repository.ListReactionOptions) ([]models.Reaction, error) {
	return nil, nil
}
func (s *postRepositoryStub) GetReaction(context.Context, models.Scope, repository.GetReactionOptions) ([]models.Reaction, paginator.Paginator, error) {
	return nil, paginator.Paginator{}, nil
}
func (s *postRepositoryStub) DeleteReaction(context.Context, models.Scope, string) error {
	s.reactionDeleteCalls++
	return nil
}
func (s *postRepositoryStub) CreateComment(context.Context, models.Scope, repository.CreateCommentOptions) (models.Comment, error) {
	return models.Comment{}, nil
}
func (s *postRepositoryStub) DetailComment(context.Context, models.Scope, string) (models.Comment, error) {
	return models.Comment{}, nil
}
func (s *postRepositoryStub) GetOneComment(context.Context, models.Scope, repository.GetOneCommentOptions) (models.Comment, error) {
	return models.Comment{}, nil
}
func (s *postRepositoryStub) ListComment(context.Context, models.Scope, repository.ListCommentOptions) ([]models.Comment, error) {
	return nil, nil
}
func (s *postRepositoryStub) GetComment(context.Context, models.Scope, repository.GetCommentOptions) ([]models.Comment, paginator.Paginator, error) {
	return nil, paginator.Paginator{}, nil
}
func (s *postRepositoryStub) UpdateComment(context.Context, models.Scope, repository.UpdateCommentOptions) (models.Comment, error) {
	return models.Comment{}, nil
}
func (s *postRepositoryStub) DeleteComment(context.Context, models.Scope, string) error { return nil }

func TestPostOwnership(t *testing.T) {
	ownerID := primitive.NewObjectID()
	otherID := primitive.NewObjectID()
	postID := primitive.NewObjectID().Hex()

	tests := []struct {
		name       string
		operation  string
		actorID    string
		wantErr    error
		wantUpdate int
		wantDelete int
	}{
		{name: "owner updates post", operation: "update", actorID: ownerID.Hex(), wantUpdate: 1},
		{name: "other member cannot update post", operation: "update", actorID: otherID.Hex(), wantErr: post.ErrPermissionDenied},
		{name: "owner deletes post", operation: "delete", actorID: ownerID.Hex(), wantDelete: 1},
		{name: "other member cannot delete post", operation: "delete", actorID: otherID.Hex(), wantErr: post.ErrPermissionDenied},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &postRepositoryStub{post: models.Post{ID: primitive.NewObjectID(), AuthorID: ownerID, Content: "original"}}
			uc := impleUsecase{repo: repo}
			var err error
			if tt.operation == "update" {
				err = uc.Update(context.Background(), models.Scope{UserID: tt.actorID}, post.UpdateInput{ID: postID, Content: "updated", Permission: "public"})
			} else {
				err = uc.Delete(context.Background(), models.Scope{UserID: tt.actorID}, postID)
			}

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, tt.wantUpdate, repo.updateCalls)
			require.Equal(t, tt.wantDelete, repo.deleteCalls)
			if tt.wantUpdate > 0 {
				require.Equal(t, "updated", repo.updatedWith.Content)
				require.Equal(t, ownerID, repo.updatedWith.Post.AuthorID)
			}
		})
	}
}
