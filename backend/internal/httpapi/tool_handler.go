package httpapi

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/iashu2k/mcp-gateway/backend/internal/domain"
	"github.com/iashu2k/mcp-gateway/backend/internal/repository"
	"github.com/iashu2k/mcp-gateway/backend/internal/service"
)

type ToolHandler struct {
	service *service.ToolService
}

func NewToolHandler(service *service.ToolService) *ToolHandler {
	return &ToolHandler{service: service}
}

func (h *ToolHandler) Create(w http.ResponseWriter, r *http.Request) {
	serverID := chi.URLParam(r, "serverID")

	var request domain.CreateMCPToolRequest
	if ok := decodeJSONBody(w, r, &request); !ok {
		return
	}

	tool, err := h.service.Create(r.Context(), serverID, request)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, tool)
}

func (h *ToolHandler) ListByServerID(w http.ResponseWriter, r *http.Request) {
	serverID := chi.URLParam(r, "serverID")

	tools, err := h.service.ListByServerID(r.Context(), serverID)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"data":  tools,
		"count": len(tools),
	})
}

func (h *ToolHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	serverID := chi.URLParam(r, "serverID")
	toolID := chi.URLParam(r, "toolID")

	tool, err := h.service.GetByID(r.Context(), serverID, toolID)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, tool)
}

func (h *ToolHandler) Update(w http.ResponseWriter, r *http.Request) {
	serverID := chi.URLParam(r, "serverID")
	toolID := chi.URLParam(r, "toolID")

	var request domain.UpdateMCPToolRequest
	if ok := decodeJSONBody(w, r, &request); !ok {
		return
	}

	tool, err := h.service.Update(r.Context(), serverID, toolID, request)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, tool)
}

func (h *ToolHandler) Delete(w http.ResponseWriter, r *http.Request) {
	serverID := chi.URLParam(r, "serverID")
	toolID := chi.URLParam(r, "toolID")

	if err := h.service.Delete(r.Context(), serverID, toolID); err != nil {
		h.handleServiceError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ToolHandler) handleServiceError(w http.ResponseWriter, err error) {
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

	case errors.Is(err, service.ErrInvalidToolID):
		writeError(
			w,
			http.StatusBadRequest,
			"invalid_tool_id",
			"toolID must be a valid UUID",
			nil,
		)

	case errors.Is(err, service.ErrDuplicateTool):
		writeError(
			w,
			http.StatusConflict,
			"duplicate_tool_name",
			"an MCP tool with this name already exists for this server",
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
