package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/loyalty/membership/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoRepo struct {
	client   *mongo.Client
	database *mongo.Database
}

func NewMongoRepo(uri, dbName string) (*MongoRepo, error) {
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	if err := client.Ping(context.TODO(), nil); err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	database := client.Database(dbName)
	
	repo := &MongoRepo{
		client:   client,
		database: database,
	}

	if err := repo.createIndexes(); err != nil {
		return nil, fmt.Errorf("failed to create indexes: %w", err)
	}

	return repo, nil
}

func (r *MongoRepo) createIndexes() error {
	ctx := context.Background()

	customersCollection := r.database.Collection("customers")
	orgsCollection := r.database.Collection("organizations")
	locationsCollection := r.database.Collection("locations")

	customerIndexes := []mongo.IndexModel{
		{Keys: bson.D{{"customer_id", 1}}, Options: options.Index().SetUnique(true)},
		{Keys: bson.D{{"org_id", 1}, {"email", 1}}, Options: options.Index().SetUnique(true)},
		{Keys: bson.D{{"org_id", 1}}},
	}

	orgIndexes := []mongo.IndexModel{
		{Keys: bson.D{{"org_id", 1}}, Options: options.Index().SetUnique(true)},
	}

	locationIndexes := []mongo.IndexModel{
		{Keys: bson.D{{"location_id", 1}}, Options: options.Index().SetUnique(true)},
		{Keys: bson.D{{"org_id", 1}}},
	}

	if _, err := customersCollection.Indexes().CreateMany(ctx, customerIndexes); err != nil {
		return err
	}

	if _, err := orgsCollection.Indexes().CreateMany(ctx, orgIndexes); err != nil {
		return err
	}

	if _, err := locationsCollection.Indexes().CreateMany(ctx, locationIndexes); err != nil {
		return err
	}

	return nil
}

func (r *MongoRepo) CreateCustomer(ctx context.Context, req *models.CreateCustomerRequest) (*models.Customer, error) {
	customerID := primitive.NewObjectID().Hex()
	
	customer := &models.Customer{
		CustomerID:  customerID,
		OrgID:       req.OrgID,
		Email:       req.Email,
		Phone:       req.Phone,
		FirstName:   req.FirstName,
		LastName:    req.LastName,
		DateOfBirth: req.DateOfBirth,
		Address:     req.Address,
		Preferences: req.Preferences,
		Tier:        "bronze",
		Status:      "active",
		Metadata:    req.Metadata,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	collection := r.database.Collection("customers")
	result, err := collection.InsertOne(ctx, customer)
	if err != nil {
		return nil, fmt.Errorf("failed to create customer: %w", err)
	}

	customer.ID = result.InsertedID.(primitive.ObjectID)
	return customer, nil
}

func (r *MongoRepo) GetCustomer(ctx context.Context, customerID string) (*models.Customer, error) {
	collection := r.database.Collection("customers")
	
	var customer models.Customer
	err := collection.FindOne(ctx, bson.M{"customer_id": customerID}).Decode(&customer)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("customer not found")
		}
		return nil, fmt.Errorf("failed to get customer: %w", err)
	}

	return &customer, nil
}

func (r *MongoRepo) GetCustomersByOrg(ctx context.Context, orgID string, limit, offset int) ([]*models.Customer, error) {
	collection := r.database.Collection("customers")
	
	opts := options.Find().
		SetLimit(int64(limit)).
		SetSkip(int64(offset)).
		SetSort(bson.D{{"created_at", -1}})

	cursor, err := collection.Find(ctx, bson.M{"org_id": orgID}, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find customers: %w", err)
	}
	defer cursor.Close(ctx)

	var customers []*models.Customer
	for cursor.Next(ctx) {
		var customer models.Customer
		if err := cursor.Decode(&customer); err != nil {
			return nil, fmt.Errorf("failed to decode customer: %w", err)
		}
		customers = append(customers, &customer)
	}

	return customers, nil
}

func (r *MongoRepo) UpdateCustomer(ctx context.Context, customerID string, updates bson.M) error {
	collection := r.database.Collection("customers")
	
	updates["updated_at"] = time.Now()
	
	result, err := collection.UpdateOne(
		ctx,
		bson.M{"customer_id": customerID},
		bson.M{"$set": updates},
	)
	if err != nil {
		return fmt.Errorf("failed to update customer: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("customer not found")
	}

	return nil
}

func (r *MongoRepo) CreateOrganization(ctx context.Context, org *models.Organization) error {
	org.CreatedAt = time.Now()
	org.UpdatedAt = time.Now()
	
	collection := r.database.Collection("organizations")
	result, err := collection.InsertOne(ctx, org)
	if err != nil {
		return fmt.Errorf("failed to create organization: %w", err)
	}

	org.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *MongoRepo) GetOrganization(ctx context.Context, orgID string) (*models.Organization, error) {
	collection := r.database.Collection("organizations")
	
	var org models.Organization
	err := collection.FindOne(ctx, bson.M{"org_id": orgID}).Decode(&org)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("organization not found")
		}
		return nil, fmt.Errorf("failed to get organization: %w", err)
	}

	return &org, nil
}

// Location Management Methods

func (r *MongoRepo) CreateLocation(ctx context.Context, req *models.CreateLocationRequest) (*models.Location, error) {
	locationID := primitive.NewObjectID().Hex()
	
	location := &models.Location{
		LocationID: locationID,
		OrgID:      req.OrgID,
		Name:       req.Name,
		Address:    req.Address,
		Manager:    req.Manager,
		Settings:   req.Settings,
		Active:     true,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	collection := r.database.Collection("locations")
	result, err := collection.InsertOne(ctx, location)
	if err != nil {
		return nil, fmt.Errorf("failed to create location: %w", err)
	}

	location.ID = result.InsertedID.(primitive.ObjectID)
	return location, nil
}

func (r *MongoRepo) GetLocation(ctx context.Context, locationID string) (*models.Location, error) {
	collection := r.database.Collection("locations")
	
	var location models.Location
	err := collection.FindOne(ctx, bson.M{"location_id": locationID}).Decode(&location)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("location not found")
		}
		return nil, fmt.Errorf("failed to get location: %w", err)
	}

	return &location, nil
}

func (r *MongoRepo) GetLocationsByOrg(ctx context.Context, orgID string, limit, offset int) ([]*models.Location, error) {
	collection := r.database.Collection("locations")
	
	opts := options.Find().
		SetLimit(int64(limit)).
		SetSkip(int64(offset)).
		SetSort(bson.D{{"created_at", -1}})

	cursor, err := collection.Find(ctx, bson.M{"org_id": orgID}, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find locations: %w", err)
	}
	defer cursor.Close(ctx)

	var locations []*models.Location
	for cursor.Next(ctx) {
		var location models.Location
		if err := cursor.Decode(&location); err != nil {
			return nil, fmt.Errorf("failed to decode location: %w", err)
		}
		locations = append(locations, &location)
	}

	return locations, nil
}

func (r *MongoRepo) UpdateLocation(ctx context.Context, locationID string, updates bson.M) error {
	collection := r.database.Collection("locations")
	
	updates["updated_at"] = time.Now()
	
	result, err := collection.UpdateOne(
		ctx,
		bson.M{"location_id": locationID},
		bson.M{"$set": updates},
	)
	if err != nil {
		return fmt.Errorf("failed to update location: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("location not found")
	}

	return nil
}

func (r *MongoRepo) Close() error {
	return r.client.Disconnect(context.TODO())
}