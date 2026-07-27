package notification

import (
	"context"
	"strings"
	"time"

	"github.com/0Hoag/cryptocheck-api/internal/models"
	pkgMongo "github.com/0Hoag/cryptocheck-api/pkg/mongo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const Collection = "user_notifications"

// Create persists an in-app notification. Self-notifications are skipped.
func Create(ctx context.Context, db pkgMongo.Database, recipientID, actorID, resourceID primitive.ObjectID, eventType, message string) error {
	if recipientID == primitive.NilObjectID || recipientID == actorID {
		return nil
	}
	n := models.UserNotification{ID: db.NewObjectID(), RecipientID: recipientID, ActorID: actorID, ResourceID: resourceID, Type: strings.TrimSpace(eventType), Message: strings.TrimSpace(message), CreatedAt: time.Now().UTC()}
	_, err := db.Collection(Collection).InsertOne(ctx, n)
	return err
}

func EnsureIndexes(ctx context.Context, db pkgMongo.Database) error {
	_, err := db.Collection(Collection).CreateIndex(ctx, bson.D{{Key: "recipient_id", Value: 1}, {Key: "created_at", Value: -1}}, options.Index().SetName("recipient_notification_feed"))
	return err
}
