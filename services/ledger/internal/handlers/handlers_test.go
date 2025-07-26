package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/loyalty/ledger/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockTigerBeetleRepo is a mock implementation of the repository
type MockTigerBeetleRepo struct {
	mock.Mock
}

func (m *MockTigerBeetleRepo) CreateAccount(ctx context.Context, req *models.CreateAccountRequest) (*models.Account, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Account), args.Error(1)
}

func (m *MockTigerBeetleRepo) CreateTransfer(ctx context.Context, req *models.CreateTransferRequest) (*models.TransferResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.TransferResponse), args.Error(1)
}

func (m *MockTigerBeetleRepo) GetAccount(ctx context.Context, accountID string) (*models.Account, error) {
	args := m.Called(ctx, accountID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Account), args.Error(1)
}

func (m *MockTigerBeetleRepo) GetBalance(ctx context.Context, orgID, customerID string) (map[string]uint64, error) {
	args := m.Called(ctx, orgID, customerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]uint64), args.Error(1)
}

func (m *MockTigerBeetleRepo) Close() error {
	args := m.Called()
	return args.Error(0)
}

// Test setup helper
func setupTest() (*gin.Engine, *MockTigerBeetleRepo, *LedgerHandler) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	
	mockRepo := &MockTigerBeetleRepo{}
	handler := &LedgerHandler{repo: mockRepo}
	
	return router, mockRepo, handler
}

// Test CreateAccount
func TestCreateAccount_Success(t *testing.T) {
	router, mockRepo, handler := setupTest()
	
	// Setup route
	router.POST("/accounts", handler.CreateAccount)
	
	// Test data
	reqBody := models.CreateAccountRequest{
		OrgID:       "test_org",
		CustomerID:  "test_customer",
		AccountType: models.AccountTypeAsset,
		Code:        1001,
	}
	
	jsonData, _ := json.Marshal(reqBody)
	
	// Mock repository response
	expectedAccount := &models.Account{
		ID:          "acc_123",
		OrgID:       "test_org",
		CustomerID:  "test_customer",
		AccountType: models.AccountTypeAsset,
		Code:        1001,
	}
	
	mockRepo.On("CreateAccount", mock.Anything, &reqBody).Return(expectedAccount, nil)
	
	// Create request
	req, _ := http.NewRequest("POST", "/accounts", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	
	// Record response
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	// Assertions
	assert.Equal(t, http.StatusCreated, w.Code)
	
	var response models.Account
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "acc_123", response.ID)
	assert.Equal(t, "test_org", response.OrgID)
	assert.Equal(t, "test_customer", response.CustomerID)
	
	mockRepo.AssertExpectations(t)
}

func TestCreateAccount_InvalidRequest(t *testing.T) {
	router, _, handler := setupTest()
	
	// Setup route
	router.POST("/accounts", handler.CreateAccount)
	
	// Invalid JSON
	req, _ := http.NewRequest("POST", "/accounts", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	
	// Record response
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	// Assertions
	assert.Equal(t, http.StatusBadRequest, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response["error"], "invalid")
}

func TestCreateAccount_RepositoryError(t *testing.T) {
	router, mockRepo, handler := setupTest()
	
	// Setup route
	router.POST("/accounts", handler.CreateAccount)
	
	// Test data
	reqBody := models.CreateAccountRequest{
		OrgID:       "test_org",
		CustomerID:  "test_customer",
		AccountType: models.AccountTypeAsset,
		Code:        1001,
	}
	
	jsonData, _ := json.Marshal(reqBody)
	
	// Mock repository error
	mockRepo.On("CreateAccount", mock.Anything, &reqBody).Return(nil, assert.AnError)
	
	// Create request
	req, _ := http.NewRequest("POST", "/accounts", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	
	// Record response
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	// Assertions
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response["error"], "assert.AnError")
	
	mockRepo.AssertExpectations(t)
}

// Test CreateTransfer
func TestCreateTransfer_Success(t *testing.T) {
	router, mockRepo, handler := setupTest()
	
	// Setup route
	router.POST("/transfers", handler.CreateTransfer)
	
	// Test data
	reqBody := models.CreateTransferRequest{
		OrgID:           "test_org",
		CustomerID:      "test_customer",
		TransactionType: "points_accrual",
		Amount:          100,
		Code:            1,
		Reference:       "test_transfer",
	}
	
	jsonData, _ := json.Marshal(reqBody)
	
	// Mock repository response
	expectedResponse := &models.TransferResponse{
		TransferID: "transfer_123",
		Status:     "success",
	}
	
	mockRepo.On("CreateTransfer", mock.Anything, &reqBody).Return(expectedResponse, nil)
	
	// Create request
	req, _ := http.NewRequest("POST", "/transfers", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	
	// Record response
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	// Assertions
	assert.Equal(t, http.StatusCreated, w.Code)
	
	var response models.TransferResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "transfer_123", response.TransferID)
	assert.Equal(t, "success", response.Status)
	
	mockRepo.AssertExpectations(t)
}

func TestCreateTransfer_InvalidRequest(t *testing.T) {
	router, _, handler := setupTest()
	
	// Setup route
	router.POST("/transfers", handler.CreateTransfer)
	
	// Invalid JSON
	req, _ := http.NewRequest("POST", "/transfers", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	
	// Record response
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	// Assertions
	assert.Equal(t, http.StatusBadRequest, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response["error"], "invalid")
}

func TestCreateTransfer_RepositoryError(t *testing.T) {
	router, mockRepo, handler := setupTest()
	
	// Setup route
	router.POST("/transfers", handler.CreateTransfer)
	
	// Test data
	reqBody := models.CreateTransferRequest{
		OrgID:           "test_org",
		CustomerID:      "test_customer",
		TransactionType: "points_accrual",
		Amount:          100,
		Code:            1,
		Reference:       "test_transfer",
	}
	
	jsonData, _ := json.Marshal(reqBody)
	
	// Mock repository error
	mockRepo.On("CreateTransfer", mock.Anything, &reqBody).Return(nil, assert.AnError)
	
	// Create request
	req, _ := http.NewRequest("POST", "/transfers", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	
	// Record response
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	// Assertions
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response["error"], "assert.AnError")
	
	mockRepo.AssertExpectations(t)
}

// Test GetAccount
func TestGetAccount_Success(t *testing.T) {
	router, mockRepo, handler := setupTest()
	
	// Setup route
	router.GET("/accounts/:id", handler.GetAccount)
	
	// Mock repository response
	expectedAccount := &models.Account{
		ID:          "acc_123",
		OrgID:       "test_org",
		CustomerID:  "test_customer",
		AccountType: models.AccountTypeAsset,
		Code:        1001,
	}
	
	mockRepo.On("GetAccount", mock.Anything, "acc_123").Return(expectedAccount, nil)
	
	// Create request
	req, _ := http.NewRequest("GET", "/accounts/acc_123", nil)
	
	// Record response
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response models.Account
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "acc_123", response.ID)
	assert.Equal(t, "test_org", response.OrgID)
	
	mockRepo.AssertExpectations(t)
}

func TestGetAccount_MissingID(t *testing.T) {
	router, _, handler := setupTest()
	
	// Setup route
	router.GET("/accounts/:id", handler.GetAccount)
	
	// Create request without ID
	req, _ := http.NewRequest("GET", "/accounts/", nil)
	
	// Record response
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	// Assertions
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestGetAccount_NotFound(t *testing.T) {
	router, mockRepo, handler := setupTest()
	
	// Setup route
	router.GET("/accounts/:id", handler.GetAccount)
	
	// Mock repository error
	mockRepo.On("GetAccount", mock.Anything, "nonexistent").Return(nil, assert.AnError)
	
	// Create request
	req, _ := http.NewRequest("GET", "/accounts/nonexistent", nil)
	
	// Record response
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	// Assertions
	assert.Equal(t, http.StatusNotFound, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response["error"], "assert.AnError")
	
	mockRepo.AssertExpectations(t)
}

// Test GetBalance
func TestGetBalance_Success(t *testing.T) {
	router, mockRepo, handler := setupTest()
	
	// Setup route
	router.GET("/balance", handler.GetBalance)
	
	// Mock repository response
	expectedBalances := map[string]uint64{
		"points": 500,
		"stamps": 10,
	}
	
	mockRepo.On("GetBalance", mock.Anything, "test_org", "test_customer").Return(expectedBalances, nil)
	
	// Create request
	req, _ := http.NewRequest("GET", "/balance?org_id=test_org&customer_id=test_customer", nil)
	
	// Record response
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "test_org", response["org_id"])
	assert.Equal(t, "test_customer", response["customer_id"])
	assert.Equal(t, float64(500), response["points_balance"])
	assert.Equal(t, float64(10), response["stamps_balance"])
	
	mockRepo.AssertExpectations(t)
}

func TestGetBalance_MissingOrgID(t *testing.T) {
	router, _, handler := setupTest()
	
	// Setup route
	router.GET("/balance", handler.GetBalance)
	
	// Create request without org_id
	req, _ := http.NewRequest("GET", "/balance?customer_id=test_customer", nil)
	
	// Record response
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	// Assertions
	assert.Equal(t, http.StatusBadRequest, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response["error"], "org_id and customer_id are required")
}

func TestGetBalance_MissingCustomerID(t *testing.T) {
	router, _, handler := setupTest()
	
	// Setup route
	router.GET("/balance", handler.GetBalance)
	
	// Create request without customer_id
	req, _ := http.NewRequest("GET", "/balance?org_id=test_org", nil)
	
	// Record response
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	// Assertions
	assert.Equal(t, http.StatusBadRequest, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response["error"], "org_id and customer_id are required")
}

func TestGetBalance_RepositoryError(t *testing.T) {
	router, mockRepo, handler := setupTest()
	
	// Setup route
	router.GET("/balance", handler.GetBalance)
	
	// Mock repository error
	mockRepo.On("GetBalance", mock.Anything, "test_org", "test_customer").Return(nil, assert.AnError)
	
	// Create request
	req, _ := http.NewRequest("GET", "/balance?org_id=test_org&customer_id=test_customer", nil)
	
	// Record response
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	// Assertions
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response["error"], "assert.AnError")
	
	mockRepo.AssertExpectations(t)
}

// Test Health
func TestHealth_Success(t *testing.T) {
	router, _, handler := setupTest()
	
	// Setup route
	router.GET("/health", handler.Health)
	
	// Create request
	req, _ := http.NewRequest("GET", "/health", nil)
	
	// Record response
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "healthy", response["status"])
	assert.Equal(t, "ledger", response["service"])
} 