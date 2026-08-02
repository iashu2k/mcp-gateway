package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/iashu2k/mcp-gateway/backend/internal/domain"
)

func TestTokenServiceGenerateAndParse(t *testing.T) {
	service := NewTokenService(
		"test-secret-that-is-at-least-thirty-two-characters",
		"mcp-gateway-test",
		time.Hour,
	)

	user := domain.User{
		ID:          uuid.New(),
		Email:       "admin@example.com",
		DisplayName: "Test Admin",
		Role:        domain.RoleAdmin,
	}

	token, _, err := service.Generate(user)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	authenticatedUser, err := service.Parse(token)
	if err != nil {
		t.Fatalf("parse token: %v", err)
	}

	if authenticatedUser.ID != user.ID {
		t.Fatalf("expected user ID %q, got %q", user.ID, authenticatedUser.ID)
	}

	if authenticatedUser.Role != domain.RoleAdmin {
		t.Fatalf(
			"expected role %q, got %q",
			domain.RoleAdmin,
			authenticatedUser.Role,
		)
	}
}

func TestTokenServiceRejectsWrongIssuer(t *testing.T) {
	issuerA := NewTokenService(
		"test-secret-that-is-at-least-thirty-two-characters",
		"issuer-a",
		time.Hour,
	)

	issuerB := NewTokenService(
		"test-secret-that-is-at-least-thirty-two-characters",
		"issuer-b",
		time.Hour,
	)

	user := domain.User{
		ID:          uuid.New(),
		Email:       "admin@example.com",
		DisplayName: "Test Admin",
		Role:        domain.RoleAdmin,
	}

	token, _, err := issuerA.Generate(user)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	_, err = issuerB.Parse(token)
	if err == nil {
		t.Fatal("expected wrong issuer token to be rejected")
	}
}
