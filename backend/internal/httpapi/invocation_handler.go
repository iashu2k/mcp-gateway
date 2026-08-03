package httpapi

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/iashu2k/mcp-gateway/backend/internal/domain"
	"github.com/iashu2k/mcp-gateway/backend/internal/repository"
	"github.com/iashu2k/mcp-gateway/backend/internal/service"
)

type InvocationHandler struct {
	service *service.InvocationService
}

func NewInvocationHandler(
	service *service.InvocationService,
) *InvocationHandler {
	return &InvocationHandler{service: service}
}

func (h *InvocationHandler) Invoke(w http.ResponseWriter, r *http.Request) {
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

	var request domain.InvokeToolRequest
	if ok := decodeJSONBody(w, r, &request); !ok {
		return
	}

	response, err := h.service.Invoke(
		r.Context(),
		user,
		chi.URLParam(r, "serverID"),
		chi.URLParam(r, "toolID"),
		request,
	)
	if err != nil {
		h.handleInvokeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, response)
}

func (h *InvocationHandler) handleInvokeError(
	w http.ResponseWriter,
	err error,
) {
	var validationError service.ValidationError

	switch {
	case errors.As(err, &validationError):
		writeError(
			w,
			http.StatusBadRequest,
			"validation_error",
			"request validation failed",
			validationError.Fields,
		)

	case errors.Is(err, service.ErrInvalidServerID):
		writeError(
			w,
			http.StatusBadRequest,
			"invalid_server_id",
			"serverID must be a valid UUID",
			nil,
		)

	case errors.Is(err, service.ErrInvalidToolID):
		writeError(
			w,
			http.StatusBadRequest,
			"invalid_tool_id",
			"toolID must be a valid UUID",
			nil,
		)

	case errors.Is(err, repository.ErrServerNotFound):
		writeError(
			w,
			http.StatusNotFound,
			"server_not_found",
			"MCP server was not found",
			nil,
		)

	case errors.Is(err, repository.ErrToolNotFound):
		writeError(
			w,
			http.StatusNotFound,
			"tool_not_found",
			"MCP tool was not found for this server",
			nil,
		)

	case errors.Is(err, service.ErrInvocationForbidden):
		writeError(
			w,
			http.StatusForbidden,
			"invocation_forbidden",
			"your role is not allowed to invoke tools",
			nil,
		)

	case errors.Is(err, service.ErrServerInactive):
		writeError(
			w,
			http.StatusConflict,
			"server_inactive",
			"MCP server must be active before a tool can be invoked",
			nil,
		)

	case errors.Is(err, service.ErrToolDisabled):
		writeError(
			w,
			http.StatusConflict,
			"tool_disabled",
			"MCP tool must be enabled before it can be invoked",
			nil,
		)

	case errors.Is(err, service.ErrToolRiskNotAllowed):
		writeError(
			w,
			http.StatusForbidden,
			"tool_risk_not_allowed",
			"only low-risk tools may be invoked in the current phase",
			nil,
		)

	default:
		slog.Error(
			"tool invocation failed",
			"error", err,
		)

		writeError(
			w,
			http.StatusInternalServerError,
			"invocation_failed",
			"tool invocation failed",
			nil,
		)
	}
}
