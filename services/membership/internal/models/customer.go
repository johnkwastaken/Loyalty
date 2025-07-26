package models

import (
	"time"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Customer struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	CustomerID   string            `bson:"customer_id" json:"customer_id"`
	OrgID        string            `bson:"org_id" json:"org_id"`
	Email        string            `bson:"email" json:"email"`
	Phone        string            `bson:"phone" json:"phone"`
	FirstName    string            `bson:"first_name" json:"first_name"`
	LastName     string            `bson:"last_name" json:"last_name"`
	DateOfBirth  *time.Time        `bson:"date_of_birth,omitempty" json:"date_of_birth,omitempty"`
	Address      Address           `bson:"address" json:"address"`
	Preferences  CustomerPrefs     `bson:"preferences" json:"preferences"`
	Tier         string            `bson:"tier" json:"tier"`
	Status       string            `bson:"status" json:"status"`
	Metadata     map[string]any    `bson:"metadata" json:"metadata"`
	CreatedAt    time.Time         `bson:"created_at" json:"created_at"`
	UpdatedAt    time.Time         `bson:"updated_at" json:"updated_at"`
}

type Address struct {
	Street  string `bson:"street" json:"street"`
	City    string `bson:"city" json:"city"`
	State   string `bson:"state" json:"state"`
	ZipCode string `bson:"zip_code" json:"zip_code"`
	Country string `bson:"country" json:"country"`
}

type CustomerPrefs struct {
	EmailMarketing bool     `bson:"email_marketing" json:"email_marketing"`
	SMSMarketing   bool     `bson:"sms_marketing" json:"sms_marketing"`
	Categories     []string `bson:"categories" json:"categories"`
	Language       string   `bson:"language" json:"language"`
}

type Location struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	LocationID  string            `bson:"location_id" json:"location_id"`
	OrgID       string            `bson:"org_id" json:"org_id"`
	Name        string            `bson:"name" json:"name"`
	Address     Address           `bson:"address" json:"address"`
	Manager     string            `bson:"manager" json:"manager"`
	Settings    LocationSettings  `bson:"settings" json:"settings"`
	Active      bool              `bson:"active" json:"active"`
	CreatedAt   time.Time         `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time         `bson:"updated_at" json:"updated_at"`
}

type LocationSettings struct {
	PointsMultiplier float64           `bson:"points_multiplier" json:"points_multiplier"`
	CustomRewards    []RewardThreshold `bson:"custom_rewards" json:"custom_rewards"`
	AllowStamps      bool              `bson:"allow_stamps" json:"allow_stamps"`
}

type CreateCustomerRequest struct {
	OrgID       string                 `json:"org_id" binding:"required"`
	Email       string                 `json:"email" binding:"required,email"`
	Phone       string                 `json:"phone"`
	FirstName   string                 `json:"first_name" binding:"required"`
	LastName    string                 `json:"last_name" binding:"required"`
	DateOfBirth *time.Time             `json:"date_of_birth"`
	Address     Address                `json:"address"`
	Preferences CustomerPrefs          `json:"preferences"`
	Metadata    map[string]interface{} `json:"metadata"`
}

type CreateLocationRequest struct {
	OrgID    string           `json:"org_id" binding:"required"`
	Name     string           `json:"name" binding:"required"`
	Address  Address          `json:"address"`
	Manager  string           `json:"manager"`
	Settings LocationSettings `json:"settings"`
}