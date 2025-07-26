package tiers

import (
	"context"
)

// TierStorageInterface defines the interface for tier storage operations
type TierStorageInterface interface {
	GetTierConfig(ctx context.Context, orgID string) (*OrgTierConfig, error)
	SaveTierConfig(ctx context.Context, config OrgTierConfig) error
	GetCustomerTier(ctx context.Context, orgID, customerID string) (*CustomerTier, error)
	SaveCustomerTier(ctx context.Context, tier CustomerTier) error
	SaveTierUpgrade(ctx context.Context, upgrade TierUpgrade) error
	GetTierUpgrades(ctx context.Context, orgID string, unnotifiedOnly bool) ([]TierUpgrade, error)
	MarkUpgradeNotified(ctx context.Context, upgradeID string) error
	GetCustomersByTier(ctx context.Context, orgID, tierName string) ([]CustomerTier, error)
	GetAllCustomerTiers(ctx context.Context, orgID string) ([]CustomerTier, error)
} 