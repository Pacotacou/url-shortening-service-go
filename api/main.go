package main

import (
	"api/interfaces"
	"api/models"
	"api/routes"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Create db connection
	connector, err := connectToDB()
	if err != nil {
		fmt.Println("Error connecting to database:", err)
		os.Exit(1)
	}
	defer connector.Close()
	// Get db and repo
	db := connector.GetDB()
	urlRepo := models.NewURLRepository(db)
	// Create router
	router := gin.Default()
	// Load routes
	routes.LoadRoutes(router, urlRepo)
	router.Run()
}

func connectToDB() (*interfaces.PostgresConnector, error) {
	if err := godotenv.Load("../.env"); err != nil {
		return nil, fmt.Errorf("error loading .env")
	}

	port, err := strconv.Atoi(os.Getenv("DB_PORT"))
	if err != nil {
		return nil, fmt.Errorf("invalid PORT in .env file")
	}

	config := interfaces.DBConfig{
		Host:     os.Getenv("DB_HOST"),
		Port:     port,
		User:     os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASSWORD"),
		DBName:   os.Getenv("DB_NAME"),
		SSLMode:  "disable",
	}

	connector := interfaces.NewPostgresConnector(config)
	if err := connector.Connect(); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
		return nil, fmt.Errorf("failed to connect to database")
	}

	return connector, nil
}
