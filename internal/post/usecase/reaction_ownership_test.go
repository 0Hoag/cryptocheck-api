package usecase

import (
	"context"
	"testing"

	"github.com/0Hoag/cryptocheck-api/internal/models"
	"github.com/0Hoag/cryptocheck-api/internal/post"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestReactionDeleteOwnership(t *testing.T) {
	ownerID := primitive.NewObjectID()
	otherID := primitive.NewObjectID()
	reactionID := primitive.NewObjectID().Hex()

	for _, tt := range []struct {
		name        string
		actorID     string
		wantErr     error
		wantDeletes int
	}{
		{name: "owner deletes reaction", actorID: ownerID.Hex(), wantDeletes: 1},
		{name: "other member cannot delete reaction", actorID: otherID.Hex(), wantErr: post.ErrPermissionDenied},
	} {
		t.Run(tt.name, func(t *testing.T) {
			repo := &postRepositoryStub{reaction: models.Reaction{ID: primitive.NewObjectID(), AuthorID: ownerID}}
			uc := impleUsecase{repo: repo}

			err := uc.DeleteReaction(context.Background(), models.Scope{UserID: tt.actorID}, reactionID)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, tt.wantDeletes, repo.reactionDeleteCalls)
		})
	}
}
