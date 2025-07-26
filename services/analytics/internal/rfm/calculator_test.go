package rfm

import (
	"context"
	"testing"
	"time"

	"github.com/loyalty/analytics/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockRFMStorage is a mock implementation of the RFM storage
type MockRFMStorage struct {
	mock.Mock
}

func (m *MockRFMStorage) GetOrCalculateQuintiles(ctx context.Context, orgID string) (models.RFMQuintiles, error) {
	args := m.Called(ctx, orgID)
	if args.Get(0) == nil {
		return models.RFMQuintiles{}, args.Error(1)
	}
	return args.Get(0).(models.RFMQuintiles), args.Error(1)
}

func (m *MockRFMStorage) SaveRFMScore(ctx context.Context, score models.RFMScore) error {
	args := m.Called(ctx, score)
	return args.Error(0)
}

func (m *MockRFMStorage) GetCustomerActivities(ctx context.Context, orgID string) ([]models.CustomerActivity, error) {
	args := m.Called(ctx, orgID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.CustomerActivity), args.Error(1)
}

// Test setup helper
func setupTestCalculator() (*RFMCalculator, *MockRFMStorage) {
	mockStorage := &MockRFMStorage{}
	calculator := NewRFMCalculator(mockStorage)
	return calculator, mockStorage
}

// Test NewRFMCalculator
func TestNewRFMCalculator(t *testing.T) {
	mockStorage := &MockRFMStorage{}
	calculator := NewRFMCalculator(mockStorage)
	
	assert.NotNil(t, calculator)
	assert.Equal(t, mockStorage, calculator.storage)
}

// Test ProcessCustomerTransaction
func TestProcessCustomerTransaction_Success(t *testing.T) {
	calculator, mockStorage := setupTestCalculator()
	ctx := context.Background()
	
	// Test data
	activity := models.CustomerActivity{
		OrgID:             "test_org",
		LocationID:        "test_location",
		CustomerID:        "test_customer",
		LastTransaction:   time.Now().AddDate(0, 0, -5), // 5 days ago
		FirstTransaction:  time.Now().AddDate(0, 0, -30), // 30 days ago
		TotalTransactions: 10,
		TotalSpent:        500.0,
	}
	
	quintiles := models.RFMQuintiles{
		OrgID:              "test_org",
		RecencyQuintiles:   []int{7, 30, 90, 180, 365},
		FrequencyQuintiles: []int{1, 2, 5, 10, 20},
		MonetaryQuintiles:  []float64{10.0, 25.0, 50.0, 100.0, 250.0},
	}
	
	// Setup expectations
	mockStorage.On("GetOrCalculateQuintiles", ctx, "test_org").Return(quintiles, nil)
	mockStorage.On("SaveRFMScore", ctx, mock.AnythingOfType("models.RFMScore")).Return(nil)
	
	// Process transaction
	err := calculator.ProcessCustomerTransaction(ctx, activity)
	
	// Assertions
	assert.NoError(t, err)
	
	// Verify the saved RFM score
	mockStorage.AssertCalled(t, "SaveRFMScore", ctx, mock.MatchedBy(func(score models.RFMScore) bool {
		return score.OrgID == "test_org" &&
			score.CustomerID == "test_customer" &&
			score.LocationID == "test_location" &&
			score.TotalTransactions == 10 &&
			score.TotalSpent == 500.0 &&
			score.AvgOrderValue == 50.0 &&
			score.DaysSinceLast == 5 &&
			score.DaysSinceFirst == 30
	}))
	
	mockStorage.AssertExpectations(t)
}

func TestProcessCustomerTransaction_GetQuintilesError(t *testing.T) {
	calculator, mockStorage := setupTestCalculator()
	ctx := context.Background()
	
	// Test data
	activity := models.CustomerActivity{
		OrgID:             "test_org",
		LocationID:        "test_location",
		CustomerID:        "test_customer",
		LastTransaction:   time.Now(),
		FirstTransaction:  time.Now(),
		TotalTransactions: 1,
		TotalSpent:        10.0,
	}
	
	// Setup expectations
	mockStorage.On("GetOrCalculateQuintiles", ctx, "test_org").Return(models.RFMQuintiles{}, assert.AnError)
	
	// Process transaction
	err := calculator.ProcessCustomerTransaction(ctx, activity)
	
	// Assertions
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get quintiles")
	
	mockStorage.AssertExpectations(t)
}

func TestProcessCustomerTransaction_SaveScoreError(t *testing.T) {
	calculator, mockStorage := setupTestCalculator()
	ctx := context.Background()
	
	// Test data
	activity := models.CustomerActivity{
		OrgID:             "test_org",
		LocationID:        "test_location",
		CustomerID:        "test_customer",
		LastTransaction:   time.Now(),
		FirstTransaction:  time.Now(),
		TotalTransactions: 1,
		TotalSpent:        10.0,
	}
	
	quintiles := models.RFMQuintiles{
		OrgID:              "test_org",
		RecencyQuintiles:   []int{7, 30, 90, 180, 365},
		FrequencyQuintiles: []int{1, 2, 5, 10, 20},
		MonetaryQuintiles:  []float64{10.0, 25.0, 50.0, 100.0, 250.0},
	}
	
	// Setup expectations
	mockStorage.On("GetOrCalculateQuintiles", ctx, "test_org").Return(quintiles, nil)
	mockStorage.On("SaveRFMScore", ctx, mock.AnythingOfType("models.RFMScore")).Return(assert.AnError)
	
	// Process transaction
	err := calculator.ProcessCustomerTransaction(ctx, activity)
	
	// Assertions
	assert.Error(t, err)
	
	mockStorage.AssertExpectations(t)
}

// Test calculateRFMScore
func TestCalculateRFMScore(t *testing.T) {
	calculator, _ := setupTestCalculator()
	
	now := time.Now()
	activity := models.CustomerActivity{
		OrgID:             "test_org",
		LocationID:        "test_location",
		CustomerID:        "test_customer",
		LastTransaction:   now.AddDate(0, 0, -5), // 5 days ago
		FirstTransaction:  now.AddDate(0, 0, -30), // 30 days ago
		TotalTransactions: 10,
		TotalSpent:        500.0,
	}
	
	quintiles := models.RFMQuintiles{
		OrgID:              "test_org",
		RecencyQuintiles:   []int{7, 30, 90, 180, 365},
		FrequencyQuintiles: []int{1, 2, 5, 10, 20},
		MonetaryQuintiles:  []float64{10.0, 25.0, 50.0, 100.0, 250.0},
	}
	
	score := calculator.calculateRFMScore(activity, quintiles)
	
	// Assertions
	assert.Equal(t, "test_org", score.OrgID)
	assert.Equal(t, "test_location", score.LocationID)
	assert.Equal(t, "test_customer", score.CustomerID)
	assert.Equal(t, 10, score.TotalTransactions)
	assert.Equal(t, 500.0, score.TotalSpent)
	assert.Equal(t, 50.0, score.AvgOrderValue)
	assert.Equal(t, 5, score.DaysSinceLast)
	assert.Equal(t, 30, score.DaysSinceFirst)
	assert.NotZero(t, score.CalculatedAt)
	assert.NotZero(t, score.UpdatedAt)
	
	// Verify scores are calculated correctly
	assert.GreaterOrEqual(t, score.RecencyScore, 1)
	assert.LessOrEqual(t, score.RecencyScore, 5)
	assert.GreaterOrEqual(t, score.FrequencyScore, 1)
	assert.LessOrEqual(t, score.FrequencyScore, 5)
	assert.GreaterOrEqual(t, score.MonetaryScore, 1)
	assert.LessOrEqual(t, score.MonetaryScore, 5)
	
	// Verify segment is assigned
	assert.NotEmpty(t, score.RFMSegment)
}

// Test getRecencyScore
func TestGetRecencyScore(t *testing.T) {
	calculator, _ := setupTestCalculator()
	
	quintiles := []int{7, 30, 90, 180, 365}
	
	tests := []struct {
		name           string
		daysSinceLast  int
		expectedScore  int
	}{
		{"very recent", 3, 5},
		{"recent", 15, 4},
		{"moderate", 60, 3},
		{"old", 200, 1},
		{"very old", 400, 1},
		{"extremely old", 500, 1},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := calculator.getRecencyScore(tt.daysSinceLast, quintiles)
			assert.Equal(t, tt.expectedScore, score)
		})
	}
}

// Test getFrequencyScore
func TestGetFrequencyScore(t *testing.T) {
	calculator, _ := setupTestCalculator()
	
	quintiles := []int{1, 2, 5, 10, 20}
	
	tests := []struct {
		name              string
		totalTransactions int
		expectedScore     int
	}{
		{"low frequency", 1, 1},
		{"moderate frequency", 3, 2},
		{"high frequency", 7, 3},
		{"very high frequency", 15, 4},
		{"extremely high frequency", 25, 5},
		{"zero transactions", 0, 1},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := calculator.getFrequencyScore(tt.totalTransactions, quintiles)
			assert.Equal(t, tt.expectedScore, score)
		})
	}
}

// Test getMonetaryScore
func TestGetMonetaryScore(t *testing.T) {
	calculator, _ := setupTestCalculator()
	
	quintiles := []float64{10.0, 25.0, 50.0, 100.0, 250.0}
	
	tests := []struct {
		name          string
		totalSpent    float64
		expectedScore int
	}{
		{"low value", 5.0, 1},
		{"moderate value", 30.0, 2},
		{"high value", 75.0, 3},
		{"very high value", 150.0, 4},
		{"extremely high value", 300.0, 5},
		{"zero spent", 0.0, 1},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := calculator.getMonetaryScore(tt.totalSpent, quintiles)
			assert.Equal(t, tt.expectedScore, score)
		})
	}
}

// Test getRFMSegment
func TestGetRFMSegment(t *testing.T) {
	calculator, _ := setupTestCalculator()
	
	tests := []struct {
		name           string
		recencyScore   int
		frequencyScore int
		monetaryScore  int
		expectedSegment string
	}{
		{"champions", 5, 5, 5, "Champions"},
		{"loyal customers", 4, 4, 4, "Loyal Customers"},
		{"potential loyalists", 4, 2, 2, "Potential Loyalists"},
		{"new customers", 5, 1, 1, "Potential Loyalists"},
		{"cannot lose them", 1, 5, 5, "Cannot Lose Them"},
		{"at risk", 2, 4, 4, "At Risk"},
		{"need attention", 3, 2, 4, "Need Attention"},
		{"about to sleep", 2, 2, 4, "Need Attention"},
		{"lost", 1, 1, 1, "Lost"},
		{"mixed scores", 3, 3, 2, "Lost"},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			segment := calculator.getRFMSegment(tt.recencyScore, tt.frequencyScore, tt.monetaryScore)
			assert.Equal(t, tt.expectedSegment, segment)
		})
	}
}

// Test CalculateQuintilesForOrg
func TestCalculateQuintilesForOrg_Success(t *testing.T) {
	calculator, mockStorage := setupTestCalculator()
	ctx := context.Background()
	
	// Test data
	now := time.Now()
	activities := []models.CustomerActivity{
		{
			OrgID:             "test_org",
			CustomerID:        "cust_1",
			LastTransaction:   now.AddDate(0, 0, -5),
			TotalTransactions: 5,
			TotalSpent:        100.0,
		},
		{
			OrgID:             "test_org",
			CustomerID:        "cust_2",
			LastTransaction:   now.AddDate(0, 0, -15),
			TotalTransactions: 10,
			TotalSpent:        200.0,
		},
		{
			OrgID:             "test_org",
			CustomerID:        "cust_3",
			LastTransaction:   now.AddDate(0, 0, -30),
			TotalTransactions: 15,
			TotalSpent:        300.0,
		},
		{
			OrgID:             "test_org",
			CustomerID:        "cust_4",
			LastTransaction:   now.AddDate(0, 0, -60),
			TotalTransactions: 20,
			TotalSpent:        400.0,
		},
		{
			OrgID:             "test_org",
			CustomerID:        "cust_5",
			LastTransaction:   now.AddDate(0, 0, -90),
			TotalTransactions: 25,
			TotalSpent:        500.0,
		},
	}
	
	// Setup expectations
	mockStorage.On("GetCustomerActivities", ctx, "test_org").Return(activities, nil)
	
	// Calculate quintiles
	quintiles, err := calculator.CalculateQuintilesForOrg(ctx, "test_org")
	
	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, "test_org", quintiles.OrgID)
	assert.Len(t, quintiles.RecencyQuintiles, 5)
	assert.Len(t, quintiles.FrequencyQuintiles, 5)
	assert.Len(t, quintiles.MonetaryQuintiles, 5)
	assert.NotZero(t, quintiles.CalculatedAt)
	
	// Verify quintiles are sorted
	for i := 1; i < len(quintiles.RecencyQuintiles); i++ {
		assert.GreaterOrEqual(t, quintiles.RecencyQuintiles[i], quintiles.RecencyQuintiles[i-1])
	}
	for i := 1; i < len(quintiles.FrequencyQuintiles); i++ {
		assert.GreaterOrEqual(t, quintiles.FrequencyQuintiles[i], quintiles.FrequencyQuintiles[i-1])
	}
	for i := 1; i < len(quintiles.MonetaryQuintiles); i++ {
		assert.GreaterOrEqual(t, quintiles.MonetaryQuintiles[i], quintiles.MonetaryQuintiles[i-1])
	}
	
	mockStorage.AssertExpectations(t)
}

func TestCalculateQuintilesForOrg_GetActivitiesError(t *testing.T) {
	calculator, mockStorage := setupTestCalculator()
	ctx := context.Background()
	
	// Setup expectations
	mockStorage.On("GetCustomerActivities", ctx, "test_org").Return(nil, assert.AnError)
	
	// Calculate quintiles
	quintiles, err := calculator.CalculateQuintilesForOrg(ctx, "test_org")
	
	// Assertions
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get customer activities")
	assert.Equal(t, models.RFMQuintiles{}, quintiles)
	
	mockStorage.AssertExpectations(t)
}

func TestCalculateQuintilesForOrg_LessThan5Customers(t *testing.T) {
	calculator, mockStorage := setupTestCalculator()
	ctx := context.Background()
	
	// Test data with less than 5 customers
	now := time.Now()
	activities := []models.CustomerActivity{
		{
			OrgID:             "test_org",
			CustomerID:        "cust_1",
			LastTransaction:   now.AddDate(0, 0, -5),
			TotalTransactions: 5,
			TotalSpent:        100.0,
		},
		{
			OrgID:             "test_org",
			CustomerID:        "cust_2",
			LastTransaction:   now.AddDate(0, 0, -15),
			TotalTransactions: 10,
			TotalSpent:        200.0,
		},
	}
	
	// Setup expectations
	mockStorage.On("GetCustomerActivities", ctx, "test_org").Return(activities, nil)
	
	// Calculate quintiles
	quintiles, err := calculator.CalculateQuintilesForOrg(ctx, "test_org")
	
	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, "test_org", quintiles.OrgID)
	assert.Equal(t, []int{7, 30, 90, 180, 365}, quintiles.RecencyQuintiles)
	assert.Equal(t, []int{1, 2, 5, 10, 20}, quintiles.FrequencyQuintiles)
	assert.Equal(t, []float64{10.0, 25.0, 50.0, 100.0, 250.0}, quintiles.MonetaryQuintiles)
	assert.NotZero(t, quintiles.CalculatedAt)
	
	mockStorage.AssertExpectations(t)
}

// Test calculateIntQuintiles
func TestCalculateIntQuintiles(t *testing.T) {
	calculator, _ := setupTestCalculator()
	
	values := []int{1, 3, 5, 7, 9, 11, 13, 15, 17, 19}
	quintiles := calculator.calculateIntQuintiles(values)
	
	// Assertions
	assert.Len(t, quintiles, 5)
	assert.Equal(t, 3, quintiles[0])  // 20th percentile
	assert.Equal(t, 7, quintiles[1])  // 40th percentile
	assert.Equal(t, 13, quintiles[2]) // 60th percentile
	assert.Equal(t, 15, quintiles[3]) // 80th percentile
	assert.Equal(t, 19, quintiles[4]) // 100th percentile
}

// Test calculateFloatQuintiles
func TestCalculateFloatQuintiles(t *testing.T) {
	calculator, _ := setupTestCalculator()
	
	values := []float64{1.0, 3.0, 5.0, 7.0, 9.0, 11.0, 13.0, 15.0, 17.0, 19.0}
	quintiles := calculator.calculateFloatQuintiles(values)
	
	// Assertions
	assert.Len(t, quintiles, 5)
	assert.Equal(t, 3.0, quintiles[0])  // 20th percentile
	assert.Equal(t, 7.0, quintiles[1])  // 40th percentile
	assert.Equal(t, 13.0, quintiles[2]) // 60th percentile
	assert.Equal(t, 15.0, quintiles[3]) // 80th percentile
	assert.Equal(t, 19.0, quintiles[4]) // 100th percentile
}

// Test getDefaultQuintiles
func TestGetDefaultQuintiles(t *testing.T) {
	calculator, _ := setupTestCalculator()
	
	quintiles := calculator.getDefaultQuintiles("test_org")
	
	// Assertions
	assert.Equal(t, "test_org", quintiles.OrgID)
	assert.Equal(t, []int{7, 30, 90, 180, 365}, quintiles.RecencyQuintiles)
	assert.Equal(t, []int{1, 2, 5, 10, 20}, quintiles.FrequencyQuintiles)
	assert.Equal(t, []float64{10.0, 25.0, 50.0, 100.0, 250.0}, quintiles.MonetaryQuintiles)
	assert.NotZero(t, quintiles.CalculatedAt)
} 