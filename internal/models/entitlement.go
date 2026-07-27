package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type Plan struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Code      string             `bson:"code" json:"code"`
	Name      string             `bson:"name" json:"name"`
	Active    bool               `bson:"active" json:"active"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
}
type Subscription struct {
	ID                primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID            primitive.ObjectID `bson:"user_id" json:"-"`
	PlanCode          string             `bson:"plan_code" json:"plan_code"`
	Provider          string             `bson:"provider,omitempty" json:"provider,omitempty"`
	ProviderReference string             `bson:"provider_reference,omitempty" json:"provider_reference,omitempty"`
	Status            string             `bson:"status" json:"status"`
	StartsAt          time.Time          `bson:"starts_at" json:"starts_at"`
	EndsAt            time.Time          `bson:"ends_at" json:"ends_at"`
	CreatedAt         time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt         time.Time          `bson:"updated_at" json:"updated_at"`
}
