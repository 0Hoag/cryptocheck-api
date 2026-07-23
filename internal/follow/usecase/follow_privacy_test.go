package usecase

import (
	"context"
	"testing"

	"github.com/0Hoag/cryptocheck-api/internal/follow"
	"github.com/0Hoag/cryptocheck-api/internal/follow/repository"
	"github.com/0Hoag/cryptocheck-api/internal/models"
	"github.com/0Hoag/cryptocheck-api/pkg/paginator"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type followRepositoryStub struct {
	follow      models.Follow
	getCalls    int
	detailCalls int
}

func (s *followRepositoryStub) Create(context.Context, models.Scope, repository.CreateOptions) (models.Follow, error) {
	return models.Follow{}, nil
}
func (s *followRepositoryStub) Detail(context.Context, models.Scope, string) (models.Follow, error) {
	s.detailCalls++
	return s.follow, nil
}
func (s *followRepositoryStub) List(context.Context, models.Scope, repository.ListOptions) ([]models.Follow, error) {
	return nil, nil
}
func (s *followRepositoryStub) Get(context.Context, models.Scope, repository.GetOptions) ([]models.Follow, paginator.Paginator, error) {
	s.getCalls++
	return []models.Follow{s.follow}, paginator.Paginator{Total: 1}, nil
}
func (s *followRepositoryStub) Delete(context.Context, models.Scope, string) error { return nil }

func TestFollowRelationshipPrivacy(t *testing.T) {
	ownerID := primitive.NewObjectID()
	otherID := primitive.NewObjectID()
	followID := primitive.NewObjectID().Hex()

	for _, tt := range []struct {
		name        string
		operation   string
		actorID     string
		authorID    string
		wantErr     error
		wantGet     int
		wantDetails int
	}{
		{name: "owner lists their relationships", operation: "get", actorID: ownerID.Hex(), authorID: ownerID.Hex(), wantGet: 1},
		{name: "other member cannot list another users relationships", operation: "get", actorID: otherID.Hex(), authorID: ownerID.Hex(), wantErr: follow.ErrPermissionDenied},
		{name: "member cannot list relationships without own author filter", operation: "get", actorID: ownerID.Hex(), wantErr: follow.ErrPermissionDenied},
		{name: "owner reads follow detail", operation: "detail", actorID: ownerID.Hex(), wantDetails: 1},
		{name: "other member cannot read follow detail", operation: "detail", actorID: otherID.Hex(), wantErr: follow.ErrPermissionDenied, wantDetails: 1},
	} {
		t.Run(tt.name, func(t *testing.T) {
			repo := &followRepositoryStub{follow: models.Follow{ID: primitive.NewObjectID(), AuthorID: ownerID, FolloweeID: primitive.NewObjectID()}}
			uc := impleUsecase{repo: repo}
			var err error
			if tt.operation == "get" {
				_, err = uc.Get(context.Background(), models.Scope{UserID: tt.actorID}, follow.GetInput{Filter: follow.Filter{AuthorID: tt.authorID}, PagQuery: paginator.PaginatorQuery{Page: 1, Limit: 1}})
			} else {
				_, err = uc.Detail(context.Background(), models.Scope{UserID: tt.actorID}, followID)
			}

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, tt.wantGet, repo.getCalls)
			require.Equal(t, tt.wantDetails, repo.detailCalls)
		})
	}
}
