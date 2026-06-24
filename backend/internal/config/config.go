package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv              string
	Port                string
	FrontendURL         string
	FrontendProdURL     string
	AWSRegion           string
	CognitoUserPoolID   string
	CognitoClientID     string
	CognitoClientSecret string
}

func LoadConfig() *Config {
	// Load .env file if it exists, otherwise fall back to environment variables
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found or error loading it, relying on system environment variables.")
	}

	cfg := &Config{
		AppEnv:              getEnv("APP_ENV", "development"),
		Port:                getEnv("PORT", "8080"),
		FrontendURL:         getEnv("FRONTEND_URL", "http://localhost:5173"),
		FrontendProdURL:     getEnv("FRONTEND_PROD_URL", "https://freel-project.vercel.app"),
		AWSRegion:           getEnv("AWS_REGION", "ap-south-1"),
		CognitoUserPoolID:   getEnv("COGNITO_USER_POOL_ID", ""),
		CognitoClientID:     getEnv("COGNITO_CLIENT_ID", ""),
		CognitoClientSecret: getEnv("COGNITO_CLIENT_SECRET", ""),
	}

	return cfg
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
