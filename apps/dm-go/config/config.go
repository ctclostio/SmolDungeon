package config

import (
	"os"
	"strconv"
)

// LLMConfig holds configuration for LLM integration
type LLMConfig struct {
	// Remote model settings
	BaseURL     string
	APIKey      string
	Model       string
	MaxTokens   int
	Temperature float32

	// Local model settings
	LocalEnabled     bool
	LocalBaseURL     string
	LocalModel       string
	LocalMaxTokens   int
	LocalTemperature float32

	// Model selection
	PreferredModel string // "remote", "local", or "auto"
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Port        string
	DBPath      string
	ScenarioDir string
	CORSOrigins string
	LLM         LLMConfig
}

// LoadServerConfig loads configuration from environment variables
func LoadServerConfig() ServerConfig {
	return ServerConfig{
		Port:        getEnv("PORT", "3000"),
		DBPath:      getEnv("DB_PATH", "./dm-server.db"),
		ScenarioDir: getEnv("SCENARIO_DIR", "../../scenarios"),
		CORSOrigins: getEnv("CORS_ORIGINS", "http://localhost:3000,http://localhost:5173"),
		LLM: LLMConfig{
			// Remote model settings
			BaseURL:     getEnv("LLM_BASE_URL", ""),
			APIKey:      getEnv("LLM_API_KEY", "dummy-key"),
			Model:       getEnv("LLM_MODEL", "gpt-3.5-turbo"),
			MaxTokens:   getEnvInt("LLM_MAX_TOKENS", 150),
			Temperature: getEnvFloat("LLM_TEMPERATURE", 0.7),

			// Local model settings
			LocalEnabled:     getEnvBool("LLM_LOCAL_ENABLED", false),
			LocalBaseURL:     getEnv("LLM_LOCAL_BASE_URL", "http://localhost:8000/v1"),
			LocalModel:       getEnv("LLM_LOCAL_MODEL", "gpt-oss-20b"),
			LocalMaxTokens:   getEnvInt("LLM_LOCAL_MAX_TOKENS", 200),
			LocalTemperature: getEnvFloat("LLM_LOCAL_TEMPERATURE", 0.8),

			// Model selection
			PreferredModel: getEnv("LLM_PREFERRED_MODEL", "auto"),
		},
	}
}

// Helper functions

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvFloat(key string, defaultValue float32) float32 {
	if value := os.Getenv(key); value != "" {
		if floatValue, err := strconv.ParseFloat(value, 32); err == nil {
			return float32(floatValue)
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}
