package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ScanHistory is a compact, user-owned audit trail of successful token scans.
// It deliberately excludes source code and full provider payloads.
type ScanHistory struct {
	ID             primitive.ObjectID `bson:"_id" json:"id"`
	OwnerID        primitive.ObjectID `bson:"owner_id" json:"-"`
	Input          string             `bson:"input" json:"input"`
	Network        string             `bson:"network" json:"network"`
	AnalysisType   string             `bson:"analysis_type" json:"analysis_type"`
	TrustScore     int                `bson:"trust_score" json:"trust_score"`
	ScoreAvailable bool               `bson:"score_available" json:"score_available"`
	EngineVersion  string             `bson:"engine_version" json:"engine_version"`
	CreatedAt      time.Time          `bson:"created_at" json:"created_at"`
}
