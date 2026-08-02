package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/iashu2k/mcp-gateway/backend/internal/domain"
	"github.com/iashu2k/mcp-gateway/backend/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

var ErrInvalidCredentials = errors.New("invalid email or password")

type UserReader interface {
	GetByEmail(ctx context.Context, email string) (domain.User, error)
}

type JWTIssuer interface {
	Generate(user domain.User) (string, time.Time, error)
}

type AuthService struct {
	users  UserReader
	tokens JWTIssuer
}

func NewAuthService(users UserReader, tokens JWTIssuer) *AuthService {
	return &AuthService{
		users:  users,
		tokens: tokens,
	}
}

func (s *AuthService) Login(
	ctx context.Context,
	request domain.LoginRequest,
) (domain.LoginResponse, error) {
	email := strings.ToLower(strings.TrimSpace(request.Email))
	password := request.Password

	if email == "" || password == "" {
		return domain.LoginResponse{}, ErrInvalidCredentials
	}

	user, err := s.users.GetByEmail(ctx, email)
	if errors.Is(err, repository.ErrUserNotFound) {
		return domain.LoginResponse{}, ErrInvalidCredentials
	}
	if err != nil {
		return domain.LoginResponse{}, err
	}

	if !user.Active {
		return domain.LoginResponse{}, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword(
		[]byte(user.PasswordHash),
		[]byte(password),
	); err != nil {
		return domain.LoginResponse{}, ErrInvalidCredentials
	}

	accessToken, expiresAt, err := s.tokens.Generate(user)
	if err != nil {
		return domain.LoginResponse{}, err
	}

	return domain.LoginResponse{
		AccessToken: accessToken,
		TokenType:   "Bearer",
		ExpiresAt:   expiresAt,
		User:        user.ToAuthenticatedUser(),
	}, nil
}
