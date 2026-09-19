package config_test

import (
	"os"
	"testing"

	"github.com/freel/backend/internal/config"
)

func TestConfig_IsDevelopmentOrTest(t *testing.T) {
	tests := []struct {
		name     string
		cfg      *config.Config
		expected bool
	}{
		{
			name:     "Explicit development AppEnv",
			cfg:      &config.Config{AppEnv: "development"},
			expected: true,
		},
		{
			name:     "Explicit test AppEnv",
			cfg:      &config.Config{AppEnv: "test"},
			expected: true,
		},
		{
			name:     "Uppercase DEVELOPMENT",
			cfg:      &config.Config{AppEnv: "DEVELOPMENT"},
			expected: true,
		},
		{
			name:     "Explicit production AppEnv",
			cfg:      &config.Config{AppEnv: "production"},
			expected: false,
		},
		{
			name:     "Explicit staging AppEnv",
			cfg:      &config.Config{AppEnv: "staging"},
			expected: false,
		},
		{
			name:     "Unknown environment sandbox",
			cfg:      &config.Config{AppEnv: "sandbox"},
			expected: false,
		},
		{
			name:     "Empty AppEnv and empty Environment",
			cfg:      &config.Config{AppEnv: "", Environment: ""},
			expected: false,
		},
		{
			name:     "Fallback to Environment field development",
			cfg:      &config.Config{AppEnv: "", Environment: "development"},
			expected: true,
		},
		{
			name:     "Fallback to Environment field production",
			cfg:      &config.Config{AppEnv: "", Environment: "production"},
			expected: false,
		},
		{
			name:     "Nil config",
			cfg:      nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.cfg.IsDevelopmentOrTest()
			if got != tt.expected {
				t.Errorf("IsDevelopmentOrTest() = %v, expected %v", got, tt.expected)
			}
		})
	}
}

func TestConfig_SafeEnvironmentLoading(t *testing.T) {
	// Save current env
	origAppEnv := os.Getenv("APP_ENV")
	origEnv := os.Getenv("ENV")
	origEnvironment := os.Getenv("ENVIRONMENT")
	defer func() {
		os.Setenv("APP_ENV", origAppEnv)
		os.Setenv("ENV", origEnv)
		os.Setenv("ENVIRONMENT", origEnvironment)
	}()

	// 1. Production must NOT default to development
	os.Setenv("APP_ENV", "production")
	os.Unsetenv("ENV")
	os.Unsetenv("ENVIRONMENT")
	cfg := config.LoadConfig()
	if cfg.AppEnv != "production" || cfg.Environment != "production" {
		t.Errorf("Expected production, got AppEnv=%s Environment=%s", cfg.AppEnv, cfg.Environment)
	}
	if cfg.IsDevelopmentOrTest() {
		t.Errorf("Production environment incorrectly evaluated as development/test")
	}

	// 2. Staging must NOT default to development
	os.Setenv("APP_ENV", "staging")
	cfgStaging := config.LoadConfig()
	if cfgStaging.IsDevelopmentOrTest() {
		t.Errorf("Staging environment incorrectly evaluated as development/test")
	}

	// 3. Unknown environment must NOT enable development/test
	os.Setenv("APP_ENV", "qa-preview")
	cfgUnknown := config.LoadConfig()
	if cfgUnknown.IsDevelopmentOrTest() {
		t.Errorf("Unknown environment qa-preview incorrectly evaluated as development/test")
	}
}
