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
