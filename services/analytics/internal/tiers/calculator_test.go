package tiers

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockTierStorage is a mock implementation of the tier storage
type MockTierStorage struct {
	mock.Mock
}

func (m *MockTierStorage) GetTierConfig(ctx context.Context, orgID string) (*OrgTierConfig, error) {
	args := m.Called(ctx, orgID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*OrgTierConfig), args.Error(1)
}

func (m *MockTierStorage) SaveTierConfig(ctx context.Context, config OrgTierConfig) error {
	args := m.Called(ctx, config)
	return args.Error(0)
}

func (m *MockTierStorage) GetCustomerTier(ctx context.Context, orgID, customerID string) (*CustomerTier, error) {
	args := m.Called(ctx, orgID, customerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*CustomerTier), args.Error(1)
}

func (m *MockTierStorage) SaveCustomerTier(ctx context.Context, tier CustomerTier) error {
	args := m.Called(ctx, tier)
	return args.Error(0)
}

func (m *MockTierStorage) SaveTierUpgrade(ctx context.Context, upgrade TierUpgrade) error {
	args := m.Called(ctx, upgrade)
	return args.Error(0)
}

func (m *MockTierStorage) GetTierUpgrades(ctx context.Context, orgID string, unnotifiedOnly bool) ([]TierUpgrade, error) {
	args := m.Called(ctx, orgID, unnotifiedOnly)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]TierUpgrade), args.Error(1)
}

func (m *MockTierStorage) MarkUpgradeNotified(ctx context.Context, upgradeID string) error {
	args := m.Called(ctx, upgradeID)
	return args.Error(0)
}

func (m *MockTierStorage) GetCustomersByTier(ctx context.Context, orgID, tierName string) ([]CustomerTier, error) {
	args := m.Called(ctx, orgID, tierName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]CustomerTier), args.Error(1)
}

func (m *MockTierStorage) GetAllCustomerTiers(ctx context.Context, orgID string) ([]CustomerTier, error) {
	args := m.Called(ctx, orgID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]CustomerTier), args.Error(1)
}

// Test setup helper
func setupTestCalculator() (*TierCalculator, *MockTierStorage) {
	mockStorage := &MockTierStorage{}
	calculator := NewTierCalculator(mockStorage)
	return calculator, mockStorage
}

// Test NewTierCalculator
func TestNewTierCalculator(t *testing.T) {
	mockStorage := &MockTierStorage{}
	calculator := NewTierCalculator(mockStorage)
	
	assert.NotNil(t, calculator)
	assert.Equal(t, mockStorage, calculator.storage)
}

// Test ProcessCustomerMetrics
func TestProcessCustomerMetrics_Success(t *testing.T) {
	calculator, mockStorage := setupTestCalculator()
	ctx := context.Background()
	
	// Test data
	metrics := CustomerMetrics{
		OrgID:            "test_org",
		LocationID:       "test_location",
		CustomerID:       "test_customer",
		TotalSpent:       5000.0,
		TotalVisits:      50,
		SpentThisYear:    2000.0,
		VisitsThisYear:   20,
		SpentThisMonth:   500.0,
		VisitsThisMonth:  5,
		LastTransaction:  time.Now(),
		TransactionAmount: 100.0,
	}
	
	// Mock tier config
	tierConfig := &OrgTierConfig{
		OrgID:     "test_org",
		TierRules: GetDefaultTierRules(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	
	// Mock current tier
	currentTier := &CustomerTier{
		OrgID:       "test_org",
		LocationID:  "test_location",
		CustomerID:  "test_customer",
		CurrentTier: "Bronze",
		TierSince:   time.Now().AddDate(0, 0, -30),
	}
	
	// Setup expectations
	mockStorage.On("GetTierConfig", ctx, "test_org").Return(tierConfig, nil)
	mockStorage.On("GetCustomerTier", ctx, "test_org", "test_customer").Return(currentTier, nil)
	mockStorage.On("SaveCustomerTier", ctx, mock.AnythingOfType("CustomerTier")).Return(nil)
	mockStorage.On("SaveTierUpgrade", ctx, mock.AnythingOfType("TierUpgrade")).Return(nil)
	
	// Process metrics
	err := calculator.ProcessCustomerMetrics(ctx, metrics)
	
	// Assertions
	assert.NoError(t, err)
	
	// Verify the saved customer tier
	mockStorage.AssertCalled(t, "SaveCustomerTier", ctx, mock.MatchedBy(func(tier CustomerTier) bool {
		return tier.OrgID == "test_org" &&
			tier.CustomerID == "test_customer" &&
			tier.LocationID == "test_location" &&
			tier.TotalSpent == 5000.0 &&
			tier.TotalVisits == 50 &&
			tier.SpentThisYear == 2000.0 &&
			tier.VisitsThisYear == 20
	}))
	
	mockStorage.AssertExpectations(t)
}

func TestProcessCustomerMetrics_NoTierConfig(t *testing.T) {
	calculator, mockStorage := setupTestCalculator()
	ctx := context.Background()
	
	// Test data
	metrics := CustomerMetrics{
		OrgID:            "test_org",
		LocationID:       "test_location",
		CustomerID:       "test_customer",
		TotalSpent:       1000.0,
		TotalVisits:      10,
		SpentThisYear:    500.0,
		VisitsThisYear:   5,
		LastTransaction:  time.Now(),
		TransactionAmount: 50.0,
	}
	
	// Mock current tier
	currentTier := &CustomerTier{
		OrgID:       "test_org",
		LocationID:  "test_location",
		CustomerID:  "test_customer",
		CurrentTier: "Bronze",
		TierSince:   time.Now().AddDate(0, 0, -30),
	}
	
	// Setup expectations - no tier config found
	mockStorage.On("GetTierConfig", ctx, "test_org").Return(nil, assert.AnError)
	mockStorage.On("SaveTierConfig", ctx, mock.AnythingOfType("OrgTierConfig")).Return(nil)
	mockStorage.On("GetCustomerTier", ctx, "test_org", "test_customer").Return(currentTier, nil)
	mockStorage.On("SaveCustomerTier", ctx, mock.AnythingOfType("CustomerTier")).Return(nil)
	mockStorage.On("SaveTierUpgrade", ctx, mock.AnythingOfType("TierUpgrade")).Return(nil)
	
	// Process metrics
	err := calculator.ProcessCustomerMetrics(ctx, metrics)
	
	// Assertions
	assert.NoError(t, err)
	
	// Verify default tier config was saved
	mockStorage.AssertCalled(t, "SaveTierConfig", ctx, mock.MatchedBy(func(config OrgTierConfig) bool {
		return config.OrgID == "test_org" && len(config.TierRules) > 0
	}))
	
	mockStorage.AssertExpectations(t)
}

func TestProcessCustomerMetrics_NoCurrentTier(t *testing.T) {
	calculator, mockStorage := setupTestCalculator()
	ctx := context.Background()
	
	// Test data
	metrics := CustomerMetrics{
		OrgID:            "test_org",
		LocationID:       "test_location",
		CustomerID:       "new_customer",
		TotalSpent:       100.0,
		TotalVisits:      2,
		SpentThisYear:    100.0,
		VisitsThisYear:   2,
		LastTransaction:  time.Now(),
		TransactionAmount: 50.0,
	}
	
	// Mock tier config
	tierConfig := &OrgTierConfig{
		OrgID:     "test_org",
		TierRules: GetDefaultTierRules(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	
	// Setup expectations - no current tier found
	mockStorage.On("GetTierConfig", ctx, "test_org").Return(tierConfig, nil)
	mockStorage.On("GetCustomerTier", ctx, "test_org", "new_customer").Return(nil, assert.AnError)
	mockStorage.On("SaveCustomerTier", ctx, mock.AnythingOfType("CustomerTier")).Return(nil)
	
	// Process metrics
	err := calculator.ProcessCustomerMetrics(ctx, metrics)
	
	// Assertions
	assert.NoError(t, err)
	
	// Verify new customer tier was created
	mockStorage.AssertCalled(t, "SaveCustomerTier", ctx, mock.MatchedBy(func(tier CustomerTier) bool {
		return tier.OrgID == "test_org" &&
			tier.CustomerID == "new_customer" &&
			tier.CurrentTier == "Bronze" // Default tier for new customers
	}))
	
	mockStorage.AssertExpectations(t)
}

func TestProcessCustomerMetrics_TierUpgrade(t *testing.T) {
	calculator, mockStorage := setupTestCalculator()
	ctx := context.Background()
	
	// Test data - high spending customer
	metrics := CustomerMetrics{
		OrgID:            "test_org",
		LocationID:       "test_location",
		CustomerID:       "test_customer",
		TotalSpent:       15000.0,
		TotalVisits:      100,
		SpentThisYear:    8000.0,
		VisitsThisYear:   40,
		LastTransaction:  time.Now(),
		TransactionAmount: 200.0,
	}
	
	// Mock tier config
	tierConfig := &OrgTierConfig{
		OrgID:     "test_org",
		TierRules: GetDefaultTierRules(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	
	// Mock current tier - Bronze
	currentTier := &CustomerTier{
		OrgID:       "test_org",
		LocationID:  "test_location",
		CustomerID:  "test_customer",
		CurrentTier: "Bronze",
		TierSince:   time.Now().AddDate(0, 0, -30),
	}
	
	// Setup expectations
	mockStorage.On("GetTierConfig", ctx, "test_org").Return(tierConfig, nil)
	mockStorage.On("GetCustomerTier", ctx, "test_org", "test_customer").Return(currentTier, nil)
	mockStorage.On("SaveCustomerTier", ctx, mock.AnythingOfType("CustomerTier")).Return(nil)
	mockStorage.On("SaveTierUpgrade", ctx, mock.AnythingOfType("TierUpgrade")).Return(nil)
	
	// Process metrics
	err := calculator.ProcessCustomerMetrics(ctx, metrics)
	
	// Assertions
	assert.NoError(t, err)
	
	// Verify tier upgrade was saved
	mockStorage.AssertCalled(t, "SaveTierUpgrade", ctx, mock.MatchedBy(func(upgrade TierUpgrade) bool {
		return upgrade.OrgID == "test_org" &&
			upgrade.CustomerID == "test_customer" &&
			upgrade.FromTier == "Bronze" &&
			upgrade.ToTier == "Diamond" && // Should upgrade to Diamond based on spending
			upgrade.TriggeredBy == "transaction" &&
			upgrade.TriggerValue == 200.0
	}))
	
	mockStorage.AssertExpectations(t)
}

func TestProcessCustomerMetrics_SaveTierError(t *testing.T) {
	calculator, mockStorage := setupTestCalculator()
	ctx := context.Background()
	
	// Test data
	metrics := CustomerMetrics{
		OrgID:            "test_org",
		LocationID:       "test_location",
		CustomerID:       "test_customer",
		TotalSpent:       1000.0,
		TotalVisits:      10,
		SpentThisYear:    500.0,
		VisitsThisYear:   5,
		LastTransaction:  time.Now(),
		TransactionAmount: 50.0,
	}
	
	// Mock tier config
	tierConfig := &OrgTierConfig{
		OrgID:     "test_org",
		TierRules: GetDefaultTierRules(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	
	// Mock current tier
	currentTier := &CustomerTier{
		OrgID:       "test_org",
		LocationID:  "test_location",
		CustomerID:  "test_customer",
		CurrentTier: "Bronze",
		TierSince:   time.Now().AddDate(0, 0, -30),
	}
	
	// Setup expectations
	mockStorage.On("GetTierConfig", ctx, "test_org").Return(tierConfig, nil)
	mockStorage.On("GetCustomerTier", ctx, "test_org", "test_customer").Return(currentTier, nil)
	mockStorage.On("SaveCustomerTier", ctx, mock.AnythingOfType("CustomerTier")).Return(assert.AnError)
	
	// Process metrics
	err := calculator.ProcessCustomerMetrics(ctx, metrics)
	
	// Assertions
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to save customer tier")
	
	mockStorage.AssertExpectations(t)
}

// Test calculateTier
func TestCalculateTier(t *testing.T) {
	calculator, _ := setupTestCalculator()
	
	rules := GetDefaultTierRules()
	
	tests := []struct {
		name    string
		metrics CustomerMetrics
		expectedTier string
	}{
		{
			name: "bronze tier",
			metrics: CustomerMetrics{
				TotalSpent:      100.0,
				TotalVisits:     2,
				SpentThisYear:   50.0,
				VisitsThisYear:  1,
			},
			expectedTier: "Bronze",
		},
		{
			name: "silver tier",
			metrics: CustomerMetrics{
				TotalSpent:      300.0,
				TotalVisits:     6,
				SpentThisYear:   150.0,
				VisitsThisYear:  4,
			},
			expectedTier: "Silver",
		},
		{
			name: "gold tier",
			metrics: CustomerMetrics{
				TotalSpent:      1000.0,
				TotalVisits:     20,
				SpentThisYear:   400.0,
				VisitsThisYear:  10,
			},
			expectedTier: "Gold",
		},
		{
			name: "platinum tier",
			metrics: CustomerMetrics{
				TotalSpent:      2500.0,
				TotalVisits:     35,
				SpentThisYear:   1000.0,
				VisitsThisYear:  20,
			},
			expectedTier: "Platinum",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tier := calculator.calculateTier(tt.metrics, rules)
			assert.Equal(t, tt.expectedTier, tier.Name)
		})
	}
}

// Test meetsRequirements
func TestMeetsRequirements(t *testing.T) {
	calculator, _ := setupTestCalculator()
	
	rule := TierRule{
		Name:             "Test Tier",
		Level:            2,
		MinSpentLifetime: 1000.0,
		MinSpentYear:     500.0,
		MinVisitsLifetime: 10,
		MinVisitsYear:    5,
		PointsMultiplier: 1.5,
		Benefits:         []string{"Free shipping"},
	}
	
	tests := []struct {
		name    string
		metrics CustomerMetrics
		expected bool
	}{
		{
			name: "meets all requirements",
			metrics: CustomerMetrics{
				TotalSpent:      1500.0,
				TotalVisits:     15,
				SpentThisYear:   600.0,
				VisitsThisYear:  6,
			},
			expected: true,
		},
		{
			name: "does not meet lifetime spent",
			metrics: CustomerMetrics{
				TotalSpent:      800.0,
				TotalVisits:     15,
				SpentThisYear:   600.0,
				VisitsThisYear:  6,
			},
			expected: false,
		},
		{
			name: "does not meet year spent",
			metrics: CustomerMetrics{
				TotalSpent:      1500.0,
				TotalVisits:     15,
				SpentThisYear:   400.0,
				VisitsThisYear:  6,
			},
			expected: false,
		},
		{
			name: "does not meet lifetime visits",
			metrics: CustomerMetrics{
				TotalSpent:      1500.0,
				TotalVisits:     8,
				SpentThisYear:   600.0,
				VisitsThisYear:  6,
			},
			expected: false,
		},
		{
			name: "does not meet year visits",
			metrics: CustomerMetrics{
				TotalSpent:      1500.0,
				TotalVisits:     15,
				SpentThisYear:   600.0,
				VisitsThisYear:  3,
			},
			expected: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculator.meetsRequirements(tt.metrics, rule)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Test updateCustomerTier
func TestUpdateCustomerTier(t *testing.T) {
	calculator, _ := setupTestCalculator()
	
	now := time.Now()
	currentTier := &CustomerTier{
		OrgID:       "test_org",
		LocationID:  "test_location",
		CustomerID:  "test_customer",
		CurrentTier: "Bronze",
		TierSince:   now.AddDate(0, 0, -30),
		PreviousTier: "",
	}
	
	newTier := TierRule{
		Name:             "Silver",
		Level:            2,
		MinSpentLifetime: 3000.0,
		MinSpentYear:     1500.0,
		MinVisitsLifetime: 25,
		MinVisitsYear:    12,
		PointsMultiplier: 1.5,
		Benefits:         []string{"Free shipping", "Priority support"},
	}
	
	metrics := CustomerMetrics{
		OrgID:            "test_org",
		LocationID:       "test_location",
		CustomerID:       "test_customer",
		TotalSpent:       3500.0,
		TotalVisits:      30,
		SpentThisYear:    1800.0,
		VisitsThisYear:   15,
		SpentThisMonth:   300.0,
		VisitsThisMonth:  3,
		LastTransaction:  now,
	}
	
	updated := calculator.updateCustomerTier(currentTier, newTier, metrics)
	
	// Assertions
	assert.Equal(t, "Silver", updated.CurrentTier)
	assert.Equal(t, "Bronze", updated.PreviousTier)
	assert.WithinDuration(t, now, updated.TierSince, time.Second)
	assert.Equal(t, 3500.0, updated.TotalSpent)
	assert.Equal(t, 30, updated.TotalVisits)
	assert.Equal(t, 1800.0, updated.SpentThisYear)
	assert.Equal(t, 15, updated.VisitsThisYear)
	assert.Equal(t, 300.0, updated.SpentThisMonth)
	assert.Equal(t, 3, updated.VisitsThisMonth)
	assert.WithinDuration(t, now, updated.LastTransaction, time.Second)
	assert.Equal(t, 1.5, updated.PointsMultiplier)
	assert.Equal(t, []string{"Free shipping", "Priority support"}, updated.Benefits)
	assert.NotEmpty(t, updated.NextTier)
	assert.GreaterOrEqual(t, updated.ProgressToNext, 0.0)
	assert.LessOrEqual(t, updated.ProgressToNext, 1.0)
}

// Test calculateNextTierProgress
func TestCalculateNextTierProgress(t *testing.T) {
	calculator, _ := setupTestCalculator()
	
	currentTier := TierRule{
		Name:             "Silver",
		Level:            2,
		MinSpentLifetime: 3000.0,
		MinSpentYear:     1500.0,
		MinVisitsLifetime: 25,
		MinVisitsYear:    12,
		PointsMultiplier: 1.5,
		Benefits:         []string{"Free shipping"},
	}
	
	tests := []struct {
		name           string
		metrics        CustomerMetrics
		expectedTier   string
		expectedProgress float64
	}{
		{
			name: "halfway to gold",
			metrics: CustomerMetrics{
				SpentThisYear:  150.0, // Half of Gold requirement (300)
				VisitsThisYear: 4,     // Half of Gold requirement (8)
			},
			expectedTier:   "Gold",
			expectedProgress: 0.5,
		},
		{
			name: "almost at gold",
			metrics: CustomerMetrics{
				SpentThisYear:  270.0, // 90% of Gold requirement (300)
				VisitsThisYear: 7,     // 90% of Gold requirement (8)
			},
			expectedTier:   "Gold",
			expectedProgress: 0.9,
		},
		{
			name: "exceeds gold requirements",
			metrics: CustomerMetrics{
				SpentThisYear:  400.0, // Exceeds Gold requirement (300)
				VisitsThisYear: 10,    // Exceeds Gold requirement (8)
			},
			expectedTier:   "Gold",
			expectedProgress: 1.0,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nextTier, progress := calculator.calculateNextTierProgress(currentTier, tt.metrics)
			assert.Equal(t, tt.expectedTier, nextTier)
			assert.InDelta(t, tt.expectedProgress, progress, 0.1) // Allow small floating point differences
		})
	}
}

// Test GetTierUpgrades
func TestGetTierUpgrades(t *testing.T) {
	calculator, mockStorage := setupTestCalculator()
	ctx := context.Background()
	
	expectedUpgrades := []TierUpgrade{
		{
			OrgID:        "test_org",
			CustomerID:   "cust_1",
			FromTier:     "Bronze",
			ToTier:       "Silver",
			TriggeredBy:  "transaction",
			TriggerValue: 100.0,
			UpgradedAt:   time.Now(),
			Notified:     false,
		},
		{
			OrgID:        "test_org",
			CustomerID:   "cust_2",
			FromTier:     "Silver",
			ToTier:       "Gold",
			TriggeredBy:  "transaction",
			TriggerValue: 200.0,
			UpgradedAt:   time.Now(),
			Notified:     true,
		},
	}
	
	// Setup expectations
	mockStorage.On("GetTierUpgrades", ctx, "test_org", false).Return(expectedUpgrades, nil)
	
	// Get upgrades
	upgrades, err := calculator.GetTierUpgrades(ctx, "test_org", false)
	
	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, expectedUpgrades, upgrades)
	
	mockStorage.AssertExpectations(t)
}

// Test MarkUpgradeNotified
func TestMarkUpgradeNotified(t *testing.T) {
	calculator, mockStorage := setupTestCalculator()
	ctx := context.Background()
	
	// Setup expectations
	mockStorage.On("MarkUpgradeNotified", ctx, "upgrade_123").Return(nil)
	
	// Mark upgrade as notified
	err := calculator.MarkUpgradeNotified(ctx, "upgrade_123")
	
	// Assertions
	assert.NoError(t, err)
	
	mockStorage.AssertExpectations(t)
}

// Test GetCustomersByTier
func TestGetCustomersByTier(t *testing.T) {
	calculator, mockStorage := setupTestCalculator()
	ctx := context.Background()
	
	expectedCustomers := []CustomerTier{
		{
			OrgID:       "test_org",
			CustomerID:  "cust_1",
			CurrentTier: "Gold",
			TierSince:   time.Now(),
		},
		{
			OrgID:       "test_org",
			CustomerID:  "cust_2",
			CurrentTier: "Gold",
			TierSince:   time.Now(),
		},
	}
	
	// Setup expectations
	mockStorage.On("GetCustomersByTier", ctx, "test_org", "Gold").Return(expectedCustomers, nil)
	
	// Get customers by tier
	customers, err := calculator.GetCustomersByTier(ctx, "test_org", "Gold")
	
	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, expectedCustomers, customers)
	
	mockStorage.AssertExpectations(t)
}

// Test RecalculateAllTiers
func TestRecalculateAllTiers(t *testing.T) {
	calculator, mockStorage := setupTestCalculator()
	ctx := context.Background()
	
	existingCustomers := []CustomerTier{
		{
			OrgID:       "test_org",
			CustomerID:  "cust_1",
			CurrentTier: "Bronze",
			TotalSpent:  1000.0,
			TotalVisits: 10,
		},
		{
			OrgID:       "test_org",
			CustomerID:  "cust_2",
			CurrentTier: "Silver",
			TotalSpent:  5000.0,
			TotalVisits: 30,
		},
	}
	
	// Setup expectations
	mockStorage.On("GetAllCustomerTiers", ctx, "test_org").Return(existingCustomers, nil)
	mockStorage.On("GetTierConfig", ctx, "test_org").Return(&OrgTierConfig{
		OrgID:     "test_org",
		TierRules: GetDefaultTierRules(),
	}, nil)
	mockStorage.On("GetCustomerTier", ctx, "test_org", "cust_1").Return(&existingCustomers[0], nil)
	mockStorage.On("GetCustomerTier", ctx, "test_org", "cust_2").Return(&existingCustomers[1], nil)
	mockStorage.On("SaveCustomerTier", ctx, mock.AnythingOfType("CustomerTier")).Return(nil).Times(2)
	mockStorage.On("SaveTierUpgrade", ctx, mock.AnythingOfType("TierUpgrade")).Return(nil).Times(1)
	
	// Recalculate all tiers
	err := calculator.RecalculateAllTiers(ctx, "test_org")
	
	// Assertions
	assert.NoError(t, err)
	
	mockStorage.AssertExpectations(t)
} 