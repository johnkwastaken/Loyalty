package models

import "time"

type EventType string

const (
	EventTypePOSTransaction   EventType = "pos.transaction"
	EventTypeLoyaltyAction    EventType = "loyalty.action"
	EventTypeCustomerUpdated  EventType = "customer.updated"
	EventTypeRewardTriggered  EventType = "reward.triggered"
)

type BaseEvent struct {
	EventID     string                 `json:"event_id"`
	EventType   EventType              `json:"event_type"`
	OrgID       string                 `json:"org_id"`
	LocationID  string                 `json:"location_id"`
	CustomerID  string                 `json:"customer_id"`
	Timestamp   time.Time              `json:"timestamp"`
	Payload     map[string]interface{} `json:"payload"`
	Metadata    map[string]interface{} `json:"metadata"`
}

type POSTransaction struct {
	TransactionID string    `json:"transaction_id"`
	Amount        float64   `json:"amount"`
	Items         []LineItem `json:"items"`
	PaymentMethod string    `json:"payment_method"`
	ReceiptNumber string    `json:"receipt_number"`
	Cashier       string    `json:"cashier"`
}

type LineItem struct {
	SKU         string  `json:"sku"`
	Name        string  `json:"name"`
	Quantity    int     `json:"quantity"`
	UnitPrice   float64 `json:"unit_price"`
	TotalPrice  float64 `json:"total_price"`
	Category    string  `json:"category"`
	Discounted  bool    `json:"discounted"`
}

type LoyaltyAction struct {
	ActionType    string                 `json:"action_type"`
	Points        int                    `json:"points"`
	Stamps        int                    `json:"stamps"`
	RewardID      string                 `json:"reward_id"`
	Reference     string                 `json:"reference"`
	ExtraData     map[string]interface{} `json:"extra_data"`
}

type ProcessingResult struct {
	EventID        string                 `json:"event_id"`
	ProcessedAt    time.Time              `json:"processed_at"`
	Success        bool                   `json:"success"`
	Error          string                 `json:"error,omitempty"`
	PointsEarned   int                    `json:"points_earned"`
	StampsEarned   int                    `json:"stamps_earned"`
	RewardsTriggered []RewardTriggered    `json:"rewards_triggered"`
	Actions        []string               `json:"actions"`
}

type RewardTriggered struct {
	RewardID     string    `json:"reward_id"`
	RewardType   string    `json:"reward_type"`
	RewardValue  string    `json:"reward_value"`
	Description  string    `json:"description"`
	TriggeredAt  time.Time `json:"triggered_at"`
}