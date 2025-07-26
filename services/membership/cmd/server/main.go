package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/loyalty/membership/internal/handlers"
	"github.com/loyalty/membership/internal/repository"
)

func main() {
	mongoURL := os.Getenv("MONGO_URL")
	if mongoURL == "" {
		mongoURL = "mongodb://admin:password@localhost:27017/loyalty?authSource=admin"
	}

	repo, err := repository.NewMongoRepo(mongoURL, "loyalty")
	if err != nil {
		log.Fatalf("Failed to create MongoDB repository: %v", err)
	}
	defer repo.Close()

	handler := handlers.NewMembershipHandler(repo)

	r := gin.Default()

	v1 := r.Group("/api/v1")
	{
		// Customer APIs
		v1.POST("/customers", handler.CreateCustomer)
		v1.GET("/customers/:id", handler.GetCustomer)
		v1.GET("/customers", handler.GetCustomersByOrg)
		v1.PATCH("/customers/:id", handler.UpdateCustomer)
		
		// Organization APIs
		v1.POST("/organizations", handler.CreateOrganization)
		v1.GET("/organizations/:id", handler.GetOrganization)
		
		// Location APIs
		v1.POST("/locations", handler.CreateLocation)
		v1.GET("/locations/:id", handler.GetLocation)
		v1.GET("/locations", handler.GetLocationsByOrg)
		v1.PATCH("/locations/:id", handler.UpdateLocation)
		v1.DELETE("/locations/:id", handler.DeactivateLocation)
		
		v1.GET("/health", handler.Health)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8002"
	}

	log.Printf("Starting membership service on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}