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
	DatabaseURL         string
	OpenAIAPIKey        string
	GeminiAPIKey        string
}

func LoadConfig() *Config {
	// Load .env file if it exists, otherwise fall back to environment variables
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found or error loading it, relying on system environment variables.")
	}

	dbHost := getEnv("DB_HOST", "127.0.0.1")
	dbPort := getEnv("DB_PORT", "3306")
	dbUser := getEnv("DB_USER", "root")
	dbPassword := getEnv("DB_PASSWORD", "")
	dbName := getEnv("DB_NAME", "freel_mysql")

	var defaultMySQLDSN string
	if dbPassword != "" {
		defaultMySQLDSN = dbUser + ":" + dbPassword + "@tcp(" + dbHost + ":" + dbPort + ")/" + dbName + "?parseTime=true&loc=UTC&multiStatements=true"
	} else {
		defaultMySQLDSN = dbUser + "@tcp(" + dbHost + ":" + dbPort + ")/" + dbName + "?parseTime=true&loc=UTC&multiStatements=true"
	}

	cfg := &Config{
		AppEnv:              getEnv("APP_ENV", "development"),
		Port:                getEnv("PORT", "8080"),
		FrontendURL:         getEnv("FRONTEND_URL", "http://localhost:5173"),
		FrontendProdURL:     getEnv("FRONTEND_PROD_URL", "https://logisticshq.in"),
		AWSRegion:           getEnv("AWS_REGION", "ap-south-1"),
		CognitoUserPoolID:   getEnv("COGNITO_USER_POOL_ID", ""),
		CognitoClientID:     getEnv("COGNITO_CLIENT_ID", ""),
		CognitoClientSecret: getEnv("COGNITO_CLIENT_SECRET", ""),
		DatabaseURL:         getEnv("DB_URL", defaultMySQLDSN),
		OpenAIAPIKey:        getEnv("OPENAI_API_KEY", ""),
		GeminiAPIKey:        getEnv("GEMINI_API_KEY", ""),
	}

	return cfg
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
