package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// PrelaunchProject is due-diligence data for a project before a deployable
// token contract exists. It intentionally has no security score.
type PrelaunchProject struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	OwnerID      primitive.ObjectID `bson:"owner_id" json:"-"`
	Name         string             `bson:"name" json:"name"`
	Symbol       string             `bson:"symbol,omitempty" json:"symbol,omitempty"`
	WebsiteURL   string             `bson:"website_url" json:"website_url"`
	SocialURLs   []string           `bson:"social_urls,omitempty" json:"social_urls,omitempty"`
	ClaimedChain string             `bson:"claimed_chain,omitempty" json:"claimed_chain,omitempty"`
	LaunchAt     *time.Time         `bson:"launch_at,omitempty" json:"launch_at,omitempty"`
	Evidence     []string           `bson:"evidence,omitempty" json:"evidence,omitempty"`
	RiskFlags    []string           `bson:"risk_flags,omitempty" json:"risk_flags,omitempty"`
	CreatedAt    time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt    time.Time          `bson:"updated_at" json:"updated_at"`
	DeletedAt    *time.Time         `bson:"deleted_at,omitempty" json:"-"`
	IsOwner      bool               `bson:"-" json:"is_owner,omitempty"`
}
