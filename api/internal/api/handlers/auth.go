package handlers

import (
	"crypto/subtle"
	"net/http"
	"os"
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
