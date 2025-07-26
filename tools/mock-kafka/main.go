package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/spf13/cobra"
)

var (
	port       int
	orgID      string
	customerID string
	interval   time.Duration
	eventCount int
)

type MockKafkaServer struct {
	consumers map[string][]*websocket.Conn
	mu        sync.RWMutex
	upgrader  websocket.Upgrader
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

func NewMockKafkaServer() *MockKafkaServer {
	return &MockKafkaServer{
		consumers: make(map[string][]*websocket.Conn),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}
}

func (s *MockKafkaServer) handleConsumer(w http.ResponseWriter, r *http.Request) {
	topic := r.URL.Query().Get("topic")
	if topic == "" {
		http.Error(w, "topic parameter required", http.StatusBadRequest)
		return
	}

	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return
	}
	defer conn.Close()

	s.mu.Lock()
	s.consumers[topic] = append(s.consumers[topic], conn)
	s.mu.Unlock()

	log.Printf("Consumer connected to topic: %s", topic)

	// Keep connection alive
	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			break
		}
	}

	// Remove consumer on disconnect
	s.mu.Lock()
	consumers := s.consumers[topic]
	for i, c := range consumers {
		if c == conn {
			s.consumers[topic] = append(consumers[:i], consumers[i+1:]...)
			break
		}
	}
	s.mu.Unlock()

	log.Printf("Consumer disconnected from topic: %s", topic)
}

func (s *MockKafkaServer) publishEvent(topic string, event BaseEvent) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	eventJSON, _ := json.Marshal(event)
	totalConsumers := 0
	
	// Find matching consumers for topic patterns
	for subscribedTopic, consumers := range s.consumers {
		if s.topicMatches(subscribedTopic, topic) {
			for _, conn := range consumers {
				if err := conn.WriteMessage(websocket.TextMessage, eventJSON); err != nil {
					log.Printf("Failed to send message to consumer: %v", err)
				} else {
					totalConsumers++
				}
			}
		}
	}

	if totalConsumers == 0 {
		log.Printf("📤 Event published to topic %s (no consumers)", topic)
	} else {
		log.Printf("📤 Event published to topic %s (%d consumers)", topic, totalConsumers)
	}
}

func (s *MockKafkaServer) topicMatches(pattern, topic string) bool {
	// Handle wildcard patterns like *.pos.transaction
	if pattern == "*" {
		return true
	}
	
	// Simple wildcard matching for *.pos.transaction pattern
	if len(pattern) > 0 && pattern[0] == '*' {
		suffix := pattern[1:] // Remove the *
		return len(topic) >= len(suffix) && topic[len(topic)-len(suffix):] == suffix
	}
	
	// Exact match
	return pattern == topic
}

func (s *MockKafkaServer) handlePublish(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST method required", http.StatusMethodNotAllowed)
		return
	}

	topic := r.URL.Query().Get("topic")
	if topic == "" {
		http.Error(w, "topic parameter required", http.StatusBadRequest)
		return
	}

	var event BaseEvent
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		http.Error(w, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
		return
	}

	s.publishEvent(topic, event)
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "published",
		"topic":  topic,
		"event_id": event.EventID,
	})
}

func (s *MockKafkaServer) handleStatus(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	topicStats := make(map[string]int)
	for topic, consumers := range s.consumers {
		topicStats[topic] = len(consumers)
	}
	s.mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":     "running",
		"topics":     topicStats,
		"timestamp":  time.Now(),
	})
}

func startServer(cmd *cobra.Command, args []string) {
	server := NewMockKafkaServer()

	http.HandleFunc("/consumer", server.handleConsumer)
	http.HandleFunc("/publish", server.handlePublish)
	http.HandleFunc("/status", server.handleStatus)

	log.Printf("🚀 Mock Kafka Server starting on port %d", port)
	log.Printf("📊 Endpoints:")
	log.Printf("   - WebSocket: ws://localhost:%d/consumer?topic=<topic>", port)
	log.Printf("   - Publish: POST http://localhost:%d/publish?topic=<topic>", port)
	log.Printf("   - Status: GET http://localhost:%d/status", port)

	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func publishEvents(cmd *cobra.Command, args []string) {
	baseURL := fmt.Sprintf("http://localhost:%d/publish", port)
	
	log.Printf("📡 Publishing %d events every %v", eventCount, interval)
	
	for i := 0; i < eventCount; i++ {
		event := generateRandomEvent()
		topic := fmt.Sprintf("%s.pos.transaction", event.OrgID)
		
		if err := publishEventToServer(baseURL, topic, event); err != nil {
			log.Printf("❌ Failed to publish event %d: %v", i+1, err)
		} else {
			log.Printf("✅ Published event %d: $%.2f for %s", i+1, 
				event.Payload["amount"], event.CustomerID)
		}
		
		if i < eventCount-1 {
			time.Sleep(interval)
		}
	}
}

func streamEvents(cmd *cobra.Command, args []string) {
	baseURL := fmt.Sprintf("http://localhost:%d/publish", port)
	
	log.Printf("🌊 Starting continuous event stream every %v", interval)
	log.Printf("Press Ctrl+C to stop...")

	eventNum := 0
	for {
		eventNum++
		event := generateRandomEvent()
		topic := fmt.Sprintf("%s.pos.transaction", event.OrgID)
		
		if err := publishEventToServer(baseURL, topic, event); err != nil {
			log.Printf("❌ Failed to publish event %d: %v", eventNum, err)
		} else {
			log.Printf("📨 Event %d: %s spent $%.2f at %s", eventNum,
				event.CustomerID, event.Payload["amount"], event.LocationID)
		}
		
		time.Sleep(interval)
	}
}

func generateRandomEvent() BaseEvent {
	customers := []string{"alice", "bob", "charlie", "diana", "cust_005"}
	locations := []string{"downtown", "westside", "eastmall"}
	
	if customerID != "" {
		customers = []string{customerID}
	}
	
	selectedCustomer := customers[rand.Intn(len(customers))]
	selectedLocation := locations[rand.Intn(len(locations))]
	selectedOrg := orgID
	if selectedOrg == "" {
		selectedOrg = "brand123"
	}
	
	amount := 5.0 + rand.Float64()*95.0 // $5-$100
	
	transaction := map[string]interface{}{
		"transaction_id": fmt.Sprintf("txn_%d", time.Now().UnixNano()),
		"amount":         amount,
		"items": []map[string]interface{}{
			{
				"sku":        "COFFEE001",
				"name":       "Large Coffee",
				"quantity":   1,
				"unit_price": amount * 0.6,
				"total_price": amount * 0.6,
				"category":   "beverages",
			},
			{
				"sku":        "MUFFIN001", 
				"name":       "Blueberry Muffin",
				"quantity":   1,
				"unit_price": amount * 0.4,
				"total_price": amount * 0.4,
				"category":   "pastries",
			},
		},
		"payment_method": "credit_card",
		"receipt_number": fmt.Sprintf("RCP%d", rand.Intn(999999)),
		"cashier":        fmt.Sprintf("emp_%d", rand.Intn(10)),
	}

	return BaseEvent{
		EventID:    fmt.Sprintf("evt_%d", time.Now().UnixNano()),
		EventType:  "pos.transaction",
		OrgID:      selectedOrg,
		LocationID: selectedLocation,
		CustomerID: selectedCustomer,
		Timestamp:  time.Now(),
		Payload:    transaction,
	}
}

func publishEventToServer(baseURL, topic string, event BaseEvent) error {
	url := fmt.Sprintf("%s?topic=%s", baseURL, topic)
	
	eventJSON, err := json.Marshal(event)
	if err != nil {
		return err
	}

	resp, err := http.Post(url, "application/json", 
		bytes.NewBuffer(eventJSON))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned status %d", resp.StatusCode)
	}

	return nil
}

func main() {
	var rootCmd = &cobra.Command{
		Use:   "mock-kafka",
		Short: "Mock Kafka server for loyalty platform testing",
		Long:  "A mock Kafka server that simulates event streaming for the loyalty platform",
	}

	var serverCmd = &cobra.Command{
		Use:   "server",
		Short: "Start the mock Kafka server",
		Run:   startServer,
	}

	var publishCmd = &cobra.Command{
		Use:   "publish",
		Short: "Publish test events to the mock server",
		Run:   publishEvents,
	}

	var streamCmd = &cobra.Command{
		Use:   "stream",
		Short: "Start continuous event streaming",
		Run:   streamEvents,
	}

	// Global flags
	rootCmd.PersistentFlags().IntVar(&port, "port", 9093, "Server port")
	rootCmd.PersistentFlags().StringVar(&orgID, "org", "brand123", "Organization ID")
	rootCmd.PersistentFlags().StringVar(&customerID, "customer", "", "Specific customer ID")

	// Publish flags
	publishCmd.Flags().IntVar(&eventCount, "count", 10, "Number of events to publish")
	publishCmd.Flags().DurationVar(&interval, "interval", 1*time.Second, "Interval between events")

	// Stream flags
	streamCmd.Flags().DurationVar(&interval, "interval", 2*time.Second, "Interval between events")

	rootCmd.AddCommand(serverCmd, publishCmd, streamCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}