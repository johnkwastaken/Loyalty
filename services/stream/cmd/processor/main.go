package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/loyalty/stream/internal/processor"
	"github.com/segmentio/kafka-go"
)

func main() {
	kafkaBrokers := os.Getenv("KAFKA_BROKERS")
	if kafkaBrokers == "" {
		kafkaBrokers = "localhost:9092"
	}

	ledgerURL := os.Getenv("LEDGER_URL")
	if ledgerURL == "" {
		ledgerURL = "http://localhost:8001"
	}

	membershipURL := os.Getenv("MEMBERSHIP_URL")
	if membershipURL == "" {
		membershipURL = "http://localhost:8002"
	}

	consumerGroupID := os.Getenv("CONSUMER_GROUP_ID")
	if consumerGroupID == "" {
		consumerGroupID = "loyalty-stream-processor"
	}

	topics := []string{
		"*.pos.transaction",
		"*.loyalty.action",
		"*.customer.updated",
	}

	eventProcessor := processor.NewEventProcessor(ledgerURL, membershipURL)

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
		log.Println("Shutting down stream processor...")
		cancel()
	}()

	log.Printf("Starting stream processor with brokers: %s", kafkaBrokers)
	log.Printf("Consumer group: %s", consumerGroupID)
	log.Printf("Ledger URL: %s", ledgerURL)
	log.Printf("Membership URL: %s", membershipURL)

	for {
		select {
		case <-ctx.Done():
			log.Println("Context cancelled, stopping processor")
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

			if shouldProcessTopic(string(message.Topic), topics) {
				result, err := eventProcessor.ProcessEvent(ctx, message)
				if err != nil {
					log.Printf("Error processing event %s: %v", 
						getEventID(message.Value), err)
				} else if result != nil {
					if result.Success {
						log.Printf("Successfully processed event %s: %d points, %d stamps, %d rewards",
							result.EventID, result.PointsEarned, result.StampsEarned, len(result.RewardsTriggered))
					} else {
						log.Printf("Failed to process event %s: %s", 
							result.EventID, result.Error)
					}
				}
			}

			if err := reader.CommitMessages(ctx, message); err != nil {
				log.Printf("Error committing message: %v", err)
			}
		}
	}
}

func shouldProcessTopic(topic string, patterns []string) bool {
	for _, pattern := range patterns {
		if strings.Contains(pattern, "*") {
			parts := strings.Split(pattern, ".")
			topicParts := strings.Split(topic, ".")
			
			if len(parts) == len(topicParts) {
				match := true
				for i, part := range parts {
					if part != "*" && part != topicParts[i] {
						match = false
						break
					}
				}
				if match {
					return true
				}
			}
		} else if topic == pattern {
			return true
		}
	}
	return false
}

func getEventID(messageValue []byte) string {
	return string(messageValue[:min(50, len(messageValue))])
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}