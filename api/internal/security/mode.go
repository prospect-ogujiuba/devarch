package security

import (
	"fmt"
	"os"
	"strings"
)

// Mode represents a security profile that controls authentication behavior API-wide
type Mode string

const (
	// ModeDevOpen disables all authentication checks (dev-friendly default)
	ModeDevOpen Mode = "dev-open"
	// ModeDevKeyed validates API key on HTTP endpoints
	ModeDevKeyed Mode = "dev-keyed"
	// ModeStrict validates API key on HTTP endpoints and enforces token auth on WebSocket
	ModeStrict Mode = "strict"
)

// ParseMode parses a raw string into a Mode, returning an error for invalid values.
// Empty string defaults to dev-open for backward compatibility.
func ParseMode(raw string) (Mode, error) {
	raw = strings.TrimSpace(strings.ToLower(raw))
	if raw == "" {
		return ModeDevOpen, nil
	}

	switch Mode(raw) {
	case ModeDevOpen, ModeDevKeyed, ModeStrict:
		return Mode(raw), nil
	default:
		return "", fmt.Errorf("invalid security mode %q: must be dev-open, dev-keyed, or strict", raw)
	}
}

// RequiresAPIKey returns true if the mode requires API key validation on HTTP endpoints
func (m Mode) RequiresAPIKey() bool {
	return m == ModeDevKeyed || m == ModeStrict
}

// RequiresWSAuth returns true if the mode requires token authentication on WebSocket connections
func (m Mode) RequiresWSAuth() bool {
	return m == ModeStrict
}

// ValidateConfig checks if the environment is properly configured for the given mode
func ValidateConfig(mode Mode) error {
	if mode.RequiresAPIKey() {
		apiKey := os.Getenv("DEVARCH_API_KEY")
		if apiKey == "" {
			return fmt.Errorf("security mode %q requires DEVARCH_API_KEY to be set", mode)
		}
	}
	return nil
}
