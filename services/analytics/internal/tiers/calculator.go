package tiers

import (
	"context"
	"fmt"
	"log"
	"sort"
	"time"
)

type TierCalculator struct {
	storage *TierStorage
}

func NewTierCalculator(storage *TierStorage) *TierCalculator {
	return &TierCalculator{storage: storage}
}

func (c *TierCalculator) ProcessCustomerMetrics(ctx context.Context, metrics CustomerMetrics) error {
	log.Printf("Processing tier calculation for customer %s in org %s at location %s", 
		metrics.CustomerID, metrics.OrgID, metrics.LocationID)

	tierConfig, err := c.storage.GetTierConfig(ctx, metrics.OrgID)
	if err != nil {
		log.Printf("No tier config found for org %s, using defaults", metrics.OrgID)
		tierConfig = &OrgTierConfig{
			OrgID:     metrics.OrgID,
			TierRules: GetDefaultTierRules(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		
		if saveErr := c.storage.SaveTierConfig(ctx, *tierConfig); saveErr != nil {
			log.Printf("Failed to save default tier config: %v", saveErr)
		}
	}

	currentTier, err := c.storage.GetCustomerTier(ctx, metrics.OrgID, metrics.CustomerID)
	if err != nil {
		currentTier = &CustomerTier{
			OrgID:       metrics.OrgID,
			LocationID:  metrics.LocationID,
			CustomerID:  metrics.CustomerID,
			CurrentTier: "Bronze",
			TierSince:   time.Now(),
		}
	}

	newTier := c.calculateTier(metrics, tierConfig.TierRules)
	
	updated := c.updateCustomerTier(currentTier, newTier, metrics)

	if err := c.storage.SaveCustomerTier(ctx, *updated); err != nil {
		return fmt.Errorf("failed to save customer tier: %w", err)
	}

	if updated.CurrentTier != updated.PreviousTier && updated.PreviousTier != "" {
		upgrade := TierUpgrade{
			OrgID:        metrics.OrgID,
			CustomerID:   metrics.CustomerID,
			FromTier:     updated.PreviousTier,
			ToTier:       updated.CurrentTier,
			TriggeredBy:  "transaction",
			TriggerValue: metrics.TransactionAmount,
			UpgradedAt:   time.Now(),
			Notified:     false,
		}
		
		if err := c.storage.SaveTierUpgrade(ctx, upgrade); err != nil {
			log.Printf("Failed to save tier upgrade: %v", err)
		} else {
			log.Printf("Customer %s upgraded from %s to %s", 
				metrics.CustomerID, updated.PreviousTier, updated.CurrentTier)
		}
	}

	return nil
}

func (c *TierCalculator) calculateTier(metrics CustomerMetrics, rules []TierRule) TierRule {
	sort.Slice(rules, func(i, j int) bool {
		return rules[i].Level > rules[j].Level
	})

	for _, rule := range rules {
		if c.meetsRequirements(metrics, rule) {
			return rule
		}
	}

	return rules[len(rules)-1]
}

func (c *TierCalculator) meetsRequirements(metrics CustomerMetrics, rule TierRule) bool {
	meetsSpentLifetime := metrics.TotalSpent >= rule.MinSpentLifetime
	meetsSpentYear := metrics.SpentThisYear >= rule.MinSpentYear
	meetsVisitsLifetime := metrics.TotalVisits >= rule.MinVisitsLifetime
	meetsVisitsYear := metrics.VisitsThisYear >= rule.MinVisitsYear

	return meetsSpentLifetime && meetsSpentYear && meetsVisitsLifetime && meetsVisitsYear
}

func (c *TierCalculator) updateCustomerTier(current *CustomerTier, newTier TierRule, metrics CustomerMetrics) *CustomerTier {
	now := time.Now()
	
	if current.CurrentTier != newTier.Name {
		current.PreviousTier = current.CurrentTier
		current.CurrentTier = newTier.Name
		current.TierSince = now
	}

	current.TotalSpent = metrics.TotalSpent
	current.TotalVisits = metrics.TotalVisits
	current.SpentThisYear = metrics.SpentThisYear
	current.VisitsThisYear = metrics.VisitsThisYear
	current.SpentThisMonth = metrics.SpentThisMonth
	current.VisitsThisMonth = metrics.VisitsThisMonth
	current.LastTransaction = metrics.LastTransaction
	current.PointsMultiplier = newTier.PointsMultiplier
	current.Benefits = newTier.Benefits
	current.CalculatedAt = now
	current.UpdatedAt = now

	nextTier, progress := c.calculateNextTierProgress(newTier, metrics)
	current.NextTier = nextTier
	current.ProgressToNext = progress

	return current
}

func (c *TierCalculator) calculateNextTierProgress(currentTier TierRule, metrics CustomerMetrics) (string, float64) {
	rules := GetDefaultTierRules()
	
	sort.Slice(rules, func(i, j int) bool {
		return rules[i].Level < rules[j].Level
	})

	for _, rule := range rules {
		if rule.Level > currentTier.Level {
			nextTier := rule
			
			spentProgress := 0.0
			visitsProgress := 0.0
			
			if nextTier.MinSpentYear > 0 {
				spentProgress = metrics.SpentThisYear / nextTier.MinSpentYear
			}
			
			if nextTier.MinVisitsYear > 0 {
				visitsProgress = float64(metrics.VisitsThisYear) / float64(nextTier.MinVisitsYear)
			}
			
			progress := (spentProgress + visitsProgress) / 2.0
			if progress > 1.0 {
				progress = 1.0
			}
			
			return nextTier.Name, progress
		}
	}
	
	return "", 1.0
}

func (c *TierCalculator) GetTierUpgrades(ctx context.Context, orgID string, unnotifiedOnly bool) ([]TierUpgrade, error) {
	return c.storage.GetTierUpgrades(ctx, orgID, unnotifiedOnly)
}

func (c *TierCalculator) MarkUpgradeNotified(ctx context.Context, upgradeID string) error {
	return c.storage.MarkUpgradeNotified(ctx, upgradeID)
}

func (c *TierCalculator) GetCustomersByTier(ctx context.Context, orgID, tierName string) ([]CustomerTier, error) {
	return c.storage.GetCustomersByTier(ctx, orgID, tierName)
}

func (c *TierCalculator) RecalculateAllTiers(ctx context.Context, orgID string) error {
	log.Printf("Recalculating all tiers for org %s", orgID)
	
	customers, err := c.storage.GetAllCustomerTiers(ctx, orgID)
	if err != nil {
		return fmt.Errorf("failed to get customers: %w", err)
	}

	for _, customer := range customers {
		metrics := CustomerMetrics{
			OrgID:            customer.OrgID,
			LocationID:       customer.LocationID,
			CustomerID:       customer.CustomerID,
			TotalSpent:       customer.TotalSpent,
			TotalVisits:      customer.TotalVisits,
			SpentThisYear:    customer.SpentThisYear,
			VisitsThisYear:   customer.VisitsThisYear,
			SpentThisMonth:   customer.SpentThisMonth,
			VisitsThisMonth:  customer.VisitsThisMonth,
			LastTransaction:  customer.LastTransaction,
		}

		if err := c.ProcessCustomerMetrics(ctx, metrics); err != nil {
			log.Printf("Failed to recalculate tier for customer %s: %v", customer.CustomerID, err)
		}
	}

	log.Printf("Completed tier recalculation for %d customers in org %s", len(customers), orgID)
	return nil
}