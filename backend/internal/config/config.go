package config

import (
	"os"
	"strings"

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
	Environment         string
	Port                string
	FrontendURL         string
	FrontendProdURL     string
	SportalURL          string
	SportalProdURL      string
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
	InternalServiceToken string

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
	if err := godotenv.Load(); err != nil {
		if err2 := godotenv.Load("backend/.env"); err2 != nil {
			_ = godotenv.Load("../backend/.env")
		}
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

	// Safe environment resolution: check APP_ENV, ENV, ENVIRONMENT
	// Do NOT silently default missing or unknown environments to development.
	// Safe default is empty string "" (non-development / non-test).
	rawEnv := os.Getenv("APP_ENV")
	if rawEnv == "" {
		rawEnv = os.Getenv("ENV")
	}
	if rawEnv == "" {
		rawEnv = os.Getenv("ENVIRONMENT")
	}
	normEnv := strings.ToLower(strings.TrimSpace(rawEnv))

	cfg := &Config{
		AppEnv:              normEnv,
		Environment:         normEnv,
		Port:                getEnv("PORT", "8080"),
		FrontendURL:         getEnv("FRONTEND_URL", "http://localhost:5173"),
		FrontendProdURL:     getEnv("FRONTEND_PROD_URL", "https://app.logisticshq.in"),
		SportalURL:          getEnv("SPORTAL_URL", "http://localhost:5174"),
		SportalProdURL:      getEnv("SPORTAL_PROD_URL", "https://sportal.logisticshq.in"),
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
		InternalServiceToken: getEnv("INTERNAL_SERVICE_TOKEN", getEnv("INTERNAL_SERVICE_KEY", os.Getenv("AI_SIDECAR_SERVICE_KEY"))),
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

// IsDevelopmentOrTest returns true only if the environment is explicitly "development" or "test".
// Any other value, empty string, staging, production, or unknown returns false.
func (c *Config) IsDevelopmentOrTest() bool {
	if c == nil {
		return false
	}
	env := strings.ToLower(strings.TrimSpace(c.AppEnv))
	if env == "" {
		env = strings.ToLower(strings.TrimSpace(c.Environment))
	}
	return env == "development" || env == "test"
}
