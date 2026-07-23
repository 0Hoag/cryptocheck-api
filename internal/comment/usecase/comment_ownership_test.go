package usecase

import (
	"context"
	"testing"

	"github.com/0Hoag/cryptocheck-api/internal/comment"
	"github.com/0Hoag/cryptocheck-api/internal/comment/repository"
	"github.com/0Hoag/cryptocheck-api/internal/models"
	"github.com/0Hoag/cryptocheck-api/pkg/paginator"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type commentRepositoryStub struct {
	comment      models.Comment
	updateCalls  int
	deleteCalls  int
	getOneFilter repository.Filter
}

func (s *commentRepositoryStub) Create(context.Context, models.Scope, repository.CreateOptions) (models.Comment, error) {
	return models.Comment{}, nil
}
func (s *commentRepositoryStub) Detail(context.Context, models.Scope, string) (models.Comment, error) {
	return s.comment, nil
}
func (s *commentRepositoryStub) GetOne(_ context.Context, _ models.Scope, opts repository.GetOneOptions) (models.Comment, error) {
	s.getOneFilter = opts.Filter
	return s.comment, nil
}
func (s *commentRepositoryStub) List(context.Context, models.Scope, repository.ListOptions) ([]models.Comment, error) {
	return nil, nil
}
func (s *commentRepositoryStub) Get(context.Context, models.Scope, repository.GetOptions) ([]models.Comment, paginator.Paginator, error) {
	return nil, paginator.Paginator{}, nil
}
func (s *commentRepositoryStub) Update(_ context.Context, _ models.Scope, _ repository.UpdateOptions) (models.Comment, error) {
	s.updateCalls++
	return s.comment, nil
}
func (s *commentRepositoryStub) Delete(context.Context, models.Scope, string) error {
	s.deleteCalls++
	return nil
}

func TestCommentOwnership(t *testing.T) {
	ownerID := primitive.NewObjectID()
	otherID := primitive.NewObjectID()
	commentID := primitive.NewObjectID().Hex()

	tests := []struct {
		name       string
		operation  string
		actorID    string
		wantErr    error
		wantUpdate int
		wantDelete int
	}{
		{name: "owner updates comment", operation: "update", actorID: ownerID.Hex(), wantUpdate: 1},
		{name: "other member cannot update comment", operation: "update", actorID: otherID.Hex(), wantErr: comment.ErrPermissionDenied},
		{name: "owner deletes comment", operation: "delete", actorID: ownerID.Hex(), wantDelete: 1},
		{name: "other member cannot delete comment", operation: "delete", actorID: otherID.Hex(), wantErr: comment.ErrPermissionDenied},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &commentRepositoryStub{comment: models.Comment{ID: primitive.NewObjectID(), AuthorID: ownerID, Content: "original"}}
			uc := impleUsecase{repo: repo}
			var err error
			if tt.operation == "update" {
				_, err = uc.Update(context.Background(), models.Scope{UserID: tt.actorID}, comment.UpdateInput{ID: commentID, Content: "updated"})
			} else {
				err = uc.Delete(context.Background(), models.Scope{UserID: tt.actorID}, commentID)
			}

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, tt.wantUpdate, repo.updateCalls)
			require.Equal(t, tt.wantDelete, repo.deleteCalls)
			if tt.operation == "update" {
				require.Equal(t, commentID, repo.getOneFilter.ID)
			}
		})
	}
}
