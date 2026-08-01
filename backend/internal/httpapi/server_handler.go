package httpapi

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/iashu2k/mcp-gateway/backend/internal/domain"
	"github.com/iashu2k/mcp-gateway/backend/internal/repository"
	"github.com/iashu2k/mcp-gateway/backend/internal/service"
)

type ServerHandler struct {
	service *service.ServerService
}

func NewServerHandler(service *service.ServerService) *ServerHandler {
	return &ServerHandler{service: service}
}

func (h *ServerHandler) Create(w http.ResponseWriter, r *http.Request) {
	var request domain.CreateMCPServerRequest

	if ok := decodeJSONBody(w, r, &request); !ok {
		return
	}

	server, err := h.service.Create(r.Context(), request)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, server)
}

func (h *ServerHandler) List(w http.ResponseWriter, r *http.Request) {
	servers, err := h.service.List(r.Context())
	if err != nil {
		writeError(
			w,
			http.StatusInternalServerError,
			"internal_error",
			"could not list MCP servers",
			nil,
		)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"data":  servers,
		"count": len(servers),
	})
}

func (h *ServerHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	serverID := chi.URLParam(r, "serverID")

	server, err := h.service.GetByID(r.Context(), serverID)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, server)
}

func (h *ServerHandler) Update(w http.ResponseWriter, r *http.Request) {
	serverID := chi.URLParam(r, "serverID")

	var request domain.UpdateMCPServerRequest
	if ok := decodeJSONBody(w, r, &request); !ok {
		return
	}

	server, err := h.service.Update(r.Context(), serverID, request)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, server)
}

func (h *ServerHandler) Delete(w http.ResponseWriter, r *http.Request) {
	serverID := chi.URLParam(r, "serverID")

	if err := h.service.Delete(r.Context(), serverID); err != nil {
		h.handleServiceError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ServerHandler) handleServiceError(w http.ResponseWriter, err error) {
	var validationError service.ValidationError

	switch {
	case errors.As(err, &validationError):
		writeError(
			w,
			http.StatusBadRequest,
			"validation_error",
			validationError.Error(),
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

	case errors.Is(err, service.ErrDuplicateName):
		writeError(
			w,
			http.StatusConflict,
			"duplicate_server_name",
			"an MCP server with this name already exists",
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

	default:
		writeError(
			w,
			http.StatusInternalServerError,
			"internal_error",
			"an unexpected error occurred",
			nil,
		)
	}
}
