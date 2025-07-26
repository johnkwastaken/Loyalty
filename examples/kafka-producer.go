package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
)

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

func main() {
	kafkaBrokers := "localhost:9092"
	
	writer := &kafka.Writer{
		Addr:     kafka.TCP(kafkaBrokers),
		Balancer: &kafka.LeastBytes{},
	}
	defer writer.Close()

	// Example 1: POS Transaction Event
	posEvent := createPOSTransactionEvent()
	if err := publishEvent(writer, "brand123.pos.transaction", posEvent); err != nil {
		log.Printf("Failed to publish POS event: %v", err)
	} else {
		log.Println("Successfully published POS transaction event")
	}

	// Example 2: Manual Loyalty Action Event
	loyaltyEvent := createLoyaltyActionEvent()
	if err := publishEvent(writer, "brand123.loyalty.action", loyaltyEvent); err != nil {
		log.Printf("Failed to publish loyalty event: %v", err)
	} else {
		log.Println("Successfully published loyalty action event")
	}

	log.Println("Done publishing test events")
}

func createPOSTransactionEvent() BaseEvent {
	transaction := POSTransaction{
		TransactionID: fmt.Sprintf("txn_%d", time.Now().Unix()),
		Amount:        25.50,
		Items: []LineItem{
			{
				SKU:        "COFFEE001",
				Name:       "Large Coffee",
				Quantity:   1,
				UnitPrice:  4.50,
				TotalPrice: 4.50,
				Category:   "beverages",
			},
			{
				SKU:        "MUFFIN001",
				Name:       "Blueberry Muffin",
				Quantity:   1,
				UnitPrice:  3.00,
				TotalPrice: 3.00,
				Category:   "pastries",
			},
		},
		PaymentMethod: "credit_card",
		ReceiptNumber: fmt.Sprintf("RCP%d", time.Now().Unix()),
		Cashier:       "emp_456",
	}

	payload := make(map[string]interface{})
	jsonData, _ := json.Marshal(transaction)
	json.Unmarshal(jsonData, &payload)

	return BaseEvent{
		EventID:    fmt.Sprintf("evt_%d", time.Now().UnixNano()),
		EventType:  "pos.transaction",
		OrgID:      "brand123",
		LocationID: "store001",
		CustomerID: "cust_789",
		Timestamp:  time.Now(),
		Payload:    payload,
	}
}

func createLoyaltyActionEvent() BaseEvent {
	action := map[string]interface{}{
		"action_type": "manual_points",
		"points":      100,
		"stamps":      0,
		"reference":   "birthday_bonus",
		"extra_data": map[string]interface{}{
			"reason": "birthday bonus points",
			"admin":  "manager_001",
		},
	}

	return BaseEvent{
		EventID:    fmt.Sprintf("evt_%d", time.Now().UnixNano()),
		EventType:  "loyalty.action",
		OrgID:      "brand123",
		LocationID: "store001",
		CustomerID: "cust_789",
		Timestamp:  time.Now(),
		Payload:    action,
	}
}

func publishEvent(writer *kafka.Writer, topic string, event BaseEvent) error {
	eventJSON, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	message := kafka.Message{
		Topic: topic,
		Key:   []byte(event.CustomerID),
		Value: eventJSON,
		Time:  time.Now(),
	}

	return writer.WriteMessages(context.Background(), message)
}