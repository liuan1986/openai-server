package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type RateLimitConfig struct {
	Capacity             int `json:"capacity"`
	RefillIntervalSecond int `json:"refill_interval_second"`
}

type OpenAIConfig struct {
	TargetURL string `json:"target_url"`
	APIKey    string `json:"api_key"`
}

type ServiceConfig struct {
	SecretKey string   `json:"secret_key"`
	Users     []string `json:"users"`
	Blacklist []string `json:"blacklist"`
}

type Config struct {
	Service   ServiceConfig   `json:"service"`
	OpenAI    OpenAIConfig    `json:"openai"`
	RateLimit RateLimitConfig `json:"rate_limit"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	if cfg.RateLimit.Capacity <= 0 {
		cfg.RateLimit.Capacity = 10
	}
	if cfg.RateLimit.RefillIntervalSecond <= 0 {
		cfg.RateLimit.RefillIntervalSecond = 60
	}

	return &cfg, nil
}
