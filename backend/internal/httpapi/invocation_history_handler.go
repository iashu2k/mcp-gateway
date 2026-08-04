package httpapi

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/iashu2k/mcp-gateway/backend/internal/repository"
	"github.com/iashu2k/mcp-gateway/backend/internal/service"
)

type InvocationHistoryHandler struct {
	service *service.InvocationHistoryService
}

func NewInvocationHistoryHandler(
	service *service.InvocationHistoryService,
) *InvocationHistoryHandler {
	return &InvocationHistoryHandler{service: service}
}

func (h *InvocationHistoryHandler) List(w http.ResponseWriter, r *http.Request) {
	user, ok := AuthenticatedUserFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthenticated", "authentication is required", nil)
		return
	}

	query := r.URL.Query()
	serverID := query.Get("serverId")
	toolID := query.Get("toolId")
	status := query.Get("status")

	limit, _ := strconv.Atoi(query.Get("limit"))
	offset, _ := strconv.Atoi(query.Get("offset"))

	invocations, err := h.service.List(
		r.Context(),
		user,
		serverID,
		toolID,
		status,
		limit,
		offset,
	)
	if err != nil {
		h.handleError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"count": len(invocations),
		"data":  invocations,
	})
}

func (h *InvocationHistoryHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	user, ok := AuthenticatedUserFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthenticated", "authentication is required", nil)
		return
	}

	invocationID := chi.URLParam(r, "invocationID")

	inv, err := h.service.GetByID(r.Context(), user, invocationID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, inv)
}

func (h *InvocationHistoryHandler) handleError(w http.ResponseWriter, err error) {
	var validationError service.ValidationError

	switch {
	case errors.As(err, &validationError):
		writeError(w, http.StatusBadRequest, "validation_error", "request validation failed", validationError.Fields)

	case errors.Is(err, service.ErrInvalidInvocationID):
		writeError(w, http.StatusBadRequest, "invalid_invocation_id", "invocation ID must be a valid UUID", nil)

	case errors.Is(err, repository.ErrInvocationNotFound):
		writeError(w, http.StatusNotFound, "invocation_not_found", "invocation was not found", nil)

	case errors.Is(err, service.ErrInvocationForbidden):
		writeError(w, http.StatusForbidden, "invocation_forbidden", "you can only view your own invocations", nil)

	default:
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to retrieve invocation history", nil)
	}
}
