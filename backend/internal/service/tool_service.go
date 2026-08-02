package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/iashu2k/mcp-gateway/backend/internal/domain"
	"github.com/jackc/pgx/v5/pgconn"
)

var (
	ErrInvalidToolID = errors.New("invalid MCP tool ID")
	ErrDuplicateTool = errors.New("an MCP tool with this name already exists for this server")
)

type ToolStore interface {
	Create(ctx context.Context, tool domain.MCPTool) (domain.MCPTool, error)
	ListByServerID(ctx context.Context, serverID uuid.UUID) ([]domain.MCPTool, error)
	GetByID(ctx context.Context, serverID uuid.UUID, toolID uuid.UUID) (domain.MCPTool, error)
	Update(ctx context.Context, tool domain.MCPTool) (domain.MCPTool, error)
	Delete(ctx context.Context, serverID uuid.UUID, toolID uuid.UUID) error
}

type ServerReader interface {
	GetByID(ctx context.Context, id uuid.UUID) (domain.MCPServer, error)
}

type ToolService struct {
	toolRepository   ToolStore
	serverRepository ServerReader
}

func NewToolService(
	toolRepository ToolStore,
	serverRepository ServerReader,
) *ToolService {
	return &ToolService{
		toolRepository:   toolRepository,
		serverRepository: serverRepository,
	}
}

func (s *ToolService) Create(
	ctx context.Context,
	serverID string,
	request domain.CreateMCPToolRequest,
) (domain.MCPTool, error) {
	id, err := parseServerID(serverID)
	if err != nil {
		return domain.MCPTool{}, err
	}

	if _, err := s.serverRepository.GetByID(ctx, id); err != nil {
		return domain.MCPTool{}, err
	}

	request.Name = strings.TrimSpace(request.Name)
	request.Title = strings.TrimSpace(request.Title)
	request.Description = strings.TrimSpace(request.Description)
	request.RiskLevel = strings.TrimSpace(request.RiskLevel)

	if request.RiskLevel == "" {
		request.RiskLevel = domain.ToolRiskLow
	}

	enabled := true
	if request.Enabled != nil {
		enabled = *request.Enabled
	}

	tool := domain.MCPTool{
		ServerID:    id,
		Name:        request.Name,
		Title:       request.Title,
		Description: request.Description,
		InputSchema: request.InputSchema,
		RiskLevel:   request.RiskLevel,
		Enabled:     enabled,
	}

	if err := validateTool(tool); err != nil {
		return domain.MCPTool{}, err
	}

	created, err := s.toolRepository.Create(ctx, tool)
	if err != nil {
		if isUniqueViolation(err) {
			return domain.MCPTool{}, ErrDuplicateTool
		}
		return domain.MCPTool{}, err
	}

	return created, nil
}

func (s *ToolService) ListByServerID(
	ctx context.Context,
	serverID string,
) ([]domain.MCPTool, error) {
	id, err := parseServerID(serverID)
	if err != nil {
		return nil, err
	}

	if _, err := s.serverRepository.GetByID(ctx, id); err != nil {
		return nil, err
	}

	return s.toolRepository.ListByServerID(ctx, id)
}

func (s *ToolService) GetByID(
	ctx context.Context,
	serverID string,
	toolID string,
) (domain.MCPTool, error) {
	serverUUID, toolUUID, err := parseIDs(serverID, toolID)
	if err != nil {
		return domain.MCPTool{}, err
	}

	if _, err := s.serverRepository.GetByID(ctx, serverUUID); err != nil {
		return domain.MCPTool{}, err
	}

	return s.toolRepository.GetByID(ctx, serverUUID, toolUUID)
}

func (s *ToolService) Update(
	ctx context.Context,
	serverID string,
	toolID string,
	request domain.UpdateMCPToolRequest,
) (domain.MCPTool, error) {
	serverUUID, toolUUID, err := parseIDs(serverID, toolID)
	if err != nil {
		return domain.MCPTool{}, err
	}

	if _, err := s.serverRepository.GetByID(ctx, serverUUID); err != nil {
		return domain.MCPTool{}, err
	}

	existing, err := s.toolRepository.GetByID(ctx, serverUUID, toolUUID)
	if err != nil {
		return domain.MCPTool{}, err
	}

	if request.Name != nil {
		existing.Name = strings.TrimSpace(*request.Name)
	}
	if request.Title != nil {
		existing.Title = strings.TrimSpace(*request.Title)
	}
	if request.Description != nil {
		existing.Description = strings.TrimSpace(*request.Description)
	}
	if request.InputSchema != nil {
		existing.InputSchema = *request.InputSchema
	}
	if request.RiskLevel != nil {
		existing.RiskLevel = strings.TrimSpace(*request.RiskLevel)
	}
	if request.Enabled != nil {
		existing.Enabled = *request.Enabled
	}

	if err := validateTool(existing); err != nil {
		return domain.MCPTool{}, err
	}

	updated, err := s.toolRepository.Update(ctx, existing)
	if err != nil {
		if isUniqueViolation(err) {
			return domain.MCPTool{}, ErrDuplicateTool
		}
		return domain.MCPTool{}, err
	}

	return updated, nil
}

func (s *ToolService) Delete(
	ctx context.Context,
	serverID string,
	toolID string,
) error {
	serverUUID, toolUUID, err := parseIDs(serverID, toolID)
	if err != nil {
		return err
	}

	if _, err := s.serverRepository.GetByID(ctx, serverUUID); err != nil {
		return err
	}

	return s.toolRepository.Delete(ctx, serverUUID, toolUUID)
}

func parseIDs(serverID, toolID string) (uuid.UUID, uuid.UUID, error) {
	serverUUID, err := parseServerID(serverID)
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}

	toolUUID, err := uuid.Parse(toolID)
	if err != nil {
		return uuid.Nil, uuid.Nil, ErrInvalidToolID
	}

	return serverUUID, toolUUID, nil
}

func validateTool(tool domain.MCPTool) error {
	fieldErrors := make([]FieldError, 0)

	if tool.Name == "" {
		fieldErrors = append(fieldErrors, FieldError{
			Field:   "name",
			Message: "name is required",
		})
	}

	if len(tool.Name) > 100 {
		fieldErrors = append(fieldErrors, FieldError{
			Field:   "name",
			Message: "name must be 100 characters or fewer",
		})
	}

	if tool.Title != "" && len(tool.Title) > 150 {
		fieldErrors = append(fieldErrors, FieldError{
			Field:   "title",
			Message: "title must be 150 characters or fewer",
		})
	}

	if tool.Description == "" {
		fieldErrors = append(fieldErrors, FieldError{
			Field:   "description",
			Message: "description is required",
		})
	}

	if len(tool.Description) > 2000 {
		fieldErrors = append(fieldErrors, FieldError{
			Field:   "description",
			Message: "description must be 2000 characters or fewer",
		})
	}

	if !isValidRiskLevel(tool.RiskLevel) {
		fieldErrors = append(fieldErrors, FieldError{
			Field:   "riskLevel",
			Message: "must be one of: low, medium, high",
		})
	}

	if err := validateInputSchema(tool.InputSchema); err != nil {
		fieldErrors = append(fieldErrors, FieldError{
			Field:   "inputSchema",
			Message: err.Error(),
		})
	}

	if len(fieldErrors) > 0 {
		return ValidationError{Fields: fieldErrors}
	}

	return nil
}

func validateInputSchema(schema json.RawMessage) error {
	if len(schema) == 0 {
		return errors.New("inputSchema is required")
	}

	var decoded any
	if err := json.Unmarshal(schema, &decoded); err != nil {
		return errors.New("must be valid JSON")
	}

	root, ok := decoded.(map[string]any)
	if !ok {
		return errors.New("must be a JSON object")
	}

	schemaType, ok := root["type"].(string)
	if !ok || schemaType != "object" {
		return errors.New(`must declare "type": "object"`)
	}

	if properties, exists := root["properties"]; exists {
		if _, ok := properties.(map[string]any); !ok {
			return errors.New(`"properties" must be a JSON object`)
		}
	}

	if required, exists := root["required"]; exists {
		requiredValues, ok := required.([]any)
		if !ok {
			return errors.New(`"required" must be a JSON array`)
		}

		for _, value := range requiredValues {
			if _, ok := value.(string); !ok {
				return errors.New(`"required" must contain only strings`)
			}
		}
	}

	return nil
}

func isValidRiskLevel(value string) bool {
	switch value {
	case domain.ToolRiskLow, domain.ToolRiskMedium, domain.ToolRiskHigh:
		return true
	default:
		return false
	}
}

func isUniqueViolation(err error) bool {
	var pgError *pgconn.PgError
	if errors.As(err, &pgError) {
		return pgError.SQLState() == "23505"
	}

	return strings.Contains(
		strings.ToLower(fmt.Sprint(err)),
		"duplicate key",
	)
}
