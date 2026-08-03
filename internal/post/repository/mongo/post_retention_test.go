package mongo

import (
	"testing"

	"github.com/0Hoag/cryptocheck-api/internal/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestCrawlerPostIDsBeyondRetention(t *testing.T) {
	first, second, third := primitive.NewObjectID(), primitive.NewObjectID(), primitive.NewObjectID()
	posts := []models.Post{{ID: first}, {ID: second}, {ID: third}}

	ids := crawlerPostIDsBeyondRetention(posts, 2)
	if len(ids) != 1 || ids[0] != third {
		t.Fatalf("retained IDs = %v, want only the oldest post", ids)
	}
	if ids := crawlerPostIDsBeyondRetention(posts, 3); len(ids) != 0 {
		t.Fatalf("retained IDs = %v, want none", ids)
	}
	if ids := crawlerPostIDsBeyondRetention(posts, -1); len(ids) != 3 {
		t.Fatalf("negative keep IDs = %v, want all posts", ids)
	}
}
