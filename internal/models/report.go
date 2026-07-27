package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type ContentReport struct {
	ID         primitive.ObjectID  `bson:"_id,omitempty" json:"id"`
	ReporterID primitive.ObjectID  `bson:"reporter_id" json:"-"`
	TargetType string              `bson:"target_type" json:"target_type"`
	TargetID   primitive.ObjectID  `bson:"target_id" json:"target_id"`
	Reason     string              `bson:"reason" json:"reason"`
	Details    string              `bson:"details,omitempty" json:"details,omitempty"`
	Status     string              `bson:"status" json:"status"`
	HandledBy  *primitive.ObjectID `bson:"handled_by,omitempty" json:"handled_by,omitempty"`
	CreatedAt  time.Time           `bson:"created_at" json:"created_at"`
	UpdatedAt  time.Time           `bson:"updated_at" json:"updated_at"`
}

type ModerationAudit struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ReportID  primitive.ObjectID `bson:"report_id" json:"report_id"`
	ActorID   primitive.ObjectID `bson:"actor_id" json:"actor_id"`
	Action    string             `bson:"action" json:"action"`
	Note      string             `bson:"note,omitempty" json:"note,omitempty"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
}
