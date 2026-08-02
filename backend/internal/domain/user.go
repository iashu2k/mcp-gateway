package domain

import (
	"time"

	"github.com/google/uuid"
)

const (
	RoleAdmin     = "admin"
	RoleDeveloper = "developer"
	RoleViewer    = "viewer"
)

type User struct {
	ID           uuid.UUID `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	DisplayName  string    `json:"displayName"`
	Role         string    `json:"role"`
	Active       bool      `json:"active"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthenticatedUser struct {
	ID          uuid.UUID `json:"id"`
	Email       string    `json:"email"`
	DisplayName string    `json:"displayName"`
	Role        string    `json:"role"`
}

func (u User) ToAuthenticatedUser() AuthenticatedUser {
	return AuthenticatedUser{
		ID:          u.ID,
		Email:       u.Email,
		DisplayName: u.DisplayName,
		Role:        u.Role,
	}
}

type LoginResponse struct {
	AccessToken string            `json:"accessToken"`
	TokenType   string            `json:"tokenType"`
	ExpiresAt   time.Time         `json:"expiresAt"`
	User        AuthenticatedUser `json:"user"`
}
