package models

import (
	"time"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type RFMScore struct {
	ID               primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	OrgID            string            `bson:"org_id" json:"org_id"`
	LocationID       string            `bson:"location_id" json:"location_id"`
	CustomerID       string            `bson:"customer_id" json:"customer_id"`
	RecencyScore     int               `bson:"recency_score" json:"recency_score"`
	FrequencyScore   int               `bson:"frequency_score" json:"frequency_score"`
	MonetaryScore    int               `bson:"monetary_score" json:"monetary_score"`
	RFMSegment       string            `bson:"rfm_segment" json:"rfm_segment"`
	LastTransaction  time.Time         `bson:"last_transaction" json:"last_transaction"`
	TotalTransactions int              `bson:"total_transactions" json:"total_transactions"`
	TotalSpent       float64           `bson:"total_spent" json:"total_spent"`
	AvgOrderValue    float64           `bson:"avg_order_value" json:"avg_order_value"`
	DaysSinceFirst   int               `bson:"days_since_first" json:"days_since_first"`
	DaysSinceLast    int               `bson:"days_since_last" json:"days_since_last"`
	CalculatedAt     time.Time         `bson:"calculated_at" json:"calculated_at"`
	UpdatedAt        time.Time         `bson:"updated_at" json:"updated_at"`
}

type CustomerActivity struct {
	OrgID             string    `json:"org_id"`
	LocationID        string    `json:"location_id"`
	CustomerID        string    `json:"customer_id"`
	TransactionDate   time.Time `json:"transaction_date"`
	Amount            float64   `json:"amount"`
	FirstTransaction  time.Time `json:"first_transaction"`
	LastTransaction   time.Time `json:"last_transaction"`
	TotalTransactions int       `json:"total_transactions"`
	TotalSpent        float64   `json:"total_spent"`
}

type RFMQuintiles struct {
	OrgID              string    `bson:"org_id" json:"org_id"`
	RecencyQuintiles   []int     `bson:"recency_quintiles" json:"recency_quintiles"`
	FrequencyQuintiles []int     `bson:"frequency_quintiles" json:"frequency_quintiles"`
	MonetaryQuintiles  []float64 `bson:"monetary_quintiles" json:"monetary_quintiles"`
	CalculatedAt       time.Time `bson:"calculated_at" json:"calculated_at"`
}