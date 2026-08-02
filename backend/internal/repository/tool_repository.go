package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/iashu2k/mcp-gateway/backend/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrToolNotFound = errors.New("MCP tool not found")

type ToolRepository struct {
	db *pgxpool.Pool
}

func NewToolRepository(db *pgxpool.Pool) *ToolRepository {
	return &ToolRepository{db: db}
}

func (r *ToolRepository) Create(
	ctx context.Context,
	tool domain.MCPTool,
) (domain.MCPTool, error) {
	const query = `
		INSERT INTO mcp_tools (
			server_id,
			name,
			title,
			description,
			input_schema,
			risk_level,
			enabled
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING
			id,
			server_id,
			name,
			title,
			description,
			input_schema,
			risk_level,
			enabled,
			created_at,
			updated_at
	`

	created, err := scanTool(r.db.QueryRow(
		ctx,
		query,
		tool.ServerID,
		tool.Name,
		tool.Title,
		tool.Description,
		tool.InputSchema,
		tool.RiskLevel,
		tool.Enabled,
	))
	if err != nil {
		return domain.MCPTool{}, fmt.Errorf("insert MCP tool: %w", err)
	}

	return created, nil
}

func (r *ToolRepository) ListByServerID(
	ctx context.Context,
	serverID uuid.UUID,
) ([]domain.MCPTool, error) {
	const query = `
		SELECT
			id,
			server_id,
			name,
			title,
			description,
			input_schema,
			risk_level,
			enabled,
			created_at,
			updated_at
		FROM mcp_tools
		WHERE server_id = $1
		ORDER BY created_at ASC
	`

	rows, err := r.db.Query(ctx, query, serverID)
	if err != nil {
		return nil, fmt.Errorf("list MCP tools: %w", err)
	}
	defer rows.Close()

	tools := make([]domain.MCPTool, 0)

	for rows.Next() {
		tool, err := scanTool(rows)
		if err != nil {
			return nil, fmt.Errorf("scan MCP tool: %w", err)
		}

		tools = append(tools, tool)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate MCP tools: %w", err)
	}

	return tools, nil
}

func (r *ToolRepository) GetByID(
	ctx context.Context,
	serverID uuid.UUID,
	toolID uuid.UUID,
) (domain.MCPTool, error) {
	const query = `
		SELECT
			id,
			server_id,
			name,
			title,
			description,
			input_schema,
			risk_level,
			enabled,
			created_at,
			updated_at
		FROM mcp_tools
		WHERE server_id = $1
		  AND id = $2
	`

	tool, err := scanTool(r.db.QueryRow(ctx, query, serverID, toolID))
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.MCPTool{}, ErrToolNotFound
	}
	if err != nil {
		return domain.MCPTool{}, fmt.Errorf("get MCP tool: %w", err)
	}

	return tool, nil
}

func (r *ToolRepository) Update(
	ctx context.Context,
	tool domain.MCPTool,
) (domain.MCPTool, error) {
	const query = `
		UPDATE mcp_tools
		SET
			name = $3,
			title = $4,
			description = $5,
			input_schema = $6,
			risk_level = $7,
			enabled = $8,
			updated_at = NOW()
		WHERE server_id = $1
		  AND id = $2
		RETURNING
			id,
			server_id,
			name,
			title,
			description,
			input_schema,
			risk_level,
			enabled,
			created_at,
			updated_at
	`

	updated, err := scanTool(r.db.QueryRow(
		ctx,
		query,
		tool.ServerID,
		tool.ID,
		tool.Name,
		tool.Title,
		tool.Description,
		tool.InputSchema,
		tool.RiskLevel,
		tool.Enabled,
	))
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.MCPTool{}, ErrToolNotFound
	}
	if err != nil {
		return domain.MCPTool{}, fmt.Errorf("update MCP tool: %w", err)
	}

	return updated, nil
}

func (r *ToolRepository) Delete(
	ctx context.Context,
	serverID uuid.UUID,
	toolID uuid.UUID,
) error {
	const query = `
		DELETE FROM mcp_tools
		WHERE server_id = $1
		  AND id = $2
	`

	commandTag, err := r.db.Exec(ctx, query, serverID, toolID)
	if err != nil {
		return fmt.Errorf("delete MCP tool: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return ErrToolNotFound
	}

	return nil
}

func scanTool(row rowScanner) (domain.MCPTool, error) {
	var tool domain.MCPTool

	err := row.Scan(
		&tool.ID,
		&tool.ServerID,
		&tool.Name,
		&tool.Title,
		&tool.Description,
		&tool.InputSchema,
		&tool.RiskLevel,
		&tool.Enabled,
		&tool.CreatedAt,
		&tool.UpdatedAt,
	)

	return tool, err
}
