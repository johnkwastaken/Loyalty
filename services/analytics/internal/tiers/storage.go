package tiers

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type TierStorage struct {
	client   *mongo.Client
	database *mongo.Database
}

func NewTierStorage(client *mongo.Client, database *mongo.Database) *TierStorage {
	return &TierStorage{
		client:   client,
		database: database,
	}
}

func (s *TierStorage) SaveCustomerTier(ctx context.Context, tier CustomerTier) error {
	collection := s.database.Collection("customer_tiers")
	
	filter := bson.M{
		"org_id":      tier.OrgID,
		"location_id": tier.LocationID,
		"customer_id": tier.CustomerID,
	}
	
	update := bson.M{
		"$set": tier,
	}
	
	opts := options.Update().SetUpsert(true)
	_, err := collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return fmt.Errorf("failed to save customer tier: %w", err)
	}
	
	return nil
}

func (s *TierStorage) GetCustomerTier(ctx context.Context, orgID, customerID string) (*CustomerTier, error) {
	collection := s.database.Collection("customer_tiers")
	
	filter := bson.M{
		"org_id":      orgID,
		"customer_id": customerID,
	}
	
	var tier CustomerTier
	err := collection.FindOne(ctx, filter).Decode(&tier)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("customer tier not found")
		}
		return nil, fmt.Errorf("failed to get customer tier: %w", err)
	}
	
	return &tier, nil
}

func (s *TierStorage) SaveTierConfig(ctx context.Context, config OrgTierConfig) error {
	collection := s.database.Collection("tier_configs")
	
	filter := bson.M{"org_id": config.OrgID}
	update := bson.M{"$set": config}
	opts := options.Update().SetUpsert(true)
	
	_, err := collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return fmt.Errorf("failed to save tier config: %w", err)
	}
	
	return nil
}

func (s *TierStorage) GetTierConfig(ctx context.Context, orgID string) (*OrgTierConfig, error) {
	collection := s.database.Collection("tier_configs")
	
	var config OrgTierConfig
	err := collection.FindOne(ctx, bson.M{"org_id": orgID}).Decode(&config)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("tier config not found")
		}
		return nil, fmt.Errorf("failed to get tier config: %w", err)
	}
	
	return &config, nil
}

func (s *TierStorage) SaveTierUpgrade(ctx context.Context, upgrade TierUpgrade) error {
	collection := s.database.Collection("tier_upgrades")
	
	_, err := collection.InsertOne(ctx, upgrade)
	if err != nil {
		return fmt.Errorf("failed to save tier upgrade: %w", err)
	}
	
	return nil
}

func (s *TierStorage) GetTierUpgrades(ctx context.Context, orgID string, unnotifiedOnly bool) ([]TierUpgrade, error) {
	collection := s.database.Collection("tier_upgrades")
	
	filter := bson.M{"org_id": orgID}
	if unnotifiedOnly {
		filter["notified"] = false
	}
	
	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to find tier upgrades: %w", err)
	}
	defer cursor.Close(ctx)
	
	var upgrades []TierUpgrade
	for cursor.Next(ctx) {
		var upgrade TierUpgrade
		if err := cursor.Decode(&upgrade); err != nil {
			return nil, fmt.Errorf("failed to decode tier upgrade: %w", err)
		}
		upgrades = append(upgrades, upgrade)
	}
	
	return upgrades, nil
}

func (s *TierStorage) MarkUpgradeNotified(ctx context.Context, upgradeID string) error {
	collection := s.database.Collection("tier_upgrades")
	
	objID, err := primitive.ObjectIDFromHex(upgradeID)
	if err != nil {
		return fmt.Errorf("invalid upgrade ID: %w", err)
	}
	
	update := bson.M{
		"$set": bson.M{
			"notified":   true,
			"updated_at": time.Now(),
		},
	}
	
	_, err = collection.UpdateOne(ctx, bson.M{"_id": objID}, update)
	if err != nil {
		return fmt.Errorf("failed to mark upgrade as notified: %w", err)
	}
	
	return nil
}

func (s *TierStorage) GetCustomersByTier(ctx context.Context, orgID, tierName string) ([]CustomerTier, error) {
	collection := s.database.Collection("customer_tiers")
	
	filter := bson.M{
		"org_id":       orgID,
		"current_tier": tierName,
	}
	
	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to find customers by tier: %w", err)
	}
	defer cursor.Close(ctx)
	
	var customers []CustomerTier
	for cursor.Next(ctx) {
		var customer CustomerTier
		if err := cursor.Decode(&customer); err != nil {
			return nil, fmt.Errorf("failed to decode customer tier: %w", err)
		}
		customers = append(customers, customer)
	}
	
	return customers, nil
}

func (s *TierStorage) GetAllCustomerTiers(ctx context.Context, orgID string) ([]CustomerTier, error) {
	collection := s.database.Collection("customer_tiers")
	
	cursor, err := collection.Find(ctx, bson.M{"org_id": orgID})
	if err != nil {
		return nil, fmt.Errorf("failed to find customer tiers: %w", err)
	}
	defer cursor.Close(ctx)
	
	var customers []CustomerTier
	for cursor.Next(ctx) {
		var customer CustomerTier
		if err := cursor.Decode(&customer); err != nil {
			return nil, fmt.Errorf("failed to decode customer tier: %w", err)
		}
		customers = append(customers, customer)
	}
	
	return customers, nil
}

func (s *TierStorage) GetCustomerTierByLocation(ctx context.Context, orgID, locationID, customerID string) (*CustomerTier, error) {
	collection := s.database.Collection("customer_tiers")
	
	filter := bson.M{
		"org_id":      orgID,
		"location_id": locationID,
		"customer_id": customerID,
	}
	
	var tier CustomerTier
	err := collection.FindOne(ctx, filter).Decode(&tier)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("customer tier not found for location")
		}
		return nil, fmt.Errorf("failed to get customer tier by location: %w", err)
	}
	
	return &tier, nil
}

func (s *TierStorage) GetCustomersByLocation(ctx context.Context, orgID, locationID string) ([]CustomerTier, error) {
	collection := s.database.Collection("customer_tiers")
	
	filter := bson.M{
		"org_id":      orgID,
		"location_id": locationID,
	}
	
	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to find customers by location: %w", err)
	}
	defer cursor.Close(ctx)
	
	var customers []CustomerTier
	for cursor.Next(ctx) {
		var customer CustomerTier
		if err := cursor.Decode(&customer); err != nil {
			return nil, fmt.Errorf("failed to decode customer tier: %w", err)
		}
		customers = append(customers, customer)
	}
	
	return customers, nil
}

func (s *TierStorage) GetCustomersByTierAndLocation(ctx context.Context, orgID, locationID, tierName string) ([]CustomerTier, error) {
	collection := s.database.Collection("customer_tiers")
	
	filter := bson.M{
		"org_id":       orgID,
		"location_id":  locationID,
		"current_tier": tierName,
	}
	
	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to find customers by tier and location: %w", err)
	}
	defer cursor.Close(ctx)
	
	var customers []CustomerTier
	for cursor.Next(ctx) {
		var customer CustomerTier
		if err := cursor.Decode(&customer); err != nil {
			return nil, fmt.Errorf("failed to decode customer tier: %w", err)
		}
		customers = append(customers, customer)
	}
	
	return customers, nil
}