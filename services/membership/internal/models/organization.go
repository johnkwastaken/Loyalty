package models

import (
	"time"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Organization struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	OrgID       string            `bson:"org_id" json:"org_id"`
	Name        string            `bson:"name" json:"name"`
	Description string            `bson:"description" json:"description"`
	Settings    OrgSettings       `bson:"settings" json:"settings"`
	CreatedAt   time.Time         `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time         `bson:"updated_at" json:"updated_at"`
}

type OrgSettings struct {
	PointsPerDollar    float64           `bson:"points_per_dollar" json:"points_per_dollar"`
	StampsPerVisit     int               `bson:"stamps_per_visit" json:"stamps_per_visit"`
	RewardThresholds   []RewardThreshold `bson:"reward_thresholds" json:"reward_thresholds"`
	TierRules          []TierRule        `bson:"tier_rules" json:"tier_rules"`
	MaxStampsPerCard   int               `bson:"max_stamps_per_card" json:"max_stamps_per_card"`
}

type RewardThreshold struct {
	Points      int    `bson:"points" json:"points"`
	Stamps      int    `bson:"stamps" json:"stamps"`
	RewardType  string `bson:"reward_type" json:"reward_type"`
	RewardValue string `bson:"reward_value" json:"reward_value"`
	Description string `bson:"description" json:"description"`
}

type TierRule struct {
	Name            string  `bson:"name" json:"name"`
	MinSpent        float64 `bson:"min_spent" json:"min_spent"`
	MinVisits       int     `bson:"min_visits" json:"min_visits"`
	PointsMultiplier float64 `bson:"points_multiplier" json:"points_multiplier"`
	Benefits        []string `bson:"benefits" json:"benefits"`
}