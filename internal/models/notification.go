package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// UserNotification is a compact in-app event delivered to one user.
type UserNotification struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	RecipientID primitive.ObjectID `bson:"recipient_id" json:"-"`
	ActorID     primitive.ObjectID `bson:"actor_id,omitempty" json:"actor_id,omitempty"`
	Type        string             `bson:"type" json:"type"`
	ResourceID  primitive.ObjectID `bson:"resource_id" json:"resource_id"`
	Message     string             `bson:"message" json:"message"`
	ReadAt      *time.Time         `bson:"read_at,omitempty" json:"read_at,omitempty"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
}
