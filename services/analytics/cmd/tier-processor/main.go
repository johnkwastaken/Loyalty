package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/loyalty/analytics/internal/storage"
	"github.com/loyalty/analytics/internal/tiers"
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
	TransactionID string    `json:"transaction_id"`
	Amount        float64   `json:"amount"`
	Timestamp     time.Time `json:"timestamp"`
}

func main() {
	kafkaBrokers := os.Getenv("KAFKA_BROKERS")
	if kafkaBrokers == "" {
		kafkaBrokers = "localhost:9092"
	}

	mongoURL := os.Getenv("MONGO_URL")
	if mongoURL == "" {
		mongoURL = "mongodb://admin:password@localhost:27017/analytics?authSource=admin"
	}

	consumerGroupID := os.Getenv("CONSUMER_GROUP_ID")
	if consumerGroupID == "" {
		consumerGroupID = "tier-processor"
	}

	mongoStorage, err := storage.NewMongoStorage(mongoURL, "analytics")
	if err != nil {
		log.Fatalf("Failed to create MongoDB storage: %v", err)
	}
	defer mongoStorage.Close()

	tierStorage := tiers.NewTierStorage(mongoStorage.GetClient(), mongoStorage.GetDatabase())
	calculator := tiers.NewTierCalculator(tierStorage)

	brokerList := strings.Split(kafkaBrokers, ",")
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     brokerList,
		GroupID:     consumerGroupID,
		Topic:       "",
		MinBytes:    10e3,
		MaxBytes:    10e6,
		MaxWait:     1 * time.Second,
		StartOffset: kafka.LastOffset,
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Shutting down tier processor...")
		cancel()
	}()

	go scheduledRecalculation(ctx, calculator)

	log.Printf("Starting tier processor with brokers: %s", kafkaBrokers)
	log.Printf("Consumer group: %s", consumerGroupID)
	log.Printf("MongoDB URL: %s", mongoURL)

	for {
		select {
		case <-ctx.Done():
			log.Println("Context cancelled, stopping tier processor")
			if err := reader.Close(); err != nil {
				log.Printf("Error closing reader: %v", err)
			}
			return
		default:
			message, err := reader.FetchMessage(ctx)
			if err != nil {
				if err == context.Canceled {
					continue
				}
				log.Printf("Error fetching message: %v", err)
				time.Sleep(5 * time.Second)
				continue
			}

			if shouldProcessMessage(string(message.Topic)) {
				if err := processMessage(ctx, message, calculator, tierStorage); err != nil {
					log.Printf("Error processing message: %v", err)
				}
			}

			if err := reader.CommitMessages(ctx, message); err != nil {
				log.Printf("Error committing message: %v", err)
			}
		}
	}
}

func shouldProcessMessage(topic string) bool {
	patterns := []string{
		".pos.transaction",
		".loyalty.action",
	}
	
	for _, pattern := range patterns {
		if strings.Contains(topic, pattern) {
			return true
		}
	}
	return false
}

func processMessage(ctx context.Context, message kafka.Message, calculator *tiers.TierCalculator, storage *tiers.TierStorage) error {
	var event BaseEvent
	if err := json.Unmarshal(message.Value, &event); err != nil {
		return err
	}

	if event.EventType != "pos.transaction" {
		return nil
	}

	var transaction POSTransaction
	transactionData, err := json.Marshal(event.Payload)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(transactionData, &transaction); err != nil {
		return err
	}

	existingTier, err := storage.GetCustomerTier(ctx, event.OrgID, event.CustomerID)
	if err != nil {
		log.Printf("Customer tier not found, will create new: %v", err)
		existingTier = &tiers.CustomerTier{
			OrgID:       event.OrgID,
			CustomerID:  event.CustomerID,
			CurrentTier: "Bronze",
			TierSince:   time.Now(),
		}
	}

	now := time.Now()
	yearStart := time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location())
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	metrics := tiers.CustomerMetrics{
		OrgID:             event.OrgID,
		CustomerID:        event.CustomerID,
		TotalSpent:        existingTier.TotalSpent + transaction.Amount,
		TotalVisits:       existingTier.TotalVisits + 1,
		SpentThisYear:     existingTier.SpentThisYear + transaction.Amount,
		VisitsThisYear:    existingTier.VisitsThisYear + 1,
		SpentThisMonth:    existingTier.SpentThisMonth + transaction.Amount,
		VisitsThisMonth:   existingTier.VisitsThisMonth + 1,
		LastTransaction:   transaction.Timestamp,
		TransactionAmount: transaction.Amount,
	}

	if existingTier.LastTransaction.Before(yearStart) {
		metrics.SpentThisYear = transaction.Amount
		metrics.VisitsThisYear = 1
	}

	if existingTier.LastTransaction.Before(monthStart) {
		metrics.SpentThisMonth = transaction.Amount
		metrics.VisitsThisMonth = 1
	}

	if err := calculator.ProcessCustomerMetrics(ctx, metrics); err != nil {
		return err
	}

	log.Printf("Updated tier for customer %s: $%.2f total, %d visits",
		event.CustomerID, metrics.TotalSpent, metrics.TotalVisits)

	return nil
}

func scheduledRecalculation(ctx context.Context, calculator *tiers.TierCalculator) {
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			log.Println("Starting scheduled tier recalculation...")
		}
	}
}