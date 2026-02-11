package security

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// TokenPayload represents the JWT-like payload for WebSocket tokens
type TokenPayload struct {
	Exp int64 `json:"exp"`
}

// GenerateWSToken creates a signed token with the given TTL using HMAC-SHA256.
// Token format: {hex-encoded-payload}.{hex-encoded-hmac}
func GenerateWSToken(secret []byte, ttl time.Duration) (string, error) {
	payload := TokenPayload{
		Exp: time.Now().Add(ttl).Unix(),
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal payload: %w", err)
	}

	hexPayload := hex.EncodeToString(payloadJSON)

	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(hexPayload))
	hexMac := hex.EncodeToString(mac.Sum(nil))

	return fmt.Sprintf("%s.%s", hexPayload, hexMac), nil
}

// ValidateWSToken validates a token and checks expiry.
// Returns nil if valid, error otherwise.
func ValidateWSToken(token string, secret []byte) error {
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return fmt.Errorf("invalid token format")
	}

	hexPayload := parts[0]
	hexMac := parts[1]

	// Verify HMAC using constant-time comparison
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(hexPayload))
	expectedMac := mac.Sum(nil)
	providedMac, err := hex.DecodeString(hexMac)
	if err != nil {
		return fmt.Errorf("invalid mac encoding")
	}

	if !hmac.Equal(expectedMac, providedMac) {
		return fmt.Errorf("invalid token signature")
	}

	// Decode and unmarshal payload
	payloadJSON, err := hex.DecodeString(hexPayload)
	if err != nil {
		return fmt.Errorf("invalid payload encoding")
	}

	var payload TokenPayload
	if err := json.Unmarshal(payloadJSON, &payload); err != nil {
		return fmt.Errorf("invalid payload format")
	}

	// Check expiry
	if payload.Exp <= time.Now().Unix() {
		return fmt.Errorf("token expired")
	}

	return nil
}
