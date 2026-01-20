package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	// Server
	Port        string
	Environment string

	// LiveKit
	LiveKitURL       string
	LiveKitAPIKey    string
	LiveKitAPISecret string

	// Deepgram
	DeepgramAPIKey string

	// Cartesia
	CartesiaAPIKey string
	CartesiaVoiceID string

	// LLM (OpenAI or compatible)
	LLMProvider   string
	LLMAPIKey     string
	LLMBaseURL    string
	LLMModel      string

	// Avatar (Beyond Presence / Tavus)
	AvatarProvider   string
	AvatarAPIKey     string
	AvatarAvatarID   string

	// Supabase
	SupabaseURL    string
	SupabaseAPIKey string

	// Pricing (per minute/token for cost estimation)
	DeepgramPricePerMin  float64
	CartesiaPricePerChar float64
	LLMPricePerToken     float64
}

var AppConfig *Config

func Load() (*Config, error) {
	_ = godotenv.Load()

	deepgramPrice, _ := strconv.ParseFloat(getEnv("DEEPGRAM_PRICE_PER_MIN", "0.0043"), 64)
	cartesiaPrice, _ := strconv.ParseFloat(getEnv("CARTESIA_PRICE_PER_CHAR", "0.000015"), 64)
	llmPrice, _ := strconv.ParseFloat(getEnv("LLM_PRICE_PER_TOKEN", "0.00003"), 64)

	AppConfig = &Config{
		Port:        getEnv("PORT", "8080"),
		Environment: getEnv("ENVIRONMENT", "development"),

		LiveKitURL:       getEnv("LIVEKIT_URL", ""),
		LiveKitAPIKey:    getEnv("LIVEKIT_API_KEY", ""),
		LiveKitAPISecret: getEnv("LIVEKIT_API_SECRET", ""),

		DeepgramAPIKey: getEnv("DEEPGRAM_API_KEY", ""),

		CartesiaAPIKey:  getEnv("CARTESIA_API_KEY", ""),
		CartesiaVoiceID: getEnv("CARTESIA_VOICE_ID", "a0e99841-438c-4a64-b679-ae501e7d6091"),

		LLMProvider: getEnv("LLM_PROVIDER", "openai"),
		LLMAPIKey:   getEnv("LLM_API_KEY", ""),
		LLMBaseURL:  getEnv("LLM_BASE_URL", "https://api.openai.com/v1"),
		LLMModel:    getEnv("LLM_MODEL", "gpt-4o"),

		AvatarProvider: getEnv("AVATAR_PROVIDER", "tavus"),
		AvatarAPIKey:   getEnv("AVATAR_API_KEY", ""),
		AvatarAvatarID: getEnv("AVATAR_ID", ""),

		SupabaseURL:    getEnv("SUPABASE_URL", ""),
		SupabaseAPIKey: getEnv("SUPABASE_API_KEY", ""),

		DeepgramPricePerMin:  deepgramPrice,
		CartesiaPricePerChar: cartesiaPrice,
		LLMPricePerToken:     llmPrice,
	}

	return AppConfig, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
