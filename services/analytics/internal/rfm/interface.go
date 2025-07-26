package rfm

import (
	"context"
	"github.com/loyalty/analytics/internal/models"
)

// RFMStorageInterface defines the interface for RFM storage operations
type RFMStorageInterface interface {
	GetOrCalculateQuintiles(ctx context.Context, orgID string) (models.RFMQuintiles, error)
	SaveRFMScore(ctx context.Context, score models.RFMScore) error
	GetCustomerActivities(ctx context.Context, orgID string) ([]models.CustomerActivity, error)
} 