package mock

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
	"github.com/loyalty/analytics/internal/models"
)

type MockKafkaClient struct {
	serverURL string
	conn      *websocket.Conn
	eventChan chan models.BaseEvent
}

func NewMockKafkaClient(serverURL string) *MockKafkaClient {
	return &MockKafkaClient{
		serverURL: serverURL,
		eventChan: make(chan models.BaseEvent, 100),
	}
}

func (c *MockKafkaClient) Connect(topic string) error {
	u := url.URL{Scheme: "ws", Host: c.serverURL, Path: "/consumer"}
	q := u.Query()
	q.Set("topic", topic)
	u.RawQuery = q.Encode()

	var err error
	c.conn, _, err = websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return fmt.Errorf("failed to connect to mock Kafka: %w", err)
	}

	log.Printf("🔗 Connected to mock Kafka topic: %s", topic)

	// Start reading messages
	go c.readMessages()

	return nil
}

func (c *MockKafkaClient) readMessages() {
	defer c.conn.Close()

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			log.Printf("Error reading message: %v", err)
			return
		}

		var event models.BaseEvent
		if err := json.Unmarshal(message, &event); err != nil {
			log.Printf("Error unmarshaling event: %v", err)
			continue
		}

		c.eventChan <- event
	}
}

func (c *MockKafkaClient) ReadEvent() (models.BaseEvent, error) {
	select {
	case event := <-c.eventChan:
		return event, nil
	case <-time.After(5 * time.Second):
		return models.BaseEvent{}, fmt.Errorf("timeout reading event")
	}
}

func (c *MockKafkaClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}