package config

import (
	"ai-service-go/internals/utils"
	"os"
	"strconv"
)

// Config holds the application configuration
type Config struct {
	Port           string
	EmbeddingModel string
	VectorSize     int
	TextGenModel   string
	NChunks        int
	OpenaiAPIkey   string
	QdrantHost     string
	LayerInfoURL   string
	AllLayersURL   string
	SupportEmail   string
}

var AppConfig *Config

func LoadConfig() {
	config := Config{
		Port:           getEnvOrDefault("PORT", "8080"),
		EmbeddingModel: getEnvOrDefault("EMBEDMODEL", "text-embedding-3-small"),
		VectorSize:     1536,
		TextGenModel:   getEnvOrDefault("GENMODEL", "gpt-4o-mini"),
		NChunks:        getEnvIntOrDefault("N_CHUNKS", 5),
		OpenaiAPIkey:   getEnvOrDefault("OPENAI_API_KEY", ""),
		QdrantHost:     getEnvOrDefault("QDRANT_HOST", "qdrant"),
		LayerInfoURL:   getEnvOrDefault("LAYER_INFORMATION_URL", ""),
		AllLayersURL:   getEnvOrDefault("ALL_LAYERS_URL", ""),
		SupportEmail:   getEnvOrDefault("SUPPORT_EMAIL", ""),
	}
	utils.Logger.Info("Config loaded successfully ✅")
	AppConfig = &config
}

// Helper functions
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	utils.Logger.Warn("Env variable not found for: " + key + ", using default value: " + defaultValue)
	return defaultValue
}

func getEnvIntOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
