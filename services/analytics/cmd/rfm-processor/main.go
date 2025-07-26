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

	"github.com/loyalty/analytics/internal/models"
	"github.com/loyalty/analytics/internal/rfm"
	"github.com/loyalty/analytics/internal/storage"
	"github.com/segmentio/kafka-go"
)

type POSTransaction struct {
	TransactionID string    `json:"transaction_id"`
	Amount        float64   `json:"amount"`
	Timestamp     time.Time `json:"timestamp"`
}

type BaseEvent struct {
	EventID    string                 `json:"event_id"`
	EventType  string                 `json:"event_type"`
	OrgID      string                 `json:"org_id"`
	LocationID string                 `json:"location_id"`
	CustomerID string                 `json:"customer_id"`
	Timestamp  time.Time              `json:"timestamp"`
	Payload    map[string]interface{} `json:"payload"`
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
		consumerGroupID = "rfm-processor"
	}

	mongoStorage, err := storage.NewMongoStorage(mongoURL, "analytics")
	if err != nil {
		log.Fatalf("Failed to create MongoDB storage: %v", err)
	}
	defer mongoStorage.Close()

	rfmStorage := rfm.NewRFMStorage(mongoStorage)
	calculator := rfm.NewRFMCalculator(rfmStorage)

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
		log.Println("Shutting down RFM processor...")
		cancel()
	}()

	log.Printf("Starting RFM processor with brokers: %s", kafkaBrokers)
	log.Printf("Consumer group: %s", consumerGroupID)
	log.Printf("MongoDB URL: %s", mongoURL)

	for {
		select {
		case <-ctx.Done():
			log.Println("Context cancelled, stopping RFM processor")
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
				if err := processMessage(ctx, message, calculator, rfmStorage); err != nil {
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

func processMessage(ctx context.Context, message kafka.Message, calculator *rfm.RFMCalculator, storage *rfm.RFMStorage) error {
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

	existingActivity, err := getExistingActivity(ctx, storage, event.OrgID, event.CustomerID)
	if err != nil {
		log.Printf("Customer activity not found, creating new: %v", err)
		existingActivity = &models.CustomerActivity{
			OrgID:             event.OrgID,
			CustomerID:        event.CustomerID,
			FirstTransaction:  transaction.Timestamp,
			LastTransaction:   transaction.Timestamp,
			TotalTransactions: 0,
			TotalSpent:        0,
		}
	}

	activity := models.CustomerActivity{
		OrgID:             event.OrgID,
		CustomerID:        event.CustomerID,
		TransactionDate:   transaction.Timestamp,
		Amount:            transaction.Amount,
		FirstTransaction:  existingActivity.FirstTransaction,
		LastTransaction:   transaction.Timestamp,
		TotalTransactions: existingActivity.TotalTransactions + 1,
		TotalSpent:        existingActivity.TotalSpent + transaction.Amount,
	}

	if transaction.Timestamp.Before(existingActivity.FirstTransaction) {
		activity.FirstTransaction = transaction.Timestamp
	}

	if err := storage.UpdateCustomerActivity(ctx, activity); err != nil {
		return err
	}

	if err := calculator.ProcessCustomerTransaction(ctx, activity); err != nil {
		return err
	}

	log.Printf("Updated RFM for customer %s: %d transactions, $%.2f total",
		event.CustomerID, activity.TotalTransactions, activity.TotalSpent)

	return nil
}

func getExistingActivity(ctx context.Context, storage *rfm.RFMStorage, orgID, customerID string) (*models.CustomerActivity, error) {
	activities, err := storage.GetCustomerActivities(ctx, orgID)
	if err != nil {
		return nil, err
	}

	for _, activity := range activities {
		if activity.CustomerID == customerID {
			return &activity, nil
		}
	}

	return nil, nil
}