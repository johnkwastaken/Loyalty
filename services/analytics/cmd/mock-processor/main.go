package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/loyalty/analytics/internal/mock"
	"github.com/loyalty/analytics/internal/models"
	"github.com/loyalty/analytics/internal/rfm"
	"github.com/loyalty/analytics/internal/storage"
	"github.com/loyalty/analytics/internal/tiers"
)

func main() {
	mockKafkaURL := os.Getenv("MOCK_KAFKA_URL")
	if mockKafkaURL == "" {
		mockKafkaURL = "localhost:9093"
	}

	mongoURL := os.Getenv("MONGO_URL")
	if mongoURL == "" {
		mongoURL = "mongodb://admin:password@localhost:27017/analytics?authSource=admin"
	}

	log.Printf("🚀 Starting Mock Analytics Processor")
	log.Printf("📡 Mock Kafka: %s", mockKafkaURL)
	log.Printf("🗄️  MongoDB: %s", mongoURL)

	// Initialize storage
	mongoStorage, err := storage.NewMongoStorage(mongoURL, "analytics")
	if err != nil {
		log.Fatalf("Failed to create MongoDB storage: %v", err)
	}
	defer mongoStorage.Close()

	// Initialize processors
	rfmStorage := rfm.NewRFMStorage(mongoStorage)
	rfmCalculator := rfm.NewRFMCalculator(rfmStorage)

	tierStorage := tiers.NewTierStorage(mongoStorage.GetClient(), mongoStorage.GetDatabase())
	tierCalculator := tiers.NewTierCalculator(tierStorage)

	// Connect to mock Kafka
	client := mock.NewMockKafkaClient(mockKafkaURL)
	if err := client.Connect("*.pos.transaction"); err != nil {
		log.Fatalf("Failed to connect to mock Kafka: %v", err)
	}
	defer client.Close()

	// Setup graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("🛑 Shutting down mock analytics processor...")
		cancel()
	}()

	log.Println("📊 Analytics processor started, waiting for events...")

	// Process events
	eventCount := 0
	log.Println("🎯 Starting event processing loop...")
	
	for {
		select {
		case <-ctx.Done():
			log.Printf("✅ Processed %d events total", eventCount)
			return
		default:
			log.Printf("🔍 Waiting for events... (processed so far: %d)", eventCount)
			event, err := client.ReadEvent()
			if err != nil {
				log.Printf("⏰ Timeout waiting for event: %v", err)
				continue // Timeout, keep listening
			}

			eventCount++
			log.Printf("📨 Received event %d: %s from %s (Amount: %v)", 
				eventCount, event.EventType, event.CustomerID, event.Payload["amount"])

			if err := processEvent(ctx, event, rfmCalculator, tierCalculator); err != nil {
				log.Printf("❌ Error processing event %d: %v", eventCount, err)
			} else {
				log.Printf("✅ Event %d processed successfully - stored in MongoDB", eventCount)
			}
		}
	}
}

func processEvent(ctx context.Context, event models.BaseEvent, rfmCalc *rfm.RFMCalculator, tierCalc *tiers.TierCalculator) error {
	log.Printf("🔍 Processing event type: %s for customer: %s in org: %s", 
		event.EventType, event.CustomerID, event.OrgID)
	
	if event.EventType != "pos.transaction" {
		log.Printf("⚠️  Skipping non-transaction event: %s", event.EventType)
		return nil
	}

	// Parse transaction
	var transaction models.POSTransaction
	transactionData, err := json.Marshal(event.Payload)
	if err != nil {
		log.Printf("❌ Failed to marshal transaction payload: %v", err)
		return err
	}

	if err := json.Unmarshal(transactionData, &transaction); err != nil {
		log.Printf("❌ Failed to unmarshal transaction: %v", err)
		return err
	}
	
	log.Printf("💰 Transaction parsed: Amount=$%.2f, ID=%s", 
		transaction.Amount, transaction.TransactionID)

	// Create customer activity for RFM
	activity := models.CustomerActivity{
		OrgID:             event.OrgID,
		LocationID:        event.LocationID,
		CustomerID:        event.CustomerID,
		TransactionDate:   event.Timestamp,
		Amount:            transaction.Amount,
		FirstTransaction:  event.Timestamp,
		LastTransaction:   event.Timestamp,
		TotalTransactions: 1,
		TotalSpent:        transaction.Amount,
	}

	// Process RFM
	log.Printf("📊 Starting RFM calculation for %s...", event.CustomerID)
	if err := rfmCalc.ProcessCustomerTransaction(ctx, activity); err != nil {
		log.Printf("❌ RFM calculation failed for %s: %v", event.CustomerID, err)
	} else {
		log.Printf("✅ RFM calculated and stored for %s: $%.2f", event.CustomerID, transaction.Amount)
	}

	// Create customer metrics for tier calculation
	metrics := tiers.CustomerMetrics{
		OrgID:             event.OrgID,
		LocationID:        event.LocationID,
		CustomerID:        event.CustomerID,
		TotalSpent:        transaction.Amount,
		TotalVisits:       1,
		SpentThisYear:     transaction.Amount,
		VisitsThisYear:    1,
		SpentThisMonth:    transaction.Amount,
		VisitsThisMonth:   1,
		LastTransaction:   event.Timestamp,
		TransactionAmount: transaction.Amount,
	}

	// Process tier calculation
	log.Printf("🏆 Starting tier calculation for %s...", event.CustomerID)
	if err := tierCalc.ProcessCustomerMetrics(ctx, metrics); err != nil {
		log.Printf("❌ Tier calculation failed for %s: %v", event.CustomerID, err)
	} else {
		log.Printf("✅ Tier calculated and stored for %s: $%.2f", event.CustomerID, transaction.Amount)
	}

	return nil
}