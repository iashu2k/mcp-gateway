package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/iashu2k/mcp-gateway/backend/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrInvocationNotFound = errors.New("tool invocation not found")

type InvocationRepository struct {
	db *pgxpool.Pool
}

func NewInvocationRepository(db *pgxpool.Pool) *InvocationRepository {
	return &InvocationRepository{db: db}
}

func (r *InvocationRepository) CreateRunning(
	ctx context.Context,
	invocation domain.ToolInvocation,
) (domain.ToolInvocation, error) {
	const query = `
		INSERT INTO tool_invocations (
			server_id,
			tool_id,
			user_id,
			status,
			request_arguments
		)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING
			id,
			server_id,
			tool_id,
			user_id,
			status,
			request_arguments,
			response_payload,
			error_code,
			error_message,
			duration_ms,
			created_at,
			completed_at
	`

	created, err := scanInvocation(r.db.QueryRow(
		ctx,
		query,
		invocation.ServerID,
		invocation.ToolID,
		invocation.UserID,
		domain.InvocationStatusRunning,
		invocation.RequestArguments,
	))
	if err != nil {
		return domain.ToolInvocation{}, fmt.Errorf("create invocation: %w", err)
	}

	return created, nil
}

func (r *InvocationRepository) MarkSucceeded(
	ctx context.Context,
	invocationID uuid.UUID,
	responsePayload json.RawMessage,
	durationMS int64,
	completedAt time.Time,
) (domain.ToolInvocation, error) {
	const query = `
		UPDATE tool_invocations
		SET
			status = $2,
			response_payload = $3,
			duration_ms = $4,
			completed_at = $5
		WHERE id = $1
		RETURNING
			id,
			server_id,
			tool_id,
			user_id,
			status,
			request_arguments,
			response_payload,
			error_code,
			error_message,
			duration_ms,
			created_at,
			completed_at
	`

	invocation, err := scanInvocation(r.db.QueryRow(
		ctx,
		query,
		invocationID,
		domain.InvocationStatusSucceeded,
		responsePayload,
		durationMS,
		completedAt,
	))
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ToolInvocation{}, ErrInvocationNotFound
	}
	if err != nil {
		return domain.ToolInvocation{}, fmt.Errorf(
			"mark invocation succeeded: %w",
			err,
		)
	}

	return invocation, nil
}

func (r *InvocationRepository) MarkFailed(
	ctx context.Context,
	invocationID uuid.UUID,
	errorCode string,
	errorMessage string,
	durationMS int64,
	completedAt time.Time,
) (domain.ToolInvocation, error) {
	const query = `
		UPDATE tool_invocations
		SET
			status = $2,
			error_code = $3,
			error_message = $4,
			duration_ms = $5,
			completed_at = $6
		WHERE id = $1
		RETURNING
			id,
			server_id,
			tool_id,
			user_id,
			status,
			request_arguments,
			response_payload,
			error_code,
			error_message,
			duration_ms,
			created_at,
			completed_at
	`

	invocation, err := scanInvocation(r.db.QueryRow(
		ctx,
		query,
		invocationID,
		domain.InvocationStatusFailed,
		errorCode,
		errorMessage,
		durationMS,
		completedAt,
	))
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ToolInvocation{}, ErrInvocationNotFound
	}
	if err != nil {
		return domain.ToolInvocation{}, fmt.Errorf(
			"mark invocation failed: %w",
			err,
		)
	}

	return invocation, nil
}

func scanInvocation(row rowScanner) (domain.ToolInvocation, error) {
	var invocation domain.ToolInvocation

	err := row.Scan(
		&invocation.ID,
		&invocation.ServerID,
		&invocation.ToolID,
		&invocation.UserID,
		&invocation.Status,
		&invocation.RequestArguments,
		&invocation.ResponsePayload,
		&invocation.ErrorCode,
		&invocation.ErrorMessage,
		&invocation.DurationMS,
		&invocation.CreatedAt,
		&invocation.CompletedAt,
	)

	return invocation, err
}

// List returns invocation history with optional filtering.
func (r *InvocationRepository) List(
	ctx context.Context,
	serverID *uuid.UUID,
	toolID *uuid.UUID,
	userID *uuid.UUID,
	status string,
	limit int,
	offset int,
) ([]domain.ToolInvocation, error) {
	query := `
		SELECT
			id,
			server_id,
			tool_id,
			user_id,
			status,
			request_arguments,
			response_payload,
			error_code,
			error_message,
			duration_ms,
			created_at,
			completed_at
		FROM tool_invocations
		WHERE 1=1
	`

	args := []any{}
	argPos := 1

	if serverID != nil {
		query += fmt.Sprintf(" AND server_id = $%d", argPos)
		args = append(args, *serverID)
		argPos++
	}

	if toolID != nil {
		query += fmt.Sprintf(" AND tool_id = $%d", argPos)
		args = append(args, *toolID)
		argPos++
	}

	if userID != nil {
		query += fmt.Sprintf(" AND user_id = $%d", argPos)
		args = append(args, *userID)
		argPos++
	}

	if status != "" {
		query += fmt.Sprintf(" AND status = $%d", argPos)
		args = append(args, status)
		argPos++
	}

	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argPos, argPos+1)
	args = append(args, limit, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list invocations: %w", err)
	}
	defer rows.Close()

	invocations := []domain.ToolInvocation{}
	for rows.Next() {
		var inv domain.ToolInvocation
		err := rows.Scan(
			&inv.ID,
			&inv.ServerID,
			&inv.ToolID,
			&inv.UserID,
			&inv.Status,
			&inv.RequestArguments,
			&inv.ResponsePayload,
			&inv.ErrorCode,
			&inv.ErrorMessage,
			&inv.DurationMS,
			&inv.CreatedAt,
			&inv.CompletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan invocation: %w", err)
		}
		invocations = append(invocations, inv)
	}

	return invocations, rows.Err()
}

// GetByID returns a single invocation by ID.
func (r *InvocationRepository) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (domain.ToolInvocation, error) {
	const query = `
		SELECT
			id,
			server_id,
			tool_id,
			user_id,
			status,
			request_arguments,
			response_payload,
			error_code,
			error_message,
			duration_ms,
			created_at,
			completed_at
		FROM tool_invocations
		WHERE id = $1
	`

	inv, err := scanInvocation(r.db.QueryRow(ctx, query, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ToolInvocation{}, ErrInvocationNotFound
	}
	if err != nil {
		return domain.ToolInvocation{}, fmt.Errorf("get invocation: %w", err)
	}

	return inv, nil
}
