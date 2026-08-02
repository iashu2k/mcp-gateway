package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/iashu2k/mcp-gateway/backend/internal/domain"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("expired token")
)

type Claims struct {
	Email       string `json:"email"`
	DisplayName string `json:"displayName"`
	Role        string `json:"role"`
	jwt.RegisteredClaims
}

type TokenService struct {
	secret []byte
	issuer string
	ttl    time.Duration
	now    func() time.Time
}

func NewTokenService(
	secret string,
	issuer string,
	ttl time.Duration,
) *TokenService {
	return &TokenService{
		secret: []byte(secret),
		issuer: issuer,
		ttl:    ttl,
		now:    time.Now,
	}
}

func (s *TokenService) Generate(
	user domain.User,
) (string, time.Time, error) {
	now := s.now().UTC()
	expiresAt := now.Add(s.ttl)

	claims := Claims{
		Email:       user.Email,
		DisplayName: user.DisplayName,
		Role:        user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.ID.String(),
			Issuer:    s.issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err := token.SignedString(s.secret)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("sign JWT: %w", err)
	}

	return signedToken, expiresAt, nil
}

func (s *TokenService) Parse(
	rawToken string,
) (domain.AuthenticatedUser, error) {
	token, err := jwt.ParseWithClaims(
		rawToken,
		&Claims{},
		func(token *jwt.Token) (any, error) {
			if token.Method.Alg() != jwt.SigningMethodHS256.Alg() {
				return nil, fmt.Errorf("%w: unexpected signing method", ErrInvalidToken)
			}

			return s.secret, nil
		},
		jwt.WithIssuer(s.issuer),
		jwt.WithExpirationRequired(),
	)
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return domain.AuthenticatedUser{}, ErrExpiredToken
		}

		return domain.AuthenticatedUser{}, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return domain.AuthenticatedUser{}, ErrInvalidToken
	}

	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return domain.AuthenticatedUser{}, ErrInvalidToken
	}

	if !isValidRole(claims.Role) {
		return domain.AuthenticatedUser{}, ErrInvalidToken
	}

	return domain.AuthenticatedUser{
		ID:          userID,
		Email:       claims.Email,
		DisplayName: claims.DisplayName,
		Role:        claims.Role,
	}, nil
}

func isValidRole(role string) bool {
	switch role {
	case domain.RoleAdmin, domain.RoleDeveloper, domain.RoleViewer:
		return true
	default:
		return false
	}
}
