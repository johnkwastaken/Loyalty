package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/spf13/cobra"
)

var (
	brokers    string
	orgID      string
	locationID string
	customerID string
	count      int
	interval   time.Duration
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

func main() {
	var rootCmd = &cobra.Command{
		Use:   "kafka-cli",
		Short: "Loyalty Platform Kafka Event Generator",
		Long:  "A CLI tool to generate test events for the loyalty platform Kafka streams",
	}

	rootCmd.PersistentFlags().StringVar(&brokers, "brokers", "localhost:9092", "Kafka broker addresses")
	rootCmd.PersistentFlags().StringVar(&orgID, "org", "brand123", "Organization ID")
	rootCmd.PersistentFlags().StringVar(&locationID, "location", "store001", "Location ID")
	rootCmd.PersistentFlags().StringVar(&customerID, "customer", "", "Customer ID (random if empty)")
	rootCmd.PersistentFlags().IntVar(&count, "count", 1, "Number of events to generate")
	rootCmd.PersistentFlags().DurationVar(&interval, "interval", 1*time.Second, "Interval between events")

	var posCmd = &cobra.Command{
		Use:   "pos",
		Short: "Generate POS transaction events",
		Long:  "Generate point-of-sale transaction events",
		Run:   generatePOSEvents,
	}

	var loyaltyCmd = &cobra.Command{
		Use:   "loyalty",
		Short: "Generate loyalty action events",
		Long:  "Generate manual loyalty action events",
		Run:   generateLoyaltyEvents,
	}

	var customerCmd = &cobra.Command{
		Use:   "customer",
		Short: "Generate customer update events",
		Long:  "Generate customer profile update events",
		Run:   generateCustomerEvents,
	}

	var streamCmd = &cobra.Command{
		Use:   "stream",
		Short: "Generate continuous stream of events",
		Long:  "Generate a continuous stream of mixed events for testing",
		Run:   generateEventStream,
	}

	var benchmarkCmd = &cobra.Command{
		Use:   "benchmark",
		Short: "Run benchmark test",
		Long:  "Generate high-volume events for performance testing",
		Run:   runBenchmark,
	}

	rootCmd.AddCommand(posCmd, loyaltyCmd, customerCmd, streamCmd, benchmarkCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func generatePOSEvents(cmd *cobra.Command, args []string) {
	writer := createKafkaWriter()
	defer writer.Close()

	for i := 0; i < count; i++ {
		event := createPOSEvent()
		topic := fmt.Sprintf("%s.pos.transaction", orgID)
		
		if err := publishEvent(writer, topic, event); err != nil {
			log.Printf("Failed to publish POS event %d: %v", i+1, err)
		} else {
			fmt.Printf("✓ Published POS transaction: $%.2f [%s]\n", 
				event.Payload["amount"], event.CustomerID)
		}

		if i < count-1 {
			time.Sleep(interval)
		}
	}
}

func generateLoyaltyEvents(cmd *cobra.Command, args []string) {
	writer := createKafkaWriter()
	defer writer.Close()

	for i := 0; i < count; i++ {
		event := createLoyaltyEvent()
		topic := fmt.Sprintf("%s.loyalty.action", orgID)
		
		if err := publishEvent(writer, topic, event); err != nil {
			log.Printf("Failed to publish loyalty event %d: %v", i+1, err)
		} else {
			action := event.Payload["action_type"]
			points := event.Payload["points"]
			fmt.Printf("✓ Published loyalty action: %s (%v points) [%s]\n", 
				action, points, event.CustomerID)
		}

		if i < count-1 {
			time.Sleep(interval)
		}
	}
}

func generateCustomerEvents(cmd *cobra.Command, args []string) {
	writer := createKafkaWriter()
	defer writer.Close()

	for i := 0; i < count; i++ {
		event := createCustomerEvent()
		topic := fmt.Sprintf("%s.customer.updated", orgID)
		
		if err := publishEvent(writer, topic, event); err != nil {
			log.Printf("Failed to publish customer event %d: %v", i+1, err)
		} else {
			fmt.Printf("✓ Published customer update [%s]\n", event.CustomerID)
		}

		if i < count-1 {
			time.Sleep(interval)
		}
	}
}

func generateEventStream(cmd *cobra.Command, args []string) {
	writer := createKafkaWriter()
	defer writer.Close()

	fmt.Printf("🚀 Starting event stream: %d events every %v\n", count, interval)
	fmt.Println("Press Ctrl+C to stop...")

	eventCount := 0
	for {
		eventType := rand.Intn(3)
		var event BaseEvent
		var topic string

		switch eventType {
		case 0:
			event = createPOSEvent()
			topic = fmt.Sprintf("%s.pos.transaction", orgID)
		case 1:
			event = createLoyaltyEvent()
			topic = fmt.Sprintf("%s.loyalty.action", orgID)
		case 2:
			event = createCustomerEvent()
			topic = fmt.Sprintf("%s.customer.updated", orgID)
		}

		if err := publishEvent(writer, topic, event); err != nil {
			log.Printf("Failed to publish event: %v", err)
		} else {
			eventCount++
			fmt.Printf("📨 Event %d: %s [%s]\n", eventCount, event.EventType, event.CustomerID)
		}

		time.Sleep(interval)
	}
}

func runBenchmark(cmd *cobra.Command, args []string) {
	writer := createKafkaWriter()
	defer writer.Close()

	fmt.Printf("🏃 Running benchmark: %d events\n", count)
	start := time.Now()

	for i := 0; i < count; i++ {
		event := createPOSEvent()
		topic := fmt.Sprintf("%s.pos.transaction", orgID)
		
		if err := publishEvent(writer, topic, event); err != nil {
			log.Printf("Failed to publish event %d: %v", i+1, err)
		}

		if (i+1)%1000 == 0 {
			fmt.Printf("📊 Sent %d events...\n", i+1)
		}
	}

	duration := time.Since(start)
	eventsPerSec := float64(count) / duration.Seconds()
	
	fmt.Printf("✅ Benchmark complete!\n")
	fmt.Printf("📈 Events: %d\n", count)
	fmt.Printf("⏱️  Duration: %v\n", duration)
	fmt.Printf("🚀 Events/sec: %.2f\n", eventsPerSec)
}

func createKafkaWriter() *kafka.Writer {
	return &kafka.Writer{
		Addr:     kafka.TCP(brokers),
		Balancer: &kafka.LeastBytes{},
	}
}

func createPOSEvent() BaseEvent {
	cust := getCustomerID()
	amount := 5.0 + rand.Float64()*95.0 // $5-$100
	
	items := generateItems(amount)
	
	transaction := map[string]interface{}{
		"transaction_id": fmt.Sprintf("txn_%d", time.Now().UnixNano()),
		"amount":         amount,
		"items":          items,
		"payment_method": randomPaymentMethod(),
		"receipt_number": fmt.Sprintf("RCP%d", rand.Intn(999999)),
		"cashier":        fmt.Sprintf("emp_%d", rand.Intn(100)),
	}

	return BaseEvent{
		EventID:    fmt.Sprintf("evt_%d", time.Now().UnixNano()),
		EventType:  "pos.transaction",
		OrgID:      orgID,
		LocationID: locationID,
		CustomerID: cust,
		Timestamp:  time.Now(),
		Payload:    transaction,
	}
}

func createLoyaltyEvent() BaseEvent {
	cust := getCustomerID()
	
	actions := []string{"manual_points", "bonus_stamps", "birthday_bonus", "referral_bonus"}
	actionType := actions[rand.Intn(len(actions))]
	
	var points, stamps int
	switch actionType {
	case "manual_points":
		points = 50 + rand.Intn(200)
	case "bonus_stamps":
		stamps = 1 + rand.Intn(5)
	case "birthday_bonus":
		points = 100
	case "referral_bonus":
		points = 250
	}

	action := map[string]interface{}{
		"action_type": actionType,
		"points":      points,
		"stamps":      stamps,
		"reference":   fmt.Sprintf("ref_%d", time.Now().UnixNano()),
		"extra_data": map[string]interface{}{
			"reason": actionType,
			"admin":  fmt.Sprintf("admin_%d", rand.Intn(10)),
		},
	}

	return BaseEvent{
		EventID:    fmt.Sprintf("evt_%d", time.Now().UnixNano()),
		EventType:  "loyalty.action",
		OrgID:      orgID,
		LocationID: locationID,
		CustomerID: cust,
		Timestamp:  time.Now(),
		Payload:    action,
	}
}

func createCustomerEvent() BaseEvent {
	cust := getCustomerID()
	
	updates := map[string]interface{}{
		"tier":        randomTier(),
		"preferences": map[string]interface{}{
			"email_marketing": rand.Intn(2) == 1,
			"sms_marketing":   rand.Intn(2) == 1,
			"language":        randomLanguage(),
		},
		"updated_at": time.Now(),
	}

	return BaseEvent{
		EventID:    fmt.Sprintf("evt_%d", time.Now().UnixNano()),
		EventType:  "customer.updated",
		OrgID:      orgID,
		LocationID: locationID,
		CustomerID: cust,
		Timestamp:  time.Now(),
		Payload:    updates,
	}
}

func generateItems(totalAmount float64) []map[string]interface{} {
	products := []map[string]interface{}{
		{"sku": "COFFEE001", "name": "Large Coffee", "category": "beverages", "base_price": 4.50},
		{"sku": "COFFEE002", "name": "Medium Coffee", "category": "beverages", "base_price": 3.50},
		{"sku": "MUFFIN001", "name": "Blueberry Muffin", "category": "pastries", "base_price": 3.00},
		{"sku": "SAND001", "name": "Turkey Sandwich", "category": "food", "base_price": 8.50},
		{"sku": "SALAD001", "name": "Caesar Salad", "category": "food", "base_price": 7.00},
		{"sku": "COOKIE001", "name": "Chocolate Chip Cookie", "category": "pastries", "base_price": 2.50},
	}

	var items []map[string]interface{}
	remaining := totalAmount
	numItems := 1 + rand.Intn(4) // 1-4 items

	for i := 0; i < numItems; i++ {
		product := products[rand.Intn(len(products))]
		quantity := 1 + rand.Intn(3) // 1-3 quantity
		
		var unitPrice float64
		if i == numItems-1 {
			// Last item - use remaining amount
			unitPrice = remaining / float64(quantity)
		} else {
			basePrice := product["base_price"].(float64)
			unitPrice = basePrice * (0.8 + rand.Float64()*0.4) // ±20% variance
		}
		
		totalPrice := unitPrice * float64(quantity)
		remaining -= totalPrice

		items = append(items, map[string]interface{}{
			"sku":        product["sku"],
			"name":       product["name"],
			"category":   product["category"],
			"quantity":   quantity,
			"unit_price": unitPrice,
			"total_price": totalPrice,
		})
	}

	return items
}

func getCustomerID() string {
	if customerID != "" {
		return customerID
	}
	return fmt.Sprintf("cust_%d", rand.Intn(1000))
}

func randomPaymentMethod() string {
	methods := []string{"credit_card", "debit_card", "cash", "mobile_pay", "gift_card"}
	return methods[rand.Intn(len(methods))]
}

func randomTier() string {
	tiers := []string{"bronze", "silver", "gold", "platinum", "diamond"}
	return tiers[rand.Intn(len(tiers))]
}

func randomLanguage() string {
	languages := []string{"en", "es", "fr", "de", "it"}
	return languages[rand.Intn(len(languages))]
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