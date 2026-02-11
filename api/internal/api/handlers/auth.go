package handlers

import (
	"crypto/subtle"
	"net/http"
	"os"
	"time"

	"github.com/priz/devarch-api/internal/api/respond"
	"github.com/priz/devarch-api/internal/security"
)

type AuthHandler struct{}

func NewAuthHandler() *AuthHandler {
	return &AuthHandler{}
}

// Validate godoc
// @Summary      Validate API key
// @Tags         auth
// @Produce      json
// @Success      200 {object} respond.SuccessEnvelope{data=object}
// @Failure      401 {object} respond.ErrorEnvelope
// @Router       /auth/validate [post]
func (h *AuthHandler) Validate(w http.ResponseWriter, r *http.Request) {
	apiKey := os.Getenv("DEVARCH_API_KEY")
	if apiKey == "" {
		respond.JSON(w, r, http.StatusOK, map[string]bool{"valid": true})
		return
	}

	provided := r.Header.Get("X-API-Key")
	if subtle.ConstantTimeCompare([]byte(provided), []byte(apiKey)) != 1 {
		respond.Unauthorized(w, r, "invalid api key")
		return
	}

	respond.JSON(w, r, http.StatusOK, map[string]bool{"valid": true})
}

// WSToken godoc
// @Summary      Generate WebSocket authentication token
// @Tags         auth
// @Produce      json
// @Success      200 {object} respond.SuccessEnvelope{data=object}
// @Failure      401 {object} respond.ErrorEnvelope
// @Failure      500 {object} respond.ErrorEnvelope
// @Router       /auth/ws-token [post]
func (h *AuthHandler) WSToken(w http.ResponseWriter, r *http.Request) {
	apiKey := os.Getenv("DEVARCH_API_KEY")
	if apiKey == "" {
		respond.JSON(w, r, http.StatusOK, map[string]string{"token": ""})
		return
	}

	provided := r.Header.Get("X-API-Key")
	if subtle.ConstantTimeCompare([]byte(provided), []byte(apiKey)) != 1 {
		respond.Unauthorized(w, r, "invalid api key")
		return
	}

	token, err := security.GenerateWSToken([]byte(apiKey), 60*time.Second)
	if err != nil {
		respond.InternalError(w, r, err)
		return
	}

	respond.JSON(w, r, http.StatusOK, map[string]string{"token": token})
}
