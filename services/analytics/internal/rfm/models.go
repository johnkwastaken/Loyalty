package rfm

import (
	"time"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type RFMScore struct {
	ID               primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	OrgID            string            `bson:"org_id" json:"org_id"`
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

var RFMSegments = map[string]string{
	"555": "Champions",
	"554": "Champions", 
	"544": "Champions",
	"545": "Champions",
	"454": "Champions",
	"455": "Champions",
	"445": "Champions",
	
	"543": "Loyal Customers",
	"444": "Loyal Customers",
	"435": "Loyal Customers",
	"355": "Loyal Customers",
	"354": "Loyal Customers",
	"345": "Loyal Customers",
	"344": "Loyal Customers",
	"335": "Loyal Customers",
	
	"512": "Potential Loyalists",
	"511": "Potential Loyalists",
	"422": "Potential Loyalists",
	"421": "Potential Loyalists",
	"412": "Potential Loyalists",
	"411": "Potential Loyalists",
	
	"333": "New Customers",
	"323": "New Customers",
	"322": "New Customers",
	"232": "New Customers",
	"241": "New Customers",
	"251": "New Customers",
	
	"155": "Cannot Lose Them",
	"154": "Cannot Lose Them",
	"144": "Cannot Lose Them",
	"214": "Cannot Lose Them",
	"215": "Cannot Lose Them",
	"115": "Cannot Lose Them",
	
	"245": "At Risk",
	"254": "At Risk",
	"253": "At Risk",
	"244": "At Risk",
	"234": "At Risk",
	
	"331": "Need Attention",
	"321": "Need Attention",
	"312": "Need Attention",
	"231": "Need Attention",
	
	"152": "About to Sleep",
	"242": "About to Sleep",
	"142": "About to Sleep",
	
	"111": "Lost",
	"112": "Lost",
	"121": "Lost",
	"131": "Lost",
	"141": "Lost",
	"151": "Lost",
}