package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

const (
	EnvMailboxEncryptionKey = "MAILBOX_ENCRYPTION_KEY"
	EnvGoogleClientID       = "GOOGLE_CLIENT_ID"
	EnvGoogleClientSecret   = "GOOGLE_CLIENT_SECRET"
	EnvGoogleRedirectURI    = "GOOGLE_REDIRECT_URI"
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
	AWSAccessKeyID      string
	AWSSecretAccessKey  string
	SESFromEmail        string
	S3Bucket            string
	AdminAPIKey         string

	// Stripe
	StripeSecretKey     string
	StripeWebhookSecret string

	// Notifications
	MailProvider        string
	SMTPHost            string
	SMTPPort            string
	SMTPUsername        string
	SMTPPassword        string

	// Connected Mailbox Encryption
	MailboxEncryptionKey string

	// Google OAuth Client Config
	GoogleClientID     string
	GoogleClientSecret string
	GoogleRedirectURI  string
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
		AWSAccessKeyID:      os.Getenv("AWS_ACCESS_KEY_ID"),
		AWSSecretAccessKey:  os.Getenv("AWS_SECRET_ACCESS_KEY"),
		SESFromEmail:        os.Getenv("SES_FROM_EMAIL"),
		S3Bucket:            os.Getenv("S3_BUCKET"),
		AdminAPIKey:         os.Getenv("ADMIN_API_KEY"),
		MailProvider:        os.Getenv("MAIL_PROVIDER"),
		SMTPHost:            os.Getenv("SMTP_HOST"),
		SMTPPort:            os.Getenv("SMTP_PORT"),
		SMTPUsername:        os.Getenv("SMTP_USERNAME"),
		SMTPPassword:        os.Getenv("SMTP_PASSWORD"),
		StripeSecretKey:     os.Getenv("STRIPE_SECRET_KEY"),
		StripeWebhookSecret: os.Getenv("STRIPE_WEBHOOK_SECRET"),
		MailboxEncryptionKey: os.Getenv(EnvMailboxEncryptionKey),
		GoogleClientID:     os.Getenv(EnvGoogleClientID),
		GoogleClientSecret: os.Getenv(EnvGoogleClientSecret),
		GoogleRedirectURI:  os.Getenv(EnvGoogleRedirectURI),
	}

	return cfg
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
