package ramalama

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Model       string
	Port        int
	IdleTimeout time.Duration
	MaxTokens   int
	Network     string
	Store       string
	RuntimeArgs []string
}

func DefaultConfig() Config {
	return Config{
		Model:       "granite3.2:8b",
		Port:        11435,
		IdleTimeout: 15 * time.Minute,
		MaxTokens:   4096,
	}
}

func LoadConfig() Config {
	cfg := DefaultConfig()

	if v := os.Getenv("LLM_MODEL"); v != "" {
		cfg.Model = v
	}
	if v := os.Getenv("LLM_PORT"); v != "" {
		if p, err := strconv.Atoi(v); err == nil {
			cfg.Port = p
		}
	}
	if v := os.Getenv("LLM_IDLE_TIMEOUT"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			cfg.IdleTimeout = d
		}
	}
	if v := os.Getenv("LLM_MAX_TOKENS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			cfg.MaxTokens = n
		}
	}
	if v := os.Getenv("LLM_NETWORK"); v != "" {
		cfg.Network = v
	}
	if v := os.Getenv("RAMALAMA_STORE"); v != "" {
		cfg.Store = v
	}

	return cfg
}

func DefaultEmbedConfig() Config {
	return Config{
		Model:       "nomic-embed-text",
		Port:        11436,
		IdleTimeout: 15 * time.Minute,
		RuntimeArgs: []string{"--embeddings"},
	}
}

func LoadEmbedConfig() Config {
	cfg := DefaultEmbedConfig()

	if v := os.Getenv("EMBED_MODEL"); v != "" {
		cfg.Model = v
	}
	if v := os.Getenv("EMBED_PORT"); v != "" {
		if p, err := strconv.Atoi(v); err == nil {
			cfg.Port = p
		}
	}
	if v := os.Getenv("EMBED_IDLE_TIMEOUT"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			cfg.IdleTimeout = d
		}
	}
	if v := os.Getenv("LLM_NETWORK"); v != "" {
		cfg.Network = v
	}
	if v := os.Getenv("RAMALAMA_STORE"); v != "" {
		cfg.Store = v
	}

	return cfg
}
