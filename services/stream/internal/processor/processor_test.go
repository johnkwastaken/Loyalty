package processor

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/loyalty/stream/internal/clients"
	"github.com/loyalty/stream/internal/models"
	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockLedgerClient is a mock implementation of the ledger client
type MockLedgerClient struct {
	mock.Mock
}

func (m *MockLedgerClient) CreatePointsTransfer(orgID, customerID string, points int, reference string) (*clients.TransferResponse, error) {
	args := m.Called(orgID, customerID, points, reference)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*clients.TransferResponse), args.Error(1)
}

func (m *MockLedgerClient) CreateStampsTransfer(orgID, customerID string, stamps int, reference string) (*clients.TransferResponse, error) {
	args := m.Called(orgID, customerID, stamps, reference)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*clients.TransferResponse), args.Error(1)
}

// MockMembershipClient is a mock implementation of the membership client
type MockMembershipClient struct {
	mock.Mock
}

func (m *MockMembershipClient) GetCustomer(customerID string) (*clients.Customer, error) {
	args := m.Called(customerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*clients.Customer), args.Error(1)
}

func (m *MockMembershipClient) GetOrganization(orgID string) (*clients.Organization, error) {
	args := m.Called(orgID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*clients.Organization), args.Error(1)
}

// Test setup helper
func setupTestProcessor() (*EventProcessor, *MockLedgerClient, *MockMembershipClient) {
	processor := &EventProcessor{}
	
	// Create mock clients
	mockLedgerClient := &MockLedgerClient{}
	mockMembershipClient := &MockMembershipClient{}
	
	// Inject mock clients
	processor.ledgerClient = mockLedgerClient
	processor.membershipClient = mockMembershipClient
	
	return processor, mockLedgerClient, mockMembershipClient
}

// Test ProcessEvent
func TestProcessEvent_POSTransaction_Success(t *testing.T) {
	processor, mockLedgerClient, mockMembershipClient := setupTestProcessor()
	
	// Test data
	transaction := models.POSTransaction{
		TransactionID: "txn_123",
		Amount:        50.0,
	}
	
	event := models.BaseEvent{
		EventID:     "evt_123",
		EventType:   models.EventTypePOSTransaction,
		OrgID:       "test_org",
		CustomerID:  "test_customer",
		Timestamp:   time.Now(),
		Payload:     map[string]interface{}{
			"transaction_id": transaction.TransactionID,
			"amount":         transaction.Amount,
		},
	}
	
	eventData, _ := json.Marshal(event)
	message := kafka.Message{
		Value: eventData,
	}
	
	// Mock responses
	mockCustomer := &clients.Customer{
		CustomerID: "test_customer",
		OrgID:      "test_org",
		Email:      "test@example.com",
		Status:     "active",
	}
	
	mockOrg := &clients.Organization{
		OrgID: "test_org",
		Settings: clients.OrgSettings{
			PointsPerDollar:    2.0,
			StampsPerVisit:     1,
			RewardThresholds:   []clients.RewardThreshold{},
		},
	}
	
	mockTransferResponse := &clients.TransferResponse{
		TransferID: "transfer_123",
		Status:     "success",
	}
	
	// Setup expectations
	mockMembershipClient.On("GetCustomer", "test_customer").Return(mockCustomer, nil)
	mockMembershipClient.On("GetOrganization", "test_org").Return(mockOrg, nil)
	mockLedgerClient.On("CreatePointsTransfer", "test_org", "test_customer", 100, "pos_transaction_txn_123").Return(mockTransferResponse, nil)
	mockLedgerClient.On("CreateStampsTransfer", "test_org", "test_customer", 1, "pos_transaction_txn_123").Return(mockTransferResponse, nil)
	
	// Process event
	result, err := processor.ProcessEvent(context.Background(), message)
	
	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "evt_123", result.EventID)
	assert.True(t, result.Success)
	assert.Equal(t, 100, result.PointsEarned)
	assert.Equal(t, 1, result.StampsEarned)
	assert.Len(t, result.Actions, 2)
	assert.Contains(t, result.Actions[0], "awarded 100 points")
	assert.Contains(t, result.Actions[1], "awarded 1 stamps")
	
	mockLedgerClient.AssertExpectations(t)
	mockMembershipClient.AssertExpectations(t)
}

func TestProcessEvent_POSTransaction_ZeroPointsPerDollar(t *testing.T) {
	processor, mockLedgerClient, mockMembershipClient := setupTestProcessor()
	
	// Test data
	transaction := models.POSTransaction{
		TransactionID: "txn_123",
		Amount:        50.0,
	}
	
	event := models.BaseEvent{
		EventID:     "evt_123",
		EventType:   models.EventTypePOSTransaction,
		OrgID:       "test_org",
		CustomerID:  "test_customer",
		Timestamp:   time.Now(),
		Payload:     map[string]interface{}{
			"transaction_id": transaction.TransactionID,
			"amount":         transaction.Amount,
		},
	}
	
	eventData, _ := json.Marshal(event)
	message := kafka.Message{
		Value: eventData,
	}
	
	// Mock responses
	mockCustomer := &clients.Customer{
		CustomerID: "test_customer",
		OrgID:      "test_org",
		Email:      "test@example.com",
		Status:     "active",
	}
	
	mockOrg := &clients.Organization{
		OrgID: "test_org",
		Settings: clients.OrgSettings{
			PointsPerDollar:    0.0, // No points per dollar
			StampsPerVisit:     1,
			RewardThresholds:   []clients.RewardThreshold{},
		},
	}
	
	mockTransferResponse := &clients.TransferResponse{
		TransferID: "transfer_123",
		Status:     "success",
	}
	
	// Setup expectations
	mockMembershipClient.On("GetCustomer", "test_customer").Return(mockCustomer, nil)
	mockMembershipClient.On("GetOrganization", "test_org").Return(mockOrg, nil)
	mockLedgerClient.On("CreateStampsTransfer", "test_org", "test_customer", 1, "pos_transaction_txn_123").Return(mockTransferResponse, nil)
	
	// Process event
	result, err := processor.ProcessEvent(context.Background(), message)
	
	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "evt_123", result.EventID)
	assert.True(t, result.Success)
	assert.Equal(t, 0, result.PointsEarned)
	assert.Equal(t, 1, result.StampsEarned)
	assert.Len(t, result.Actions, 1)
	assert.Contains(t, result.Actions[0], "awarded 1 stamps")
	
	mockLedgerClient.AssertExpectations(t)
	mockMembershipClient.AssertExpectations(t)
}

func TestProcessEvent_POSTransaction_CustomerNotFound(t *testing.T) {
	processor, _, mockMembershipClient := setupTestProcessor()
	
	// Test data
	transaction := models.POSTransaction{
		TransactionID: "txn_123",
		Amount:        50.0,
	}
	
	event := models.BaseEvent{
		EventID:     "evt_123",
		EventType:   models.EventTypePOSTransaction,
		OrgID:       "test_org",
		CustomerID:  "nonexistent_customer",
		Timestamp:   time.Now(),
		Payload:     map[string]interface{}{"transaction_id": transaction.TransactionID, "amount": transaction.Amount},
	}
	
	eventData, _ := json.Marshal(event)
	message := kafka.Message{
		Value: eventData,
	}
	
	// Mock customer not found
	mockMembershipClient.On("GetCustomer", "nonexistent_customer").Return(nil, assert.AnError)
	
	// Process event
	result, err := processor.ProcessEvent(context.Background(), message)
	
	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "evt_123", result.EventID)
	assert.False(t, result.Success)
	assert.Contains(t, result.Error, "failed to get customer")
	
	mockMembershipClient.AssertExpectations(t)
}

func TestProcessEvent_POSTransaction_OrganizationNotFound(t *testing.T) {
	processor, _, mockMembershipClient := setupTestProcessor()
	
	// Test data
	transaction := models.POSTransaction{
		TransactionID: "txn_123",
		Amount:        50.0,
	}
	
	event := models.BaseEvent{
		EventID:     "evt_123",
		EventType:   models.EventTypePOSTransaction,
		OrgID:       "nonexistent_org",
		CustomerID:  "test_customer",
		Timestamp:   time.Now(),
		Payload:     map[string]interface{}{"transaction_id": transaction.TransactionID, "amount": transaction.Amount},
	}
	
	eventData, _ := json.Marshal(event)
	message := kafka.Message{
		Value: eventData,
	}
	
	// Mock responses
	mockCustomer := &clients.Customer{
		CustomerID: "test_customer",
		OrgID:      "test_org",
		Email:      "test@example.com",
		Status:     "active",
	}
	
	// Setup expectations
	mockMembershipClient.On("GetCustomer", "test_customer").Return(mockCustomer, nil)
	mockMembershipClient.On("GetOrganization", "nonexistent_org").Return(nil, assert.AnError)
	
	// Process event
	result, err := processor.ProcessEvent(context.Background(), message)
	
	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "evt_123", result.EventID)
	assert.False(t, result.Success)
	assert.Contains(t, result.Error, "failed to get organization")
	
	mockMembershipClient.AssertExpectations(t)
}

func TestProcessEvent_POSTransaction_PointsTransferError(t *testing.T) {
	processor, mockLedgerClient, mockMembershipClient := setupTestProcessor()
	
	// Test data
	transaction := models.POSTransaction{
		TransactionID: "txn_123",
		Amount:        50.0,
	}
	
	event := models.BaseEvent{
		EventID:     "evt_123",
		EventType:   models.EventTypePOSTransaction,
		OrgID:       "test_org",
		CustomerID:  "test_customer",
		Timestamp:   time.Now(),
		Payload:     map[string]interface{}{"transaction_id": transaction.TransactionID, "amount": transaction.Amount},
	}
	
	eventData, _ := json.Marshal(event)
	message := kafka.Message{
		Value: eventData,
	}
	
	// Mock responses
	mockCustomer := &clients.Customer{
		CustomerID: "test_customer",
		OrgID:      "test_org",
		Email:      "test@example.com",
		Status:     "active",
	}
	
	mockOrg := &clients.Organization{
		OrgID: "test_org",
		Settings: clients.OrgSettings{
			PointsPerDollar:    2.0,
			StampsPerVisit:     1,
			RewardThresholds:   []clients.RewardThreshold{},
		},
	}
	
	// Setup expectations
	mockMembershipClient.On("GetCustomer", "test_customer").Return(mockCustomer, nil)
	mockMembershipClient.On("GetOrganization", "test_org").Return(mockOrg, nil)
	mockLedgerClient.On("CreatePointsTransfer", "test_org", "test_customer", 100, "pos_transaction_txn_123").Return(nil, assert.AnError)
	
	// Process event
	result, err := processor.ProcessEvent(context.Background(), message)
	
	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "evt_123", result.EventID)
	assert.False(t, result.Success)
	assert.Contains(t, result.Error, "failed to create points transfer")
	
	mockLedgerClient.AssertExpectations(t)
	mockMembershipClient.AssertExpectations(t)
}

// Test LoyaltyAction processing
func TestProcessEvent_LoyaltyAction_ManualPoints_Success(t *testing.T) {
	processor, mockLedgerClient, _ := setupTestProcessor()
	
	// Test data
	action := models.LoyaltyAction{
		ActionType: "manual_points",
		Points:     50,
		Reference:  "manual_award",
	}
	
	event := models.BaseEvent{
		EventID:     "evt_123",
		EventType:   models.EventTypeLoyaltyAction,
		OrgID:       "test_org",
		CustomerID:  "test_customer",
		Timestamp:   time.Now(),
		Payload:     map[string]interface{}{"action_type": action.ActionType, "points": action.Points, "stamps": action.Stamps, "reference": action.Reference},
	}
	
	eventData, _ := json.Marshal(event)
	message := kafka.Message{
		Value: eventData,
	}
	
	// Mock responses
	mockTransferResponse := &clients.TransferResponse{
		TransferID: "transfer_123",
		Status:     "success",
	}
	
	// Setup expectations
	mockLedgerClient.On("CreatePointsTransfer", "test_org", "test_customer", 50, "manual_award").Return(mockTransferResponse, nil)
	
	// Process event
	result, err := processor.ProcessEvent(context.Background(), message)
	
	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "evt_123", result.EventID)
	assert.True(t, result.Success)
	assert.Equal(t, 50, result.PointsEarned)
	assert.Len(t, result.Actions, 1)
	assert.Contains(t, result.Actions[0], "manual award: 50 points")
	
	mockLedgerClient.AssertExpectations(t)
}

func TestProcessEvent_LoyaltyAction_BonusStamps_Success(t *testing.T) {
	processor, mockLedgerClient, _ := setupTestProcessor()
	
	// Test data
	action := models.LoyaltyAction{
		ActionType: "bonus_stamps",
		Stamps:     5,
		Reference:  "bonus_award",
	}
	
	event := models.BaseEvent{
		EventID:     "evt_123",
		EventType:   models.EventTypeLoyaltyAction,
		OrgID:       "test_org",
		CustomerID:  "test_customer",
		Timestamp:   time.Now(),
		Payload:     map[string]interface{}{"action_type": action.ActionType, "points": action.Points, "stamps": action.Stamps, "reference": action.Reference},
	}
	
	eventData, _ := json.Marshal(event)
	message := kafka.Message{
		Value: eventData,
	}
	
	// Mock responses
	mockTransferResponse := &clients.TransferResponse{
		TransferID: "transfer_123",
		Status:     "success",
	}
	
	// Setup expectations
	mockLedgerClient.On("CreateStampsTransfer", "test_org", "test_customer", 5, "bonus_award").Return(mockTransferResponse, nil)
	
	// Process event
	result, err := processor.ProcessEvent(context.Background(), message)
	
	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "evt_123", result.EventID)
	assert.True(t, result.Success)
	assert.Equal(t, 5, result.StampsEarned)
	assert.Len(t, result.Actions, 1)
	assert.Contains(t, result.Actions[0], "bonus stamps: 5")
	
	mockLedgerClient.AssertExpectations(t)
}

func TestProcessEvent_LoyaltyAction_UnknownActionType(t *testing.T) {
	processor, _, _ := setupTestProcessor()
	
	// Test data
	action := models.LoyaltyAction{
		ActionType: "unknown_action",
		Points:     50,
		Reference:  "test",
	}
	
	event := models.BaseEvent{
		EventID:     "evt_123",
		EventType:   models.EventTypeLoyaltyAction,
		OrgID:       "test_org",
		CustomerID:  "test_customer",
		Timestamp:   time.Now(),
		Payload:     map[string]interface{}{"action_type": action.ActionType, "points": action.Points, "stamps": action.Stamps, "reference": action.Reference},
	}
	
	eventData, _ := json.Marshal(event)
	message := kafka.Message{
		Value: eventData,
	}
	
	// Process event
	result, err := processor.ProcessEvent(context.Background(), message)
	
	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "evt_123", result.EventID)
	assert.False(t, result.Success)
	assert.Contains(t, result.Error, "unknown loyalty action type")
}

func TestProcessEvent_LoyaltyAction_ZeroPoints(t *testing.T) {
	processor, mockLedgerClient, _ := setupTestProcessor()
	
	// Test data
	action := models.LoyaltyAction{
		ActionType: "manual_points",
		Points:     0, // Zero points
		Reference:  "manual_award",
	}
	
	event := models.BaseEvent{
		EventID:     "evt_123",
		EventType:   models.EventTypeLoyaltyAction,
		OrgID:       "test_org",
		CustomerID:  "test_customer",
		Timestamp:   time.Now(),
		Payload:     map[string]interface{}{"action_type": action.ActionType, "points": action.Points, "stamps": action.Stamps, "reference": action.Reference},
	}
	
	eventData, _ := json.Marshal(event)
	message := kafka.Message{
		Value: eventData,
	}
	
	// Process event
	result, err := processor.ProcessEvent(context.Background(), message)
	
	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "evt_123", result.EventID)
	assert.True(t, result.Success)
	assert.Equal(t, 0, result.PointsEarned)
	assert.Len(t, result.Actions, 0)
	
	// Should not call ledger client for zero points
	mockLedgerClient.AssertNotCalled(t, "CreatePointsTransfer")
}

func TestProcessEvent_UnknownEventType(t *testing.T) {
	processor, _, _ := setupTestProcessor()
	
	// Test data
	event := models.BaseEvent{
		EventID:     "evt_123",
		EventType:   "unknown_event_type",
		OrgID:       "test_org",
		CustomerID:  "test_customer",
		Timestamp:   time.Now(),
		Payload:     map[string]interface{}{},
	}
	
	eventData, _ := json.Marshal(event)
	message := kafka.Message{
		Value: eventData,
	}
	
	// Process event
	result, err := processor.ProcessEvent(context.Background(), message)
	
	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "evt_123", result.EventID)
	assert.False(t, result.Success)
	assert.Contains(t, result.Error, "unknown event type")
}

func TestProcessEvent_InvalidJSON(t *testing.T) {
	processor, _, _ := setupTestProcessor()
	
	// Invalid JSON message
	message := kafka.Message{
		Value: []byte("invalid json"),
	}
	
	// Process event
	_, err := processor.ProcessEvent(context.Background(), message)
	
	// Assertions
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to unmarshal event")
}

// Test calculatePoints
func TestCalculatePoints(t *testing.T) {
	processor, _, _ := setupTestProcessor()
	
	tests := []struct {
		name            string
		amount          float64
		pointsPerDollar float64
		expected        int
	}{
		{"normal calculation", 50.0, 2.0, 100},
		{"zero amount", 0.0, 2.0, 0},
		{"zero points per dollar", 50.0, 0.0, 0},
		{"decimal result", 25.5, 1.5, 38}, // 25.5 * 1.5 = 38.25, floor = 38
		{"negative amount", -10.0, 2.0, -20},
		{"negative points per dollar", 50.0, -1.0, 0},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := processor.calculatePoints(tt.amount, tt.pointsPerDollar)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Test checkRewardThresholds
func TestCheckRewardThresholds(t *testing.T) {
	processor, _, _ := setupTestProcessor()
	
	thresholds := []clients.RewardThreshold{
		{
			Points:       100,
			Stamps:       0,
			RewardType:   "discount",
			RewardValue:  "10.0",
			Description:  "10% discount at 100 points",
		},
		{
			Points:       0,
			Stamps:       5,
			RewardType:   "free_item",
			RewardValue:  "0.0",
			Description:  "Free item at 5 stamps",
		},
		{
			Points:       200,
			Stamps:       10,
			RewardType:   "voucher",
			RewardValue:  "25.0",
			Description:  "$25 voucher at 200 points and 10 stamps",
		},
	}
	
	tests := []struct {
		name           string
		points         int
		stamps         int
		expectedCount  int
		expectedReward string
	}{
		{"no rewards", 50, 2, 0, ""},
		{"points threshold only", 150, 2, 1, "10% discount at 100 points"},
		{"stamps threshold only", 50, 7, 1, "Free item at 5 stamps"},
		{"both thresholds", 250, 12, 3, "10% discount at 100 points"},
		{"exact points match", 100, 0, 1, "10% discount at 100 points"},
		{"exact stamps match", 0, 5, 1, "Free item at 5 stamps"},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rewards := processor.checkRewardThresholds(thresholds, tt.points, tt.stamps)
			assert.Len(t, rewards, tt.expectedCount)
			
			if tt.expectedCount > 0 {
				found := false
				for _, reward := range rewards {
					if reward.Description == tt.expectedReward {
						found = true
						break
					}
				}
				assert.True(t, found, "Expected reward not found: %s", tt.expectedReward)
			}
		})
	}
}

// Test NewEventProcessor
func TestNewEventProcessor(t *testing.T) {
	ledgerURL := "http://localhost:8001"
	membershipURL := "http://localhost:8002"
	
	processor := NewEventProcessor(ledgerURL, membershipURL)
	
	assert.NotNil(t, processor)
	assert.NotNil(t, processor.ledgerClient)
	assert.NotNil(t, processor.membershipClient)
} 