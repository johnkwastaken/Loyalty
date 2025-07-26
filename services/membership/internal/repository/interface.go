package repository

import (
	"context"
	"github.com/loyalty/membership/internal/models"
	"go.mongodb.org/mongo-driver/bson"
)

// MongoRepoInterface defines the interface for MongoDB repository operations
type MongoRepoInterface interface {
	CreateCustomer(ctx context.Context, req *models.CreateCustomerRequest) (*models.Customer, error)
	GetCustomer(ctx context.Context, customerID string) (*models.Customer, error)
	GetCustomersByOrg(ctx context.Context, orgID string, limit, offset int) ([]*models.Customer, error)
	UpdateCustomer(ctx context.Context, customerID string, updates bson.M) error
	CreateOrganization(ctx context.Context, org *models.Organization) error
	GetOrganization(ctx context.Context, orgID string) (*models.Organization, error)
	CreateLocation(ctx context.Context, req *models.CreateLocationRequest) (*models.Location, error)
	GetLocation(ctx context.Context, locationID string) (*models.Location, error)
	GetLocationsByOrg(ctx context.Context, orgID string, limit, offset int) ([]*models.Location, error)
	UpdateLocation(ctx context.Context, locationID string, updates bson.M) error
	Close() error
} 