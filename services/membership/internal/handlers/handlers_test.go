package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/loyalty/membership/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// MockMongoRepo is a mock implementation of the repository
type MockMongoRepo struct {
	mock.Mock
}

func (m *MockMongoRepo) CreateCustomer(ctx context.Context, req *models.CreateCustomerRequest) (*models.Customer, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Customer), args.Error(1)
}

func (m *MockMongoRepo) GetCustomer(ctx context.Context, customerID string) (*models.Customer, error) {
	args := m.Called(ctx, customerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Customer), args.Error(1)
}

func (m *MockMongoRepo) GetCustomersByOrg(ctx context.Context, orgID string, limit, offset int) ([]*models.Customer, error) {
	args := m.Called(ctx, orgID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Customer), args.Error(1)
}

func (m *MockMongoRepo) UpdateCustomer(ctx context.Context, customerID string, updates bson.M) error {
	args := m.Called(ctx, customerID, updates)
	return args.Error(0)
}

func (m *MockMongoRepo) CreateOrganization(ctx context.Context, org *models.Organization) error {
	args := m.Called(ctx, org)
	return args.Error(0)
}

func (m *MockMongoRepo) GetOrganization(ctx context.Context, orgID string) (*models.Organization, error) {
	args := m.Called(ctx, orgID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Organization), args.Error(1)
}

func (m *MockMongoRepo) CreateLocation(ctx context.Context, req *models.CreateLocationRequest) (*models.Location, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Location), args.Error(1)
}

func (m *MockMongoRepo) GetLocation(ctx context.Context, locationID string) (*models.Location, error) {
	args := m.Called(ctx, locationID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Location), args.Error(1)
}

func (m *MockMongoRepo) GetLocationsByOrg(ctx context.Context, orgID string, limit, offset int) ([]*models.Location, error) {
	args := m.Called(ctx, orgID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Location), args.Error(1)
}

func (m *MockMongoRepo) UpdateLocation(ctx context.Context, locationID string, updates bson.M) error {
	args := m.Called(ctx, locationID, updates)
	return args.Error(0)
}

func (m *MockMongoRepo) Close() error {
	args := m.Called()
	return args.Error(0)
}

// Test setup helper
func setupTest() (*gin.Engine, *MockMongoRepo, *MembershipHandler) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	
	mockRepo := &MockMongoRepo{}
	handler := &MembershipHandler{repo: mockRepo}
	
	return router, mockRepo, handler
}

// Test CreateCustomer
func TestCreateCustomer_Success(t *testing.T) {
	router, mockRepo, handler := setupTest()
	
	// Setup route
	router.POST("/customers", handler.CreateCustomer)
	
	// Test data
	reqBody := models.CreateCustomerRequest{
		OrgID:     "test_org",
		Email:     "test@example.com",
		FirstName: "John",
		LastName:  "Doe",
		Phone:     "+1234567890",
	}
	
	jsonData, _ := json.Marshal(reqBody)
	
	// Mock repository response
	expectedCustomer := &models.Customer{
		ID:         primitive.NewObjectID(),
		CustomerID: "cust_123",
		OrgID:      "test_org",
		Email:      "test@example.com",
		FirstName:  "John",
		LastName:   "Doe",
		Phone:      "+1234567890",
		Status:     "active",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	
	mockRepo.On("CreateCustomer", mock.Anything, &reqBody).Return(expectedCustomer, nil)
	
	// Create request
	req, _ := http.NewRequest("POST", "/customers", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	
	// Record response
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	// Assertions
	assert.Equal(t, http.StatusCreated, w.Code)
	
	var response models.Customer
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "test_org", response.OrgID)
	assert.Equal(t, "test@example.com", response.Email)
	assert.Equal(t, "John", response.FirstName)
	assert.Equal(t, "Doe", response.LastName)
	
	mockRepo.AssertExpectations(t)
}

func TestCreateCustomer_InvalidRequest(t *testing.T) {
	router, _, handler := setupTest()
	
	// Setup route
	router.POST("/customers", handler.CreateCustomer)
	
	// Invalid JSON
	req, _ := http.NewRequest("POST", "/customers", bytes.NewBufferString("invalid json"))
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

func TestCreateCustomer_RepositoryError(t *testing.T) {
	router, mockRepo, handler := setupTest()
	
	// Setup route
	router.POST("/customers", handler.CreateCustomer)
	
	// Test data
	reqBody := models.CreateCustomerRequest{
		OrgID:     "test_org",
		Email:     "test@example.com",
		FirstName: "John",
		LastName:  "Doe",
	}
	
	jsonData, _ := json.Marshal(reqBody)
	
	// Mock repository error
	mockRepo.On("CreateCustomer", mock.Anything, &reqBody).Return(nil, assert.AnError)
	
	// Create request
	req, _ := http.NewRequest("POST", "/customers", bytes.NewBuffer(jsonData))
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

// Test GetCustomer
func TestGetCustomer_Success(t *testing.T) {
	router, mockRepo, handler := setupTest()
	
	// Setup route
	router.GET("/customers/:id", handler.GetCustomer)
	
	// Mock repository response
	expectedCustomer := &models.Customer{
		ID:         primitive.NewObjectID(),
		CustomerID: "cust_123",
		OrgID:      "test_org",
		Email:      "test@example.com",
		FirstName:  "John",
		LastName:   "Doe",
		Status:     "active",
	}
	
	mockRepo.On("GetCustomer", mock.Anything, "cust_123").Return(expectedCustomer, nil)
	
	// Create request
	req, _ := http.NewRequest("GET", "/customers/cust_123", nil)
	
	// Record response
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response models.Customer
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "cust_123", response.CustomerID)
	assert.Equal(t, "test@example.com", response.Email)
	
	mockRepo.AssertExpectations(t)
}

func TestGetCustomer_MissingID(t *testing.T) {
	router, _, handler := setupTest()
	
	// Setup route
	router.GET("/customers/:id", handler.GetCustomer)
	
	// Create request without ID
	req, _ := http.NewRequest("GET", "/customers/", nil)
	
	// Record response
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	// Assertions
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestGetCustomer_NotFound(t *testing.T) {
	router, mockRepo, handler := setupTest()
	
	// Setup route
	router.GET("/customers/:id", handler.GetCustomer)
	
	// Mock repository error
	mockRepo.On("GetCustomer", mock.Anything, "nonexistent").Return(nil, assert.AnError)
	
	// Create request
	req, _ := http.NewRequest("GET", "/customers/nonexistent", nil)
	
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

// Test GetCustomersByOrg
func TestGetCustomersByOrg_Success(t *testing.T) {
	router, mockRepo, handler := setupTest()
	
	// Setup route
	router.GET("/customers", handler.GetCustomersByOrg)
	
	// Mock repository response
	expectedCustomers := []*models.Customer{
		{
			ID:         primitive.NewObjectID(),
			CustomerID: "cust_1",
			OrgID:      "test_org",
			Email:      "customer1@example.com",
			FirstName:  "John",
			LastName:   "Doe",
		},
		{
			ID:         primitive.NewObjectID(),
			CustomerID: "cust_2",
			OrgID:      "test_org",
			Email:      "customer2@example.com",
			FirstName:  "Jane",
			LastName:   "Smith",
		},
	}
	
	mockRepo.On("GetCustomersByOrg", mock.Anything, "test_org", 10, 0).Return(expectedCustomers, nil)
	
	// Create request
	req, _ := http.NewRequest("GET", "/customers?org_id=test_org&limit=10&offset=0", nil)
	
	// Record response
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, float64(2), response["count"])
	assert.Equal(t, float64(10), response["limit"])
	assert.Equal(t, float64(0), response["offset"])
	
	customers := response["customers"].([]interface{})
	assert.Len(t, customers, 2)
	
	mockRepo.AssertExpectations(t)
}

func TestGetCustomersByOrg_MissingOrgID(t *testing.T) {
	router, _, handler := setupTest()
	
	// Setup route
	router.GET("/customers", handler.GetCustomersByOrg)
	
	// Create request without org_id
	req, _ := http.NewRequest("GET", "/customers", nil)
	
	// Record response
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	// Assertions
	assert.Equal(t, http.StatusBadRequest, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response["error"], "org_id is required")
}

func TestGetCustomersByOrg_InvalidLimit(t *testing.T) {
	router, _, handler := setupTest()
	
	// Setup route
	router.GET("/customers", handler.GetCustomersByOrg)
	
	// Create request with invalid limit
	req, _ := http.NewRequest("GET", "/customers?org_id=test_org&limit=invalid", nil)
	
	// Record response
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	// Assertions
	assert.Equal(t, http.StatusBadRequest, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response["error"], "invalid limit parameter")
}

func TestGetCustomersByOrg_InvalidOffset(t *testing.T) {
	router, _, handler := setupTest()
	
	// Setup route
	router.GET("/customers", handler.GetCustomersByOrg)
	
	// Create request with invalid offset
	req, _ := http.NewRequest("GET", "/customers?org_id=test_org&offset=invalid", nil)
	
	// Record response
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	// Assertions
	assert.Equal(t, http.StatusBadRequest, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response["error"], "invalid offset parameter")
}

// Test UpdateCustomer
func TestUpdateCustomer_Success(t *testing.T) {
	router, mockRepo, handler := setupTest()
	
	// Setup route
	router.PATCH("/customers/:id", handler.UpdateCustomer)
	
	// Test data
	updates := map[string]interface{}{
		"first_name": "Updated",
		"last_name":  "Name",
		"email":      "updated@example.com",
	}
	
	jsonData, _ := json.Marshal(updates)
	
	// Mock repository
	mockRepo.On("UpdateCustomer", mock.Anything, "cust_123", bson.M(updates)).Return(nil)
	
	// Create request
	req, _ := http.NewRequest("PATCH", "/customers/cust_123", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	
	// Record response
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response["message"], "customer updated successfully")
	
	mockRepo.AssertExpectations(t)
}

func TestUpdateCustomer_MissingID(t *testing.T) {
	router, _, handler := setupTest()
	
	// Setup route
	router.PATCH("/customers/:id", handler.UpdateCustomer)
	
	// Create request without ID
	req, _ := http.NewRequest("PATCH", "/customers/", nil)
	
	// Record response
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	// Assertions
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestUpdateCustomer_InvalidRequest(t *testing.T) {
	router, _, handler := setupTest()
	
	// Setup route
	router.PATCH("/customers/:id", handler.UpdateCustomer)
	
	// Invalid JSON
	req, _ := http.NewRequest("PATCH", "/customers/cust_123", bytes.NewBufferString("invalid json"))
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

// Test CreateOrganization
func TestCreateOrganization_Success(t *testing.T) {
	router, mockRepo, handler := setupTest()
	
	// Setup route
	router.POST("/organizations", handler.CreateOrganization)
	
	// Test data
	org := models.Organization{
		OrgID:        "test_org",
		Name:         "Test Organization",
		Description:  "Test organization description",
		Settings:     models.OrgSettings{},
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	
	jsonData, _ := json.Marshal(org)
	
	// Mock repository
	mockRepo.On("CreateOrganization", mock.Anything, mock.AnythingOfType("*models.Organization")).Return(nil)
	
	// Create request
	req, _ := http.NewRequest("POST", "/organizations", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	
	// Record response
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	// Assertions
	assert.Equal(t, http.StatusCreated, w.Code)
	
	var response models.Organization
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "test_org", response.OrgID)
	assert.Equal(t, "Test Organization", response.Name)
	
	mockRepo.AssertExpectations(t)
}

// Test GetOrganization
func TestGetOrganization_Success(t *testing.T) {
	router, mockRepo, handler := setupTest()
	
	// Setup route
	router.GET("/organizations/:id", handler.GetOrganization)
	
	// Mock repository response
	expectedOrg := &models.Organization{
		OrgID:        "test_org",
		Name:         "Test Organization",
		Description:  "Test organization description",
		Settings:     models.OrgSettings{},
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	
	mockRepo.On("GetOrganization", mock.Anything, "test_org").Return(expectedOrg, nil)
	
	// Create request
	req, _ := http.NewRequest("GET", "/organizations/test_org", nil)
	
	// Record response
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response models.Organization
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "test_org", response.OrgID)
	assert.Equal(t, "Test Organization", response.Name)
	
	mockRepo.AssertExpectations(t)
}

func TestGetOrganization_MissingID(t *testing.T) {
	router, _, handler := setupTest()
	
	// Setup route
	router.GET("/organizations/:id", handler.GetOrganization)
	
	// Create request without ID
	req, _ := http.NewRequest("GET", "/organizations/", nil)
	
	// Record response
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	// Assertions
	assert.Equal(t, http.StatusNotFound, w.Code)
}

// Test CreateLocation
func TestCreateLocation_Success(t *testing.T) {
	router, mockRepo, handler := setupTest()
	
	// Setup route
	router.POST("/locations", handler.CreateLocation)
	
	// Test data
	reqBody := models.CreateLocationRequest{
		OrgID:   "test_org",
		Name:    "Test Location",
		Address: models.Address{
			Street:  "123 Test St",
			City:    "Test City",
			State:   "Test State",
			ZipCode: "12345",
			Country: "Test Country",
		},
		Settings: models.LocationSettings{},
	}
	
	jsonData, _ := json.Marshal(reqBody)
	
	// Mock repository response
	expectedLocation := &models.Location{
		ID:         primitive.NewObjectID(),
		LocationID: "loc_123",
		OrgID:      "test_org",
		Name:       "Test Location",
		Address: models.Address{
			Street:  "123 Test St",
			City:    "Test City",
			State:   "Test State",
			ZipCode: "12345",
			Country: "Test Country",
		},
		Settings: models.LocationSettings{},
		Active:   true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	
	mockRepo.On("CreateLocation", mock.Anything, &reqBody).Return(expectedLocation, nil)
	
	// Create request
	req, _ := http.NewRequest("POST", "/locations", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	
	// Record response
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	// Assertions
	assert.Equal(t, http.StatusCreated, w.Code)
	
	var response models.Location
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "test_org", response.OrgID)
	assert.Equal(t, "Test Location", response.Name)
	assert.Equal(t, "123 Test St", response.Address.Street)
	
	mockRepo.AssertExpectations(t)
}

// Test GetLocation
func TestGetLocation_Success(t *testing.T) {
	router, mockRepo, handler := setupTest()
	
	// Setup route
	router.GET("/locations/:id", handler.GetLocation)
	
	// Mock repository response
	expectedLocation := &models.Location{
		ID:         primitive.NewObjectID(),
		LocationID: "loc_123",
		OrgID:      "test_org",
		Name:       "Test Location",
		Address: models.Address{
			Street: "123 Test St",
		},
		Settings: models.LocationSettings{},
		Active:   true,
	}
	
	mockRepo.On("GetLocation", mock.Anything, "loc_123").Return(expectedLocation, nil)
	
	// Create request
	req, _ := http.NewRequest("GET", "/locations/loc_123", nil)
	
	// Record response
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response models.Location
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "loc_123", response.LocationID)
	assert.Equal(t, "Test Location", response.Name)
	
	mockRepo.AssertExpectations(t)
}

// Test GetLocationsByOrg
func TestGetLocationsByOrg_Success(t *testing.T) {
	router, mockRepo, handler := setupTest()
	
	// Setup route
	router.GET("/locations", handler.GetLocationsByOrg)
	
	// Mock repository response
	expectedLocations := []*models.Location{
		{
			ID:         primitive.NewObjectID(),
			LocationID: "loc_1",
			OrgID:      "test_org",
			Name:       "Location 1",
			Address: models.Address{
				Street: "123 Test St",
			},
			Settings: models.LocationSettings{},
			Active:   true,
		},
		{
			ID:         primitive.NewObjectID(),
			LocationID: "loc_2",
			OrgID:      "test_org",
			Name:       "Location 2",
			Address: models.Address{
				Street: "456 Test St",
			},
			Settings: models.LocationSettings{},
			Active:   true,
		},
	}
	
	mockRepo.On("GetLocationsByOrg", mock.Anything, "test_org", 10, 0).Return(expectedLocations, nil)
	
	// Create request
	req, _ := http.NewRequest("GET", "/locations?org_id=test_org&limit=10&offset=0", nil)
	
	// Record response
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, float64(2), response["count"])
	assert.Equal(t, float64(10), response["limit"])
	assert.Equal(t, float64(0), response["offset"])
	
	locations := response["locations"].([]interface{})
	assert.Len(t, locations, 2)
	
	mockRepo.AssertExpectations(t)
}

// Test UpdateLocation
func TestUpdateLocation_Success(t *testing.T) {
	router, mockRepo, handler := setupTest()
	
	// Setup route
	router.PATCH("/locations/:id", handler.UpdateLocation)
	
	// Test data
	updates := map[string]interface{}{
		"name":    "Updated Location",
		"address": "456 Updated St",
	}
	
	jsonData, _ := json.Marshal(updates)
	
	// Mock repository
	mockRepo.On("UpdateLocation", mock.Anything, "loc_123", bson.M(updates)).Return(nil)
	
	// Create request
	req, _ := http.NewRequest("PATCH", "/locations/loc_123", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	
	// Record response
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response["message"], "location updated successfully")
	
	mockRepo.AssertExpectations(t)
}

// Test DeactivateLocation
func TestDeactivateLocation_Success(t *testing.T) {
	router, mockRepo, handler := setupTest()
	
	// Setup route
	router.DELETE("/locations/:id", handler.DeactivateLocation)
	
	// Mock repository
	expectedUpdates := bson.M{"active": false}
	mockRepo.On("UpdateLocation", mock.Anything, "loc_123", expectedUpdates).Return(nil)
	
	// Create request
	req, _ := http.NewRequest("DELETE", "/locations/loc_123", nil)
	
	// Record response
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response["message"], "location deactivated successfully")
	
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
	assert.Equal(t, "membership", response["service"])
} 