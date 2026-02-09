package crypto

import (
	"crypto/rand"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

const keySize = 32

func LoadOrGenerateKey() ([]byte, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	devarchDir := filepath.Join(home, ".devarch")
	keyPath := filepath.Join(devarchDir, "secret.key")

	if _, err := os.Stat(keyPath); err == nil {
		key, err := os.ReadFile(keyPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read key file: %w", err)
		}
		if len(key) != keySize {
			return nil, fmt.Errorf("invalid key size: expected %d bytes, got %d", keySize, len(key))
		}
		return key, nil
	}

	if err := os.MkdirAll(devarchDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create .devarch directory: %w", err)
	}

	key := make([]byte, keySize)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return nil, fmt.Errorf("failed to generate random key: %w", err)
	}

	if err := os.WriteFile(keyPath, key, 0600); err != nil {
		return nil, fmt.Errorf("failed to write key file: %w", err)
	}

	return key, nil
}
