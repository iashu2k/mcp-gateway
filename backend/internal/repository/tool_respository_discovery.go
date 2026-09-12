package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/iashu2k/mcp-gateway/backend/internal/domain"
)

// UpsertDiscovered syncs tools found via upstream tools/list discovery
// (Phase 9.4, D4). New tools insert with source='discovered'; existing rows
// (matched on the server_id+name unique constraint) refresh only the
// discovered metadata — risk_level and enabled are admin governance settings
// and are never overwritten by a sync. Runs in one transaction so a failed
// sync never leaves a half-updated catalog.
func (r *ToolRepository) UpsertDiscovered(
	ctx context.Context,
	serverID uuid.UUID,
	tools []domain.MCPTool,
) (int, error) {
	const query = `
    INSERT INTO mcp_tools (
      server_id,
      name,
      title,
      description,
      input_schema,
      risk_level,
      enabled,
      source
    )
    VALUES ($1, $2, $3, $4, $5, $6, $7, 'discovered')
    ON CONFLICT (server_id, name) DO UPDATE SET
      title = EXCLUDED.title,
      description = EXCLUDED.description,
      input_schema = EXCLUDED.input_schema,
      source = 'discovered',
      updated_at = NOW()
  `

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("begin discovery upsert: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	synced := 0
	for _, tool := range tools {
		if _, err := tx.Exec(
			ctx,
			query,
			serverID,
			tool.Name,
			tool.Title,
			tool.Description,
			tool.InputSchema,
			tool.RiskLevel,
			tool.Enabled,
		); err != nil {
			return 0, fmt.Errorf("upsert discovered tool %q: %w", tool.Name, err)
		}
		synced++
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("commit discovery upsert: %w", err)
	}

	return synced, nil
}
