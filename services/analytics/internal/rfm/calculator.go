package rfm

import (
	"context"
	"fmt"
	"log"
	"math"
	"sort"
	"time"

	"github.com/loyalty/analytics/internal/models"
)

type RFMCalculator struct {
	storage RFMStorageInterface
}

func NewRFMCalculator(storage RFMStorageInterface) *RFMCalculator {
	return &RFMCalculator{storage: storage}
}

func (c *RFMCalculator) ProcessCustomerTransaction(ctx context.Context, activity models.CustomerActivity) error {
	log.Printf("Processing RFM for customer %s in org %s at location %s", 
		activity.CustomerID, activity.OrgID, activity.LocationID)

	quintiles, err := c.storage.GetOrCalculateQuintiles(ctx, activity.OrgID)
	if err != nil {
		return fmt.Errorf("failed to get quintiles: %w", err)
	}

	rfmScore := c.calculateRFMScore(activity, quintiles)
	
	return c.storage.SaveRFMScore(ctx, rfmScore)
}

func (c *RFMCalculator) calculateRFMScore(activity models.CustomerActivity, quintiles models.RFMQuintiles) models.RFMScore {
	now := time.Now()
	
	daysSinceLast := int(now.Sub(activity.LastTransaction).Hours() / 24)
	daysSinceFirst := int(now.Sub(activity.FirstTransaction).Hours() / 24)
	avgOrderValue := activity.TotalSpent / float64(activity.TotalTransactions)
	
	recencyScore := c.getRecencyScore(daysSinceLast, quintiles.RecencyQuintiles)
	frequencyScore := c.getFrequencyScore(activity.TotalTransactions, quintiles.FrequencyQuintiles)
	monetaryScore := c.getMonetaryScore(activity.TotalSpent, quintiles.MonetaryQuintiles)
	
	segment := c.getRFMSegment(recencyScore, frequencyScore, monetaryScore)
	
	return models.RFMScore{
		OrgID:             activity.OrgID,
		LocationID:        activity.LocationID,
		CustomerID:        activity.CustomerID,
		RecencyScore:      recencyScore,
		FrequencyScore:    frequencyScore,
		MonetaryScore:     monetaryScore,
		RFMSegment:        segment,
		LastTransaction:   activity.LastTransaction,
		TotalTransactions: activity.TotalTransactions,
		TotalSpent:        activity.TotalSpent,
		AvgOrderValue:     avgOrderValue,
		DaysSinceFirst:    daysSinceFirst,
		DaysSinceLast:     daysSinceLast,
		CalculatedAt:      time.Now(),
		UpdatedAt:         time.Now(),
	}
}

func (c *RFMCalculator) getRecencyScore(daysSinceLast int, quintiles []int) int {
	for i, threshold := range quintiles {
		if daysSinceLast <= threshold {
			return 5 - i
		}
	}
	return 1
}

func (c *RFMCalculator) getFrequencyScore(totalTransactions int, quintiles []int) int {
	for i := len(quintiles) - 1; i >= 0; i-- {
		if totalTransactions >= quintiles[i] {
			return i + 1
		}
	}
	return 1
}

func (c *RFMCalculator) getMonetaryScore(totalSpent float64, quintiles []float64) int {
	for i := len(quintiles) - 1; i >= 0; i-- {
		if totalSpent >= quintiles[i] {
			return i + 1
		}
	}
	return 1
}

func (c *RFMCalculator) getRFMSegment(r, f, m int) string {
	key := fmt.Sprintf("%d%d%d", r, f, m)
	if segment, exists := RFMSegments[key]; exists {
		return segment
	}
	
	if r >= 4 && f >= 4 && m >= 4 {
		return "Champions"
	} else if r >= 3 && f >= 3 && m >= 3 {
		return "Loyal Customers"
	} else if r >= 3 && f <= 2 && m <= 2 {
		return "Potential Loyalists"
	} else if r >= 4 && f <= 2 && m <= 2 {
		return "New Customers"
	} else if r <= 2 && f >= 4 && m >= 4 {
		return "Cannot Lose Them"
	} else if r <= 2 && f >= 3 && m >= 3 {
		return "At Risk"
	} else if r <= 3 && f <= 2 && m >= 3 {
		return "Need Attention"
	} else if r <= 2 && f <= 2 && m >= 3 {
		return "About to Sleep"
	} else {
		return "Lost"
	}
}

func (c *RFMCalculator) CalculateQuintilesForOrg(ctx context.Context, orgID string) (models.RFMQuintiles, error) {
	activities, err := c.storage.GetCustomerActivities(ctx, orgID)
	if err != nil {
		return models.RFMQuintiles{}, fmt.Errorf("failed to get customer activities: %w", err)
	}

	if len(activities) < 5 {
		return c.getDefaultQuintiles(orgID), nil
	}

	var recencyDays []int
	var frequencies []int
	var monetaryValues []float64
	now := time.Now()

	for _, activity := range activities {
		daysSinceLast := int(now.Sub(activity.LastTransaction).Hours() / 24)
		recencyDays = append(recencyDays, daysSinceLast)
		frequencies = append(frequencies, activity.TotalTransactions)
		monetaryValues = append(monetaryValues, activity.TotalSpent)
	}

	return models.RFMQuintiles{
		OrgID:              orgID,
		RecencyQuintiles:   c.calculateIntQuintiles(recencyDays),
		FrequencyQuintiles: c.calculateIntQuintiles(frequencies),
		MonetaryQuintiles:  c.calculateFloatQuintiles(monetaryValues),
		CalculatedAt:       time.Now(),
	}, nil
}

func (c *RFMCalculator) calculateIntQuintiles(values []int) []int {
	sort.Ints(values)
	n := len(values)
	quintiles := make([]int, 5)
	
	for i := 0; i < 5; i++ {
		percentile := float64(i+1) * 0.2
		index := int(math.Ceil(percentile*float64(n))) - 1
		if index >= n {
			index = n - 1
		}
		quintiles[i] = values[index]
	}
	
	return quintiles
}

func (c *RFMCalculator) calculateFloatQuintiles(values []float64) []float64 {
	sort.Float64s(values)
	n := len(values)
	quintiles := make([]float64, 5)
	
	for i := 0; i < 5; i++ {
		percentile := float64(i+1) * 0.2
		index := int(math.Ceil(percentile*float64(n))) - 1
		if index >= n {
			index = n - 1
		}
		quintiles[i] = values[index]
	}
	
	return quintiles
}

func (c *RFMCalculator) getDefaultQuintiles(orgID string) models.RFMQuintiles {
	return models.RFMQuintiles{
		OrgID:              orgID,
		RecencyQuintiles:   []int{7, 30, 90, 180, 365},
		FrequencyQuintiles: []int{1, 2, 5, 10, 20},
		MonetaryQuintiles:  []float64{10.0, 25.0, 50.0, 100.0, 250.0},
		CalculatedAt:       time.Now(),
	}
}