package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// MMEConfig holds all MME-related configuration
type MMEConfig struct {
	MaxEdgesPerTag    int     `json:"maxEdgesPerTag"`
	MaxDepth          int     `json:"maxDepth"`
	BeamWidth         int     `json:"beamWidth"`
	MaxSeedTags       int     `json:"maxSeedTags"`
	DecayAlpha        float64 `json:"decayAlpha"`
	MinActivation     float64 `json:"minActivation"`
	TokenBudget       int     `json:"tokenBudget"`
	DiversityLambda   float64 `json:"diversityLambda"`
	RecencyTauDays    int     `json:"recencyTauDays"`
	CacheTTLSeconds   int     `json:"cacheTTLSeconds"`
}

// Global MME configuration instance
var MME MMEConfig

// LoadEnv loads environment variables from .env file
func LoadEnv() {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, using system environment variables")
	}
}

// LoadMMEConfig loads MME configuration from environment variables
func LoadMMEConfig() {
	MME = MMEConfig{
		MaxEdgesPerTag:    getEnvInt("MME_MAX_EDGES_PER_TAG", 32),
		MaxDepth:          getEnvInt("MME_MAX_DEPTH", 2),
		BeamWidth:         getEnvInt("MME_BEAM_WIDTH", 128),
		MaxSeedTags:       getEnvInt("MME_MAX_SEED_TAGS", 5),
		DecayAlpha:        getEnvFloat("MME_DECAY_ALPHA", 0.85),
		MinActivation:     getEnvFloat("MME_MIN_ACTIVATION", 0.05),
		TokenBudget:       getEnvInt("MME_TOKEN_BUDGET", 2048),
		DiversityLambda:   getEnvFloat("MME_DIVERSITY_LAMBDA", 0.15),
		RecencyTauDays:    getEnvInt("MME_RECENCY_TAU_DAYS", 60),
		CacheTTLSeconds:   getEnvInt("MME_CACHE_TTL_SECS", 300),
	}

	log.Printf("MME Config loaded: MaxEdgesPerTag=%d, MaxDepth=%d, BeamWidth=%d, MaxSeedTags=%d, DecayAlpha=%.2f, MinActivation=%.2f, TokenBudget=%d, DiversityLambda=%.2f, RecencyTauDays=%d, CacheTTLSeconds=%d",
		MME.MaxEdgesPerTag, MME.MaxDepth, MME.BeamWidth, MME.MaxSeedTags, MME.DecayAlpha, MME.MinActivation, MME.TokenBudget, MME.DiversityLambda, MME.RecencyTauDays, MME.CacheTTLSeconds)
}

// Helper functions to get environment variables with defaults
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
		log.Printf("Warning: Invalid integer value for %s, using default %d", key, defaultValue)
	}
	return defaultValue
}

func getEnvFloat(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if floatValue, err := strconv.ParseFloat(value, 64); err == nil {
			return floatValue
		}
		log.Printf("Warning: Invalid float value for %s, using default %.2f", key, defaultValue)
	}
	return defaultValue
}

// GetMMEConfig returns the current MME configuration
func GetMMEConfig() MMEConfig {
	return MME
}
