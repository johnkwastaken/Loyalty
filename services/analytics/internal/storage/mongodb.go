package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/loyalty/analytics/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoStorage struct {
	client   *mongo.Client
	database *mongo.Database
}

func NewMongoStorage(uri, dbName string) (*MongoStorage, error) {
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	if err := client.Ping(context.TODO(), nil); err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	database := client.Database(dbName)
	
	storage := &MongoStorage{
		client:   client,
		database: database,
	}

	if err := storage.createIndexes(); err != nil {
		return nil, fmt.Errorf("failed to create indexes: %w", err)
	}

	return storage, nil
}

func (s *MongoStorage) createIndexes() error {
	ctx := context.Background()

	rfmCollection := s.database.Collection("rfm_scores")
	quintilesCollection := s.database.Collection("rfm_quintiles")
	activitiesCollection := s.database.Collection("customer_activities")
	tiersCollection := s.database.Collection("customer_tiers")
	tierConfigsCollection := s.database.Collection("tier_configs")
	tierUpgradesCollection := s.database.Collection("tier_upgrades")

	rfmIndexes := []mongo.IndexModel{
		{Keys: bson.D{{"org_id", 1}, {"location_id", 1}, {"customer_id", 1}}, Options: options.Index().SetUnique(true).SetName("rfm_org_location_customer_unique")},
		{Keys: bson.D{{"org_id", 1}, {"location_id", 1}}, Options: options.Index().SetName("rfm_org_location")},
		{Keys: bson.D{{"org_id", 1}, {"rfm_segment", 1}}, Options: options.Index().SetName("rfm_org_segment")},
		{Keys: bson.D{{"calculated_at", -1}}, Options: options.Index().SetName("rfm_calculated_at")},
	}

	quintilesIndexes := []mongo.IndexModel{
		{Keys: bson.D{{"org_id", 1}}, Options: options.Index().SetUnique(true)},
	}

	activityIndexes := []mongo.IndexModel{
		{Keys: bson.D{{"org_id", 1}, {"location_id", 1}, {"customer_id", 1}}, Options: options.Index().SetUnique(true).SetName("activity_org_location_customer_unique")},
		{Keys: bson.D{{"org_id", 1}, {"location_id", 1}}, Options: options.Index().SetName("activity_org_location")},
		{Keys: bson.D{{"org_id", 1}}, Options: options.Index().SetName("activity_org")},
		{Keys: bson.D{{"last_transaction", -1}}, Options: options.Index().SetName("activity_last_transaction")},
	}

	tierIndexes := []mongo.IndexModel{
		{Keys: bson.D{{"org_id", 1}, {"location_id", 1}, {"customer_id", 1}}, Options: options.Index().SetUnique(true).SetName("tier_org_location_customer_unique")},
		{Keys: bson.D{{"org_id", 1}, {"location_id", 1}}, Options: options.Index().SetName("tier_org_location")},
		{Keys: bson.D{{"org_id", 1}, {"current_tier", 1}}, Options: options.Index().SetName("tier_org_tier")},
		{Keys: bson.D{{"org_id", 1}, {"location_id", 1}, {"current_tier", 1}}, Options: options.Index().SetName("tier_org_location_tier")},
		{Keys: bson.D{{"tier_since", -1}}, Options: options.Index().SetName("tier_since")},
	}

	tierConfigIndexes := []mongo.IndexModel{
		{Keys: bson.D{{"org_id", 1}}, Options: options.Index().SetUnique(true)},
	}

	tierUpgradeIndexes := []mongo.IndexModel{
		{Keys: bson.D{{"org_id", 1}, {"customer_id", 1}}},
		{Keys: bson.D{{"org_id", 1}, {"notified", 1}}},
		{Keys: bson.D{{"upgraded_at", -1}}},
	}

	if _, err := rfmCollection.Indexes().CreateMany(ctx, rfmIndexes); err != nil {
		return err
	}

	if _, err := quintilesCollection.Indexes().CreateMany(ctx, quintilesIndexes); err != nil {
		return err
	}

	if _, err := activitiesCollection.Indexes().CreateMany(ctx, activityIndexes); err != nil {
		return err
	}

	if _, err := tiersCollection.Indexes().CreateMany(ctx, tierIndexes); err != nil {
		return err
	}

	if _, err := tierConfigsCollection.Indexes().CreateMany(ctx, tierConfigIndexes); err != nil {
		return err
	}

	if _, err := tierUpgradesCollection.Indexes().CreateMany(ctx, tierUpgradeIndexes); err != nil {
		return err
	}

	return nil
}

func (s *MongoStorage) SaveRFMScore(ctx context.Context, score models.RFMScore) error {
	collection := s.database.Collection("rfm_scores")
	
	filter := bson.M{
		"org_id":      score.OrgID,
		"location_id": score.LocationID,
		"customer_id": score.CustomerID,
	}
	
	update := bson.M{
		"$set": score,
	}
	
	opts := options.Update().SetUpsert(true)
	_, err := collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return fmt.Errorf("failed to save RFM score: %w", err)
	}
	
	return nil
}

func (s *MongoStorage) GetRFMScore(ctx context.Context, orgID, customerID string) (*models.RFMScore, error) {
	collection := s.database.Collection("rfm_scores")
	
	filter := bson.M{
		"org_id":      orgID,
		"customer_id": customerID,
	}
	
	var score models.RFMScore
	err := collection.FindOne(ctx, filter).Decode(&score)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("RFM score not found")
		}
		return nil, fmt.Errorf("failed to get RFM score: %w", err)
	}
	
	return &score, nil
}

func (s *MongoStorage) SaveQuintiles(ctx context.Context, quintiles models.RFMQuintiles) error {
	collection := s.database.Collection("rfm_quintiles")
	
	filter := bson.M{"org_id": quintiles.OrgID}
	update := bson.M{"$set": quintiles}
	opts := options.Update().SetUpsert(true)
	
	_, err := collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return fmt.Errorf("failed to save quintiles: %w", err)
	}
	
	return nil
}

func (s *MongoStorage) GetQuintiles(ctx context.Context, orgID string) (*models.RFMQuintiles, error) {
	collection := s.database.Collection("rfm_quintiles")
	
	var quintiles models.RFMQuintiles
	err := collection.FindOne(ctx, bson.M{"org_id": orgID}).Decode(&quintiles)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("quintiles not found")
		}
		return nil, fmt.Errorf("failed to get quintiles: %w", err)
	}
	
	return &quintiles, nil
}

func (s *MongoStorage) UpdateCustomerActivity(ctx context.Context, activity models.CustomerActivity) error {
	collection := s.database.Collection("customer_activities")
	
	filter := bson.M{
		"org_id":      activity.OrgID,
		"location_id": activity.LocationID,
		"customer_id": activity.CustomerID,
	}
	
	update := bson.M{
		"$set": bson.M{
			"org_id":               activity.OrgID,
			"location_id":          activity.LocationID,
			"customer_id":          activity.CustomerID,
			"last_transaction":     activity.LastTransaction,
			"total_transactions":   activity.TotalTransactions,
			"total_spent":          activity.TotalSpent,
			"updated_at":           time.Now(),
		},
		"$setOnInsert": bson.M{
			"first_transaction": activity.FirstTransaction,
			"created_at":        time.Now(),
		},
	}
	
	opts := options.Update().SetUpsert(true)
	_, err := collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return fmt.Errorf("failed to update customer activity: %w", err)
	}
	
	return nil
}

func (s *MongoStorage) GetCustomerActivities(ctx context.Context, orgID string) ([]models.CustomerActivity, error) {
	collection := s.database.Collection("customer_activities")
	
	cursor, err := collection.Find(ctx, bson.M{"org_id": orgID})
	if err != nil {
		return nil, fmt.Errorf("failed to find customer activities: %w", err)
	}
	defer cursor.Close(ctx)
	
	var activities []models.CustomerActivity
	for cursor.Next(ctx) {
		var activity models.CustomerActivity
		if err := cursor.Decode(&activity); err != nil {
			return nil, fmt.Errorf("failed to decode activity: %w", err)
		}
		activities = append(activities, activity)
	}
	
	return activities, nil
}

func (s *MongoStorage) GetRFMScoresBySegment(ctx context.Context, orgID, segment string) ([]models.RFMScore, error) {
	collection := s.database.Collection("rfm_scores")
	
	filter := bson.M{
		"org_id":      orgID,
		"rfm_segment": segment,
	}
	
	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to find RFM scores: %w", err)
	}
	defer cursor.Close(ctx)
	
	var scores []models.RFMScore
	for cursor.Next(ctx) {
		var score models.RFMScore
		if err := cursor.Decode(&score); err != nil {
			return nil, fmt.Errorf("failed to decode RFM score: %w", err)
		}
		scores = append(scores, score)
	}
	
	return scores, nil
}

func (s *MongoStorage) GetClient() *mongo.Client {
	return s.client
}

func (s *MongoStorage) GetDatabase() *mongo.Database {
	return s.database
}

func (s *MongoStorage) GetRFMScoreByLocation(ctx context.Context, orgID, locationID, customerID string) (*models.RFMScore, error) {
	collection := s.database.Collection("rfm_scores")
	
	filter := bson.M{
		"org_id":      orgID,
		"location_id": locationID,
		"customer_id": customerID,
	}
	
	var score models.RFMScore
	err := collection.FindOne(ctx, filter).Decode(&score)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("RFM score not found for location")
		}
		return nil, fmt.Errorf("failed to get RFM score by location: %w", err)
	}
	
	return &score, nil
}

func (s *MongoStorage) GetRFMScoresByLocation(ctx context.Context, orgID, locationID string) ([]models.RFMScore, error) {
	collection := s.database.Collection("rfm_scores")
	
	filter := bson.M{
		"org_id":      orgID,
		"location_id": locationID,
	}
	
	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to find RFM scores by location: %w", err)
	}
	defer cursor.Close(ctx)
	
	var scores []models.RFMScore
	for cursor.Next(ctx) {
		var score models.RFMScore
		if err := cursor.Decode(&score); err != nil {
			return nil, fmt.Errorf("failed to decode RFM score: %w", err)
		}
		scores = append(scores, score)
	}
	
	return scores, nil
}

func (s *MongoStorage) GetCustomerActivitiesByLocation(ctx context.Context, orgID, locationID string) ([]models.CustomerActivity, error) {
	collection := s.database.Collection("customer_activities")
	
	filter := bson.M{
		"org_id":      orgID,
		"location_id": locationID,
	}
	
	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to find customer activities by location: %w", err)
	}
	defer cursor.Close(ctx)
	
	var activities []models.CustomerActivity
	for cursor.Next(ctx) {
		var activity models.CustomerActivity
		if err := cursor.Decode(&activity); err != nil {
			return nil, fmt.Errorf("failed to decode activity: %w", err)
		}
		activities = append(activities, activity)
	}
	
	return activities, nil
}

func (s *MongoStorage) Close() error {
	return s.client.Disconnect(context.TODO())
}