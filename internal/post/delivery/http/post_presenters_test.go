package http

import (
	"testing"

	"github.com/0Hoag/cryptocheck-api/internal/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestPostRequestsAcceptFollowersPermission(t *testing.T) {
	create := createReq{Content: "visible to followers", Permission: string(models.PrivacyTypeFollowers)}
	if err := create.validate(); err != nil {
		t.Fatalf("expected follower-only create request to be valid: %v", err)
	}

	update := updateReq{
		ID:         primitive.NewObjectID().Hex(),
		Content:    "visible to followers",
		Permission: string(models.PrivacyTypeFollowers),
	}
	if err := update.validate(); err != nil {
		t.Fatalf("expected follower-only update request to be valid: %v", err)
	}
}

func TestPostResponseIncludesPermission(t *testing.T) {
	p := models.Post{ID: primitive.NewObjectID(), AuthorID: primitive.NewObjectID(), Permission: models.PrivacyTypeFollowers}
	if got := (handler{}).newPostDataResp(p).Permission; got != string(models.PrivacyTypeFollowers) {
		t.Fatalf("expected followers permission, got %q", got)
	}
}

func TestGetReqRejectsUnsupportedSort(t *testing.T) {
	if err := (getReq{Sort: "popular"}).validate(); err == nil {
		t.Fatal("unsupported sort must be rejected")
	}
	for _, sort := range []string{"", "newest", "oldest"} {
		if err := (getReq{Sort: sort}).validate(); err != nil {
			t.Fatalf("sort %q error = %v", sort, err)
		}
	}
}
