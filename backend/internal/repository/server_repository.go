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

var ErrServerNotFound = errors.New("MCP server not found")

type ServerRepository struct {
	db *pgxpool.Pool
}

func NewServerRepository(db *pgxpool.Pool) *ServerRepository {
	return &ServerRepository{db: db}
}

func (r *ServerRepository) Create(
	ctx context.Context,
	server domain.MCPServer,
) (domain.MCPServer, error) {
	const query = `
		INSERT INTO mcp_servers (
			name,
			description,
			base_url,
			transport_type,
			status,
			owner_team
		)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING
			id,
			name,
			description,
			base_url,
			transport_type,
			status,
			owner_team,
			created_at,
			updated_at
	`

	created, err := scanServer(r.db.QueryRow(
		ctx,
		query,
		server.Name,
		server.Description,
		server.BaseURL,
		server.TransportType,
		server.Status,
		server.OwnerTeam,
	))
	if err != nil {
		return domain.MCPServer{}, fmt.Errorf("insert MCP server: %w", err)
	}

	return created, nil
}

func (r *ServerRepository) List(ctx context.Context) ([]domain.MCPServer, error) {
	const query = `
		SELECT
			id,
			name,
			description,
			base_url,
			transport_type,
			status,
			owner_team,
			created_at,
			updated_at
		FROM mcp_servers
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list MCP servers: %w", err)
	}
	defer rows.Close()

	servers := make([]domain.MCPServer, 0)

	for rows.Next() {
		server, err := scanServer(rows)
		if err != nil {
			return nil, fmt.Errorf("scan MCP server: %w", err)
		}

		servers = append(servers, server)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate MCP servers: %w", err)
	}

	return servers, nil
}

func (r *ServerRepository) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (domain.MCPServer, error) {
	const query = `
		SELECT
			id,
			name,
			description,
			base_url,
			transport_type,
			status,
			owner_team,
			created_at,
			updated_at
		FROM mcp_servers
		WHERE id = $1
	`

	server, err := scanServer(r.db.QueryRow(ctx, query, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.MCPServer{}, ErrServerNotFound
	}
	if err != nil {
		return domain.MCPServer{}, fmt.Errorf("get MCP server: %w", err)
	}

	return server, nil
}

func (r *ServerRepository) Update(
	ctx context.Context,
	server domain.MCPServer,
) (domain.MCPServer, error) {
	const query = `
		UPDATE mcp_servers
		SET
			name = $2,
			description = $3,
			base_url = $4,
			transport_type = $5,
			status = $6,
			owner_team = $7,
			updated_at = NOW()
		WHERE id = $1
		RETURNING
			id,
			name,
			description,
			base_url,
			transport_type,
			status,
			owner_team,
			created_at,
			updated_at
	`

	updated, err := scanServer(r.db.QueryRow(
		ctx,
		query,
		server.ID,
		server.Name,
		server.Description,
		server.BaseURL,
		server.TransportType,
		server.Status,
		server.OwnerTeam,
	))
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.MCPServer{}, ErrServerNotFound
	}
	if err != nil {
		return domain.MCPServer{}, fmt.Errorf("update MCP server: %w", err)
	}

	return updated, nil
}

func (r *ServerRepository) Delete(
	ctx context.Context,
	id uuid.UUID,
) error {
	const query = `DELETE FROM mcp_servers WHERE id = $1`

	commandTag, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete MCP server: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return ErrServerNotFound
	}

	return nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanServer(row rowScanner) (domain.MCPServer, error) {
	var server domain.MCPServer

	err := row.Scan(
		&server.ID,
		&server.Name,
		&server.Description,
		&server.BaseURL,
		&server.TransportType,
		&server.Status,
		&server.OwnerTeam,
		&server.CreatedAt,
		&server.UpdatedAt,
	)

	return server, err
}
