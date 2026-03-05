package ramalama

import (
	"database/sql"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Model       string
	Image       string
	Port        int
	IdleTimeout time.Duration
	MaxTokens   int
	GPUMode     string
	Network     string
}

func DefaultConfig() Config {
	return Config{
		Model:       "granite3.2:8b",
		Image:       "quay.io/ramalama/ramalama:latest",
		Port:        11435,
		IdleTimeout: 15 * time.Minute,
		MaxTokens:   4096,
		GPUMode:     "auto",
	}
}

func LoadConfig(db *sql.DB) Config {
	cfg := DefaultConfig()

	if v := os.Getenv("LLM_MODEL"); v != "" {
		cfg.Model = v
	}
	if v := os.Getenv("LLM_IMAGE"); v != "" {
		cfg.Image = v
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
	if v := os.Getenv("LLM_GPU_MODE"); v != "" {
		cfg.GPUMode = v
	}
	if v := os.Getenv("LLM_NETWORK"); v != "" {
		cfg.Network = v
	}

	// Override from DB if ai_config table exists
	if db != nil {
		overrideFromDB(db, &cfg)
	}

	return cfg
}

func overrideFromDB(db *sql.DB, cfg *Config) {
	rows, err := db.Query("SELECT key, value FROM ai_config")
	if err != nil {
		return // table may not exist yet
	}
	defer rows.Close()

	for rows.Next() {
		var k, v string
		if err := rows.Scan(&k, &v); err != nil {
			continue
		}
		switch k {
		case "model":
			cfg.Model = v
		case "image":
			cfg.Image = v
		case "port":
			if p, err := strconv.Atoi(v); err == nil {
				cfg.Port = p
			}
		case "idle_timeout":
			if d, err := time.ParseDuration(v); err == nil {
				cfg.IdleTimeout = d
			}
		case "max_tokens":
			if n, err := strconv.Atoi(v); err == nil {
				cfg.MaxTokens = n
			}
		case "gpu_mode":
			cfg.GPUMode = v
		case "network":
			cfg.Network = v
		}
	}
}

func DetectGPU(mode string) []string {
	if mode == "cpu" {
		return nil
	}

	if mode == "nvidia" || (mode == "auto" && hasNvidiaGPU()) {
		return []string{"--device", "nvidia.com/gpu=all"}
	}
	if mode == "amd" || (mode == "auto" && hasAMDGPU()) {
		return []string{"--device", "/dev/kfd", "--device", "/dev/dri"}
	}

	return nil
}

func hasNvidiaGPU() bool {
	_, err := os.Stat("/dev/nvidia0")
	return err == nil
}

func hasAMDGPU() bool {
	_, err := os.Stat("/dev/kfd")
	return err == nil
}
