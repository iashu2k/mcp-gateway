package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/iashu2k/mcp-gateway/backend/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrUserNotFound = errors.New("user not found")

type UserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) GetByEmail(
	ctx context.Context,
	email string,
) (domain.User, error) {
	const query = `
		SELECT
			id,
			email,
			password_hash,
			display_name,
			role,
			active,
			created_at,
			updated_at
		FROM users
		WHERE email = $1
	`

	user, err := scanUser(r.db.QueryRow(ctx, query, email))
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.User{}, ErrUserNotFound
	}
	if err != nil {
		return domain.User{}, fmt.Errorf("get user by email: %w", err)
	}

	return user, nil
}

func scanUser(row rowScanner) (domain.User, error) {
	var user domain.User

	err := row.Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.DisplayName,
		&user.Role,
		&user.Active,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	return user, err
}
