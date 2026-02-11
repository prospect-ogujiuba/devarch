package handlers

import (
	"crypto/subtle"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/priz/devarch-api/internal/security"
)

type AuthHandler struct{}

func NewAuthHandler() *AuthHandler {
	return &AuthHandler{}
}

func (h *AuthHandler) Validate(w http.ResponseWriter, r *http.Request) {
	apiKey := os.Getenv("DEVARCH_API_KEY")
	if apiKey == "" {
		w.WriteHeader(http.StatusOK)
		return
	}

	provided := r.Header.Get("X-API-Key")
	if subtle.ConstantTimeCompare([]byte(provided), []byte(apiKey)) != 1 {
		http.Error(w, "invalid api key", http.StatusUnauthorized)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *AuthHandler) WSToken(w http.ResponseWriter, r *http.Request) {
	apiKey := os.Getenv("DEVARCH_API_KEY")
	if apiKey == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"token": ""})
		return
	}

	provided := r.Header.Get("X-API-Key")
	if subtle.ConstantTimeCompare([]byte(provided), []byte(apiKey)) != 1 {
		http.Error(w, "invalid api key", http.StatusUnauthorized)
		return
	}

	token, err := security.GenerateWSToken([]byte(apiKey), 60*time.Second)
	if err != nil {
		http.Error(w, "failed to generate token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"token": token})
}
