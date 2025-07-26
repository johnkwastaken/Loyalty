package tiers

import (
	"time"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CustomerTier struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	OrgID           string            `bson:"org_id" json:"org_id"`
	LocationID      string            `bson:"location_id" json:"location_id"`
	CustomerID      string            `bson:"customer_id" json:"customer_id"`
	CurrentTier     string            `bson:"current_tier" json:"current_tier"`
	PreviousTier    string            `bson:"previous_tier" json:"previous_tier"`
	TierSince       time.Time         `bson:"tier_since" json:"tier_since"`
	NextTier        string            `bson:"next_tier" json:"next_tier"`
	ProgressToNext  float64           `bson:"progress_to_next" json:"progress_to_next"`
	
	// Metrics for tier calculation
	TotalSpent       float64   `bson:"total_spent" json:"total_spent"`
	TotalVisits      int       `bson:"total_visits" json:"total_visits"`
	SpentThisYear    float64   `bson:"spent_this_year" json:"spent_this_year"`
	VisitsThisYear   int       `bson:"visits_this_year" json:"visits_this_year"`
	SpentThisMonth   float64   `bson:"spent_this_month" json:"spent_this_month"`
	VisitsThisMonth  int       `bson:"visits_this_month" json:"visits_this_month"`
	LastTransaction  time.Time `bson:"last_transaction" json:"last_transaction"`
	
	// Benefits and multipliers
	PointsMultiplier float64   `bson:"points_multiplier" json:"points_multiplier"`
	Benefits         []string  `bson:"benefits" json:"benefits"`
	
	CalculatedAt     time.Time `bson:"calculated_at" json:"calculated_at"`
	UpdatedAt        time.Time `bson:"updated_at" json:"updated_at"`
}

type TierRule struct {
	Name              string    `bson:"name" json:"name"`
	Level             int       `bson:"level" json:"level"`
	MinSpentLifetime  float64   `bson:"min_spent_lifetime" json:"min_spent_lifetime"`
	MinSpentYear      float64   `bson:"min_spent_year" json:"min_spent_year"`
	MinVisitsLifetime int       `bson:"min_visits_lifetime" json:"min_visits_lifetime"`
	MinVisitsYear     int       `bson:"min_visits_year" json:"min_visits_year"`
	PointsMultiplier  float64   `bson:"points_multiplier" json:"points_multiplier"`
	Benefits          []string  `bson:"benefits" json:"benefits"`
	Color             string    `bson:"color" json:"color"`
	Icon              string    `bson:"icon" json:"icon"`
}

type OrgTierConfig struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	OrgID      string            `bson:"org_id" json:"org_id"`
	TierRules  []TierRule        `bson:"tier_rules" json:"tier_rules"`
	CreatedAt  time.Time         `bson:"created_at" json:"created_at"`
	UpdatedAt  time.Time         `bson:"updated_at" json:"updated_at"`
}

type TierUpgrade struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	OrgID        string            `bson:"org_id" json:"org_id"`
	CustomerID   string            `bson:"customer_id" json:"customer_id"`
	FromTier     string            `bson:"from_tier" json:"from_tier"`
	ToTier       string            `bson:"to_tier" json:"to_tier"`
	TriggeredBy  string            `bson:"triggered_by" json:"triggered_by"`
	TriggerValue float64           `bson:"trigger_value" json:"trigger_value"`
	UpgradedAt   time.Time         `bson:"upgraded_at" json:"upgraded_at"`
	Notified     bool              `bson:"notified" json:"notified"`
}

type CustomerMetrics struct {
	OrgID            string    `json:"org_id"`
	LocationID       string    `json:"location_id"`
	CustomerID       string    `json:"customer_id"`
	TotalSpent       float64   `json:"total_spent"`
	TotalVisits      int       `json:"total_visits"`
	SpentThisYear    float64   `json:"spent_this_year"`
	VisitsThisYear   int       `json:"visits_this_year"`
	SpentThisMonth   float64   `json:"spent_this_month"`
	VisitsThisMonth  int       `json:"visits_this_month"`
	LastTransaction  time.Time `json:"last_transaction"`
	TransactionAmount float64  `json:"transaction_amount"`
}

func GetDefaultTierRules() []TierRule {
	return []TierRule{
		{
			Name:              "Bronze",
			Level:             1,
			MinSpentLifetime:  0,
			MinSpentYear:      0,
			MinVisitsLifetime: 0,
			MinVisitsYear:     0,
			PointsMultiplier:  1.0,
			Benefits:          []string{"Basic rewards", "Birthday bonus"},
			Color:             "#CD7F32",
			Icon:              "bronze-medal",
		},
		{
			Name:              "Silver",
			Level:             2,
			MinSpentLifetime:  250,
			MinSpentYear:      100,
			MinVisitsLifetime: 5,
			MinVisitsYear:     3,
			PointsMultiplier:  1.25,
			Benefits:          []string{"25% bonus points", "Priority support", "Birthday bonus", "Monthly offers"},
			Color:             "#C0C0C0",
			Icon:              "silver-medal",
		},
		{
			Name:              "Gold",
			Level:             3,
			MinSpentLifetime:  750,
			MinSpentYear:      300,
			MinVisitsLifetime: 15,
			MinVisitsYear:     8,
			PointsMultiplier:  1.5,
			Benefits:          []string{"50% bonus points", "Free shipping", "Early access", "VIP support", "Birthday bonus"},
			Color:             "#FFD700",
			Icon:              "gold-medal",
		},
		{
			Name:              "Platinum",
			Level:             4,
			MinSpentLifetime:  2000,
			MinSpentYear:      800,
			MinVisitsLifetime: 30,
			MinVisitsYear:     15,
			PointsMultiplier:  2.0,
			Benefits:          []string{"Double points", "Exclusive events", "Personal concierge", "Premium support"},
			Color:             "#E5E4E2",
			Icon:              "platinum-medal",
		},
		{
			Name:              "Diamond",
			Level:             5,
			MinSpentLifetime:  5000,
			MinSpentYear:      2000,
			MinVisitsLifetime: 50,
			MinVisitsYear:     25,
			PointsMultiplier:  3.0,
			Benefits:          []string{"Triple points", "Lifetime benefits", "Dedicated manager", "Annual gifts"},
			Color:             "#B9F2FF",
			Icon:              "diamond",
		},
	}
}