package mongo

import (
	"testing"

	"github.com/0Hoag/cryptocheck-api/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestVisiblePostFilters(t *testing.T) {
	viewerID := primitive.NewObjectID()
	followeeID := primitive.NewObjectID()
	publicFilter := bson.M{"permission": bson.M{"$in": bson.A{models.PrivacyTypePublic, ""}}}

	filters := visiblePostFilters(publicFilter, viewerID, []primitive.ObjectID{followeeID})
	if len(filters) != 3 {
		t.Fatalf("expected public, owner and follower filters; got %d", len(filters))
	}

	followerFilter, ok := filters[2].(bson.M)
	if !ok {
		t.Fatalf("expected follower filter to be bson.M, got %T", filters[2])
	}
	if got := followerFilter["permission"]; got != models.PrivacyTypeFollowers {
		t.Fatalf("expected followers permission, got %#v", got)
	}
	authorFilter, ok := followerFilter["author_id"].(bson.M)
	if !ok {
		t.Fatalf("expected author filter to be bson.M, got %T", followerFilter["author_id"])
	}
	ids := authorFilter["$in"].([]primitive.ObjectID)
	if len(ids) != 1 || ids[0] != followeeID {
		t.Fatalf("expected followed author in filter, got %#v", ids)
	}
}

func TestVisiblePostFiltersOmitsFollowerBranchWithoutFollows(t *testing.T) {
	filters := visiblePostFilters(bson.M{"permission": "public"}, primitive.NewObjectID(), nil)
	if len(filters) != 2 {
		t.Fatalf("expected public and owner filters only, got %d", len(filters))
	}
}
