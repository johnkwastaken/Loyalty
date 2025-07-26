package models

import "time"

type BaseEvent struct {
	EventID    string                 `json:"event_id"`
	EventType  string                 `json:"event_type"`
	OrgID      string                 `json:"org_id"`
	LocationID string                 `json:"location_id"`
	CustomerID string                 `json:"customer_id"`
	Timestamp  time.Time              `json:"timestamp"`
	Payload    map[string]interface{} `json:"payload"`
}

type POSTransaction struct {
	TransactionID string     `json:"transaction_id"`
	Amount        float64    `json:"amount"`
	Items         []LineItem `json:"items"`
	PaymentMethod string     `json:"payment_method"`
	ReceiptNumber string     `json:"receipt_number"`
	Cashier       string     `json:"cashier"`
}

type LineItem struct {
	SKU        string  `json:"sku"`
	Name       string  `json:"name"`
	Quantity   int     `json:"quantity"`
	UnitPrice  float64 `json:"unit_price"`
	TotalPrice float64 `json:"total_price"`
	Category   string  `json:"category"`
}