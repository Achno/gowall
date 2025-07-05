package config

import (
	"os"
	"strconv"

	"github.com/Achno/gowall/internal/logger"
	"github.com/joho/godotenv"
)

type EnvConfig struct {
	// OpenAI & OpenAI compatible services
	OPENAI_BASE_URL                       string
	OPENAI_MAX_RETRIES                    int
	OPENAI_API_KEY                        string
	OPENROUTER_API_KEY                    string
	OPENAI_API_COMPATIBLE_SERVICE_API_KEY string

	// Ollama
	OLLAMA_HOST string

	// Docling
	DOCLING_BASE_URL string

	// Mistral
	MISTRAL_API_KEY string

	// Gemini
	GEMINI_API_KEY string
}

func GetEnvConfig(envFilePath string) *EnvConfig {

	err := godotenv.Load(envFilePath)
	if err != nil {
		// do nothing, because a .env is optional, and continue trying to load envs from the environment
	}

	// use the helper methods below and get them
	return &EnvConfig{
		OPENAI_BASE_URL:                       GetString("OPENAI_BASE_URL", ""),
		OPENAI_MAX_RETRIES:                    GetInt("OPENAI_MAX_RETRIES", 2),
		OPENAI_API_KEY:                        GetString("OPENAI_API_KEY", ""),
		OPENROUTER_API_KEY:                    GetString("OPENROUTER_API_KEY", ""),
		OPENAI_API_COMPATIBLE_SERVICE_API_KEY: GetString("OPENAI_API_COMPATIBLE_SERVICE_API_KEY", ""),
		OLLAMA_HOST:                           GetString("OLLAMA_HOST", ""),
		DOCLING_BASE_URL:                      GetString("DOCLING_BASE_URL", ""),
		MISTRAL_API_KEY:                       GetString("MISTRAL_API_KEY", ""),
		GEMINI_API_KEY:                        GetString("GEMINI_API_KEY", ""),
	}
}

func GetString(key string, defaultValue string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue
	}
	return value
}

func GetInt(key string, defaultValue int) int {
	value, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue
	}

	intValue, err := strconv.Atoi(value)
	if err != nil {
		logger.Fatalf("environment variable %s=%q cannot be converted to an int", key, value)
	}
	return intValue
}
