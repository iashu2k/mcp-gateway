package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"regexp"
	"time"

	"net/url"
	"strings"

	"github.com/google/uuid"
	"github.com/iashu2k/mcp-gateway/backend/internal/domain"
)

var (
	ErrDuplicateName   = errors.New("an MCP server with this name already exists")
	ErrInvalidServerID = errors.New("invalid MCP server ID")
)

type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type ValidationError struct {
	Fields []FieldError `json:"fields"`
}

func (e ValidationError) Error() string {
	return "request validation failed"
}

type ServerStore interface {
	Create(ctx context.Context, server domain.MCPServer) (domain.MCPServer, error)
	List(ctx context.Context) ([]domain.MCPServer, error)
	GetByID(ctx context.Context, id uuid.UUID) (domain.MCPServer, error)
	Update(ctx context.Context, server domain.MCPServer) (domain.MCPServer, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

// ToolDiscoverer runs tools/list against a live upstream. Satisfied by
// executor.MCPExecutor (Phase 9.4, D4).
type ToolDiscoverer interface {
	DiscoverTools(ctx context.Context, server domain.MCPServer) ([]domain.MCPTool, error)
}

// DiscoveredToolStore persists discovered tools. Satisfied by
// repository.ToolRepository.
type DiscoveredToolStore interface {
	UpsertDiscovered(ctx context.Context, serverID uuid.UUID, tools []domain.MCPTool) (int, error)
}

// DiscoveryDeps wires optional registration-time tool discovery. Passed
// variadically to NewServerService so existing callers/tests that
// construct the service without discovery keep compiling.
type DiscoveryDeps struct {
	Discoverer ToolDiscoverer
	Tools      DiscoveredToolStore
}

type ServerService struct {
	repository ServerStore
	discovery  *DiscoveryDeps
}

func NewServerService(
	repository ServerStore,
	discovery ...DiscoveryDeps,
) *ServerService {
	service := &ServerService{repository: repository}
	if len(discovery) > 0 && discovery[0].Discoverer != nil && discovery[0].Tools != nil {
		deps := discovery[0]
		service.discovery = &deps
	}
	return service
}

func (s *ServerService) Create(
	ctx context.Context,
	request domain.CreateMCPServerRequest,
) (domain.MCPServer, error) {
	request.Name = strings.TrimSpace(request.Name)
	request.Description = strings.TrimSpace(request.Description)
	request.BaseURL = strings.TrimSpace(request.BaseURL)
	request.TransportType = strings.TrimSpace(request.TransportType)
	request.OwnerTeam = strings.TrimSpace(request.OwnerTeam)

	if request.TransportType == "" {
		request.TransportType = domain.TransportStreamableHTTP
	}

	connectionConfig := json.RawMessage(`{}`)
	if request.ConnectionConfig != nil {
		connectionConfig = json.RawMessage(
			strings.TrimSpace(string(*request.ConnectionConfig)),
		)
	}

	if err := validateCreateRequest(request); err != nil {
		return domain.MCPServer{}, err
	}

	server := domain.MCPServer{
		Name:             request.Name,
		Description:      request.Description,
		BaseURL:          request.BaseURL,
		TransportType:    request.TransportType,
		Status:           domain.ServerStatusActive,
		OwnerTeam:        request.OwnerTeam,
		ConnectionConfig: connectionConfig,
	}

	if err := validateConnectionConfigField(server.ConnectionConfig); err != nil {
		return domain.MCPServer{}, err
	}

	created, err := s.repository.Create(ctx, server)
	if err != nil {
		if isUniqueViolation(err) {
			return domain.MCPServer{}, ErrDuplicateName
		}
		return domain.MCPServer{}, err
	}

	s.syncDiscoveredTools(created)

	return created, nil
}

func (s *ServerService) List(ctx context.Context) ([]domain.MCPServer, error) {
	return s.repository.List(ctx)
}

func (s *ServerService) GetByID(
	ctx context.Context,
	serverID string,
) (domain.MCPServer, error) {
	id, err := parseServerID(serverID)
	if err != nil {
		return domain.MCPServer{}, err
	}

	return s.repository.GetByID(ctx, id)
}

func (s *ServerService) Update(
	ctx context.Context,
	serverID string,
	request domain.UpdateMCPServerRequest,
) (domain.MCPServer, error) {
	id, err := parseServerID(serverID)
	if err != nil {
		return domain.MCPServer{}, err
	}

	existing, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return domain.MCPServer{}, err
	}

	if request.Name != nil {
		existing.Name = strings.TrimSpace(*request.Name)
	}
	if request.Description != nil {
		existing.Description = strings.TrimSpace(*request.Description)
	}
	if request.BaseURL != nil {
		existing.BaseURL = strings.TrimSpace(*request.BaseURL)
	}
	if request.TransportType != nil {
		existing.TransportType = strings.TrimSpace(*request.TransportType)
	}
	if request.Status != nil {
		existing.Status = strings.TrimSpace(*request.Status)
	}
	if request.OwnerTeam != nil {
		existing.OwnerTeam = strings.TrimSpace(*request.OwnerTeam)
	}
	if request.ConnectionConfig != nil {
		existing.ConnectionConfig = json.RawMessage(
			strings.TrimSpace(string(*request.ConnectionConfig)),
		)
	}

	if err := validateServer(existing); err != nil {
		return domain.MCPServer{}, err
	}

	if err := validateConnectionConfigField(existing.ConnectionConfig); err != nil {
		return domain.MCPServer{}, err
	}

	updated, err := s.repository.Update(ctx, existing)
	if err != nil {
		if isUniqueViolation(err) {
			return domain.MCPServer{}, ErrDuplicateName
		}
		return domain.MCPServer{}, err
	}

	s.syncDiscoveredTools(updated)

	return updated, nil
}

func (s *ServerService) Delete(
	ctx context.Context,
	serverID string,
) error {
	id, err := parseServerID(serverID)
	if err != nil {
		return err
	}

	return s.repository.Delete(ctx, id)
}

// syncDiscoveredTools runs registration-time tools/list discovery for
// streamable_http servers (D4). Failure keeps the server row and only logs
// the error — discovery is best-effort, and an unreachable upstream must not
// block catalog management (D4 failure mode). The sync is detached from the
// request context and bounded, so it completes even if the client
// disconnects, without hanging the HTTP response indefinitely.
func (s *ServerService) syncDiscoveredTools(server domain.MCPServer) {
	if s.discovery == nil ||
		server.TransportType != domain.TransportStreamableHTTP {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	tools, err := s.discovery.Discoverer.DiscoverTools(ctx, server)
	if err != nil {
		slog.Error("tool discovery failed; server saved without discovered tools",
			"server_id", server.ID,
			"server_name", server.Name,
			"error", err,
		)
		return
	}

	synced, err := s.discovery.Tools.UpsertDiscovered(ctx, server.ID, tools)
	if err != nil {
		slog.Error("failed to persist discovered tools",
			"server_id", server.ID,
			"server_name", server.Name,
			"error", err,
		)
		return
	}

	slog.Info("discovered tools synced",
		"server_id", server.ID,
		"server_name", server.Name,
		"tools", synced,
	)
}

func parseServerID(serverID string) (uuid.UUID, error) {
	id, err := uuid.Parse(serverID)
	if err != nil {
		return uuid.Nil, ErrInvalidServerID
	}

	return id, nil
}

func validateCreateRequest(request domain.CreateMCPServerRequest) error {
	server := domain.MCPServer{
		Name:          request.Name,
		Description:   request.Description,
		BaseURL:       request.BaseURL,
		TransportType: request.TransportType,
		Status:        domain.ServerStatusActive,
		OwnerTeam:     request.OwnerTeam,
	}

	return validateServer(server)
}

func validateServer(server domain.MCPServer) error {
	fieldErrors := make([]FieldError, 0)

	if server.Name == "" {
		fieldErrors = append(fieldErrors, FieldError{
			Field:   "name",
			Message: "name is required",
		})
	}

	if len(server.Name) > 100 {
		fieldErrors = append(fieldErrors, FieldError{
			Field:   "name",
			Message: "name must be 100 characters or fewer",
		})
	}

	if server.Description == "" {
		fieldErrors = append(fieldErrors, FieldError{
			Field:   "description",
			Message: "description is required",
		})
	}

	if len(server.Description) > 1000 {
		fieldErrors = append(fieldErrors, FieldError{
			Field:   "description",
			Message: "description must be 1000 characters or fewer",
		})
	}

	if err := validateBaseURL(server.BaseURL, server.TransportType); err != nil {
		fieldErrors = append(fieldErrors, FieldError{
			Field:   "baseUrl",
			Message: err.Error(),
		})
	}

	if !isValidTransport(server.TransportType) {
		fieldErrors = append(fieldErrors, FieldError{
			Field:   "transportType",
			Message: "must be one of: streamable_http, sse, stdio",
		})
	}

	if !isValidStatus(server.Status) {
		fieldErrors = append(fieldErrors, FieldError{
			Field:   "status",
			Message: "must be one of: active, inactive, unhealthy",
		})
	}

	if server.OwnerTeam == "" {
		fieldErrors = append(fieldErrors, FieldError{
			Field:   "ownerTeam",
			Message: "ownerTeam is required",
		})
	}

	if len(server.OwnerTeam) > 100 {
		fieldErrors = append(fieldErrors, FieldError{
			Field:   "ownerTeam",
			Message: "ownerTeam must be 100 characters or fewer",
		})
	}

	if len(fieldErrors) > 0 {
		return ValidationError{Fields: fieldErrors}
	}

	return nil
}

// envVarNamePattern enforces the Phase 9 credential-reference rule (D5
// security note): connection_config header values name environment
// variables, they never hold raw secrets.
var envVarNamePattern = regexp.MustCompile(`^[A-Z_][A-Z0-9_]*$`)

// validateConnectionConfigField wraps connection-config validation in the
// service's standard ValidationError shape.
func validateConnectionConfigField(raw json.RawMessage) error {
	if err := validateConnectionConfig(raw); err != nil {
		return ValidationError{
			Fields: []FieldError{{Field: "connectionConfig", Message: err.Error()}},
		}
	}
	return nil
}

func validateConnectionConfig(raw json.RawMessage) error {
	if len(raw) == 0 {
		return errors.New("connectionConfig is required")
	}

	var decoded any
	if err := json.Unmarshal(raw, &decoded); err != nil {
		return errors.New("must be valid JSON")
	}

	root, ok := decoded.(map[string]any)
	if !ok {
		return errors.New("must be a JSON object")
	}

	headers, exists := root["headers"]
	if !exists {
		return nil
	}

	headerMap, ok := headers.(map[string]any)
	if !ok {
		return errors.New(`"headers" must be a JSON object`)
	}

	for name, value := range headerMap {
		reference, ok := value.(string)
		if !ok {
			return fmt.Errorf("header %q must name an environment variable", name)
		}
		if !envVarNamePattern.MatchString(reference) {
			return fmt.Errorf(
				"header %q must reference an environment variable name (e.g. GITHUB_TOKEN), not a raw secret",
				name,
			)
		}
	}

	return nil
}

func validateBaseURL(value, transportType string) error {
	if transportType == domain.TransportStdio && value == "" {
		return nil
	}

	if value == "" {
		return errors.New("baseUrl is required")
	}

	parsedURL, err := url.ParseRequestURI(value)
	if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
		return errors.New("must be a valid absolute HTTP URL")
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return errors.New("must use http or https")
	}

	return nil
}

func isValidTransport(value string) bool {
	switch value {
	case domain.TransportStreamableHTTP, domain.TransportSSE, domain.TransportStdio:
		return true
	default:
		return false
	}
}

func isValidStatus(value string) bool {
	switch value {
	case domain.ServerStatusActive,
		domain.ServerStatusInactive,
		domain.ServerStatusUnhealthy:
		return true
	default:
		return false
	}
}
