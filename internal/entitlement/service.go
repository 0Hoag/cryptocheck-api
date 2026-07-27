package entitlement

import (
	"context"
	"github.com/0Hoag/cryptocheck-api/internal/models"
	pkgMongo "github.com/0Hoag/cryptocheck-api/pkg/mongo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

const PlansCollection = "plans"
const SubscriptionsCollection = "subscriptions"

func Has(ctx context.Context, db pkgMongo.Database, userID primitive.ObjectID, planCode string, now time.Time) bool {
	var s models.Subscription
	err := db.Collection(SubscriptionsCollection).FindOne(ctx, bson.M{"user_id": userID, "plan_code": planCode, "status": "active", "starts_at": bson.M{"$lte": now}, "ends_at": bson.M{"$gt": now}}).Decode(&s)
	return err == nil
}
func EnsureIndexes(ctx context.Context, db pkgMongo.Database) error {
	if _, err := db.Collection(PlansCollection).CreateIndex(ctx, bson.D{{Key: "code", Value: 1}}, options.Index().SetUnique(true)); err != nil {
		return err
	}
	_, err := db.Collection(SubscriptionsCollection).CreateIndex(ctx, bson.D{{Key: "user_id", Value: 1}, {Key: "plan_code", Value: 1}, {Key: "ends_at", Value: -1}}, options.Index().SetName("user_entitlements"))
	return err
}
func Active(s models.Subscription, now time.Time) bool {
	return s.Status == "active" && !s.StartsAt.After(now) && s.EndsAt.After(now)
}
