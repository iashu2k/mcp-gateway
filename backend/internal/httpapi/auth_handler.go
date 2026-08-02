package httpapi

import (
	"errors"
	"net/http"

	"github.com/iashu2k/mcp-gateway/backend/internal/domain"
	"github.com/iashu2k/mcp-gateway/backend/internal/service"
)

type AuthHandler struct {
	service *service.AuthService
}

func NewAuthHandler(service *service.AuthService) *AuthHandler {
	return &AuthHandler{service: service}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var request domain.LoginRequest

	if ok := decodeJSONBody(w, r, &request); !ok {
		return
	}

	response, err := h.service.Login(r.Context(), request)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			writeError(
				w,
				http.StatusUnauthorized,
				"invalid_credentials",
				"email or password is incorrect",
				nil,
			)
			return
		}

		writeError(
			w,
			http.StatusInternalServerError,
			"internal_error",
			"could not complete login",
			nil,
		)
		return
	}

	writeJSON(w, http.StatusOK, response)
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	user, ok := AuthenticatedUserFromContext(r.Context())
	if !ok {
		writeError(
			w,
			http.StatusUnauthorized,
			"unauthenticated",
			"authentication is required",
			nil,
		)
		return
	}

	writeJSON(w, http.StatusOK, user)
}
