package service

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/iashu2k/mcp-gateway/backend/internal/domain"
)

var ErrInvalidInvocationID = errors.New("invocation ID must be a valid UUID")

type InvocationHistoryStore interface {
	List(
		ctx context.Context,
		serverID *uuid.UUID,
		toolID *uuid.UUID,
		userID *uuid.UUID,
		status string,
		limit int,
		offset int,
	) ([]domain.ToolInvocation, error)

	GetByID(
		ctx context.Context,
		id uuid.UUID,
	) (domain.ToolInvocation, error)
}

type InvocationHistoryService struct {
	store InvocationHistoryStore
}

func NewInvocationHistoryService(
	store InvocationHistoryStore,
) *InvocationHistoryService {
	return &InvocationHistoryService{store: store}
}

func (s *InvocationHistoryService) List(
	ctx context.Context,
	user domain.AuthenticatedUser,
	serverID string,
	toolID string,
	status string,
	limit int,
	offset int,
) ([]domain.ToolInvocation, error) {
	var serverUUID *uuid.UUID
	if serverID != "" {
		parsed, err := uuid.Parse(serverID)
		if err != nil {
			return nil, ValidationError{
				Fields: []FieldError{{Field: "serverId", Message: "must be a valid UUID"}},
			}
		}
		serverUUID = &parsed
	}

	var toolUUID *uuid.UUID
	if toolID != "" {
		parsed, err := uuid.Parse(toolID)
		if err != nil {
			return nil, ValidationError{
				Fields: []FieldError{{Field: "toolId", Message: "must be a valid UUID"}},
			}
		}
		toolUUID = &parsed
	}

	// Non-admin users can only see their own invocations
	var userUUID *uuid.UUID
	if user.Role != domain.RoleAdmin {
		userUUID = &user.ID
	}

	// Validate status if provided
	if status != "" {
		validStatuses := map[string]bool{
			domain.InvocationStatusRunning:   true,
			domain.InvocationStatusSucceeded: true,
			domain.InvocationStatusFailed:    true,
			domain.InvocationStatusDenied:    true,
		}
		if !validStatuses[status] {
			return nil, ValidationError{
				Fields: []FieldError{{Field: "status", Message: "must be running, succeeded, failed, or denied"}},
			}
		}
	}

	// Enforce pagination limits
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	return s.store.List(ctx, serverUUID, toolUUID, userUUID, status, limit, offset)
}

func (s *InvocationHistoryService) GetByID(
	ctx context.Context,
	user domain.AuthenticatedUser,
	invocationID string,
) (domain.ToolInvocation, error) {
	id, err := uuid.Parse(invocationID)
	if err != nil {
		return domain.ToolInvocation{}, ErrInvalidInvocationID
	}

	inv, err := s.store.GetByID(ctx, id)
	if err != nil {
		return domain.ToolInvocation{}, err
	}

	// Non-admin users can only see their own invocations
	if user.Role != domain.RoleAdmin && inv.UserID != user.ID {
		return domain.ToolInvocation{}, ErrInvocationForbidden
	}

	return inv, nil
}