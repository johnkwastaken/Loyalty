package rfm

import (
	"context"
	"fmt"
	"time"

	"github.com/loyalty/analytics/internal/models"
	"github.com/loyalty/analytics/internal/storage"
)

type RFMStorage struct {
	mongo *storage.MongoStorage
}

func NewRFMStorage(mongo *storage.MongoStorage) *RFMStorage {
	return &RFMStorage{mongo: mongo}
}

func (s *RFMStorage) SaveRFMScore(ctx context.Context, score models.RFMScore) error {
	return s.mongo.SaveRFMScore(ctx, score)
}

func (s *RFMStorage) GetRFMScore(ctx context.Context, orgID, customerID string) (*models.RFMScore, error) {
	return s.mongo.GetRFMScore(ctx, orgID, customerID)
}

func (s *RFMStorage) GetRFMScoreByLocation(ctx context.Context, orgID, locationID, customerID string) (*models.RFMScore, error) {
	return s.mongo.GetRFMScoreByLocation(ctx, orgID, locationID, customerID)
}

func (s *RFMStorage) GetOrCalculateQuintiles(ctx context.Context, orgID string) (models.RFMQuintiles, error) {
	quintiles, err := s.mongo.GetQuintiles(ctx, orgID)
	if err != nil {
		calculator := NewRFMCalculator(s)
		newQuintiles, calcErr := calculator.CalculateQuintilesForOrg(ctx, orgID)
		if calcErr != nil {
			return models.RFMQuintiles{}, fmt.Errorf("failed to calculate quintiles: %w", calcErr)
		}
		
		if saveErr := s.mongo.SaveQuintiles(ctx, newQuintiles); saveErr != nil {
			return newQuintiles, nil
		}
		
		return newQuintiles, nil
	}
	
	if time.Since(quintiles.CalculatedAt) > 24*time.Hour {
		calculator := NewRFMCalculator(s)
		newQuintiles, calcErr := calculator.CalculateQuintilesForOrg(ctx, orgID)
		if calcErr == nil {
			s.mongo.SaveQuintiles(ctx, newQuintiles)
			return newQuintiles, nil
		}
	}
	
	return *quintiles, nil
}

func (s *RFMStorage) UpdateCustomerActivity(ctx context.Context, activity models.CustomerActivity) error {
	return s.mongo.UpdateCustomerActivity(ctx, activity)
}

func (s *RFMStorage) GetCustomerActivities(ctx context.Context, orgID string) ([]models.CustomerActivity, error) {
	return s.mongo.GetCustomerActivities(ctx, orgID)
}

func (s *RFMStorage) GetRFMScoresBySegment(ctx context.Context, orgID, segment string) ([]models.RFMScore, error) {
	return s.mongo.GetRFMScoresBySegment(ctx, orgID, segment)
}

func (s *RFMStorage) GetRFMScoresByLocation(ctx context.Context, orgID, locationID string) ([]models.RFMScore, error) {
	return s.mongo.GetRFMScoresByLocation(ctx, orgID, locationID)
}

func (s *RFMStorage) GetCustomerActivitiesByLocation(ctx context.Context, orgID, locationID string) ([]models.CustomerActivity, error) {
	return s.mongo.GetCustomerActivitiesByLocation(ctx, orgID, locationID)
}