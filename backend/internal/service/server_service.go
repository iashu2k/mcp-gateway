package service

import (
	"context"
	"errors"

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

type ServerService struct {
	repository ServerStore
}

func NewServerService(repository ServerStore) *ServerService {
	return &ServerService{repository: repository}
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

	if err := validateCreateRequest(request); err != nil {
		return domain.MCPServer{}, err
	}

	server := domain.MCPServer{
		Name:          request.Name,
		Description:   request.Description,
		BaseURL:       request.BaseURL,
		TransportType: request.TransportType,
		Status:        domain.ServerStatusActive,
		OwnerTeam:     request.OwnerTeam,
	}

	created, err := s.repository.Create(ctx, server)
	if err != nil {
		if isUniqueViolation(err) {
			return domain.MCPServer{}, ErrDuplicateName
		}
		return domain.MCPServer{}, err
	}

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

	if err := validateServer(existing); err != nil {
		return domain.MCPServer{}, err
	}

	updated, err := s.repository.Update(ctx, existing)
	if err != nil {
		if isUniqueViolation(err) {
			return domain.MCPServer{}, ErrDuplicateName
		}
		return domain.MCPServer{}, err
	}

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
