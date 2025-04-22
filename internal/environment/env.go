package env

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Load loads environment variables from a .env file if available
func Load() {
	// Try to load .env file
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, using default environment variables")
	}

	// Set default environment if not set
	env := os.Getenv("ENV")
	if env == "" {
		env = "DEV"
		os.Setenv("ENV", env)
		log.Printf("Environment not set, defaulting to %s", env)
	} else {
		log.Printf("Using environment: %s", env)
	}
}

// GetEnvironment returns the current environment
func GetEnvironment() string {
	env := os.Getenv("ENV")
	if env == "" {
		return "DEV"
	}
	return env
}
