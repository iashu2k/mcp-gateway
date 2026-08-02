package httpapi

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/iashu2k/mcp-gateway/backend/internal/auth"
	"github.com/iashu2k/mcp-gateway/backend/internal/domain"
)

type contextKey string

const authenticatedUserContextKey contextKey = "authenticated_user"

type TokenParser interface {
	Parse(rawToken string) (domain.AuthenticatedUser, error)
}

func RequireAuthentication(
	tokenParser TokenParser,
) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rawToken, err := extractBearerToken(r.Header.Get("Authorization"))
			if err != nil {
				writeError(
					w,
					http.StatusUnauthorized,
					"missing_or_invalid_authorization",
					"authorization header must use Bearer token format",
					nil,
				)
				return
			}

			user, err := tokenParser.Parse(rawToken)
			if err != nil {
				errorCode := "invalid_token"
				message := "access token is invalid"

				if errors.Is(err, auth.ErrExpiredToken) {
					errorCode = "expired_token"
					message = "access token has expired"
				}

				writeError(
					w,
					http.StatusUnauthorized,
					errorCode,
					message,
					nil,
				)
				return
			}

			ctx := context.WithValue(
				r.Context(),
				authenticatedUserContextKey,
				user,
			)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func RequireRoles(
	allowedRoles ...string,
) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, ok := AuthenticatedUserFromContext(r.Context())
			if !ok {
				writeError(
					w,
					http.StatusUnauthorized,
					"unauthenticated",
					"authentication is required",
					nil,
				)
				return
			}

			for _, role := range allowedRoles {
				if user.Role == role {
					next.ServeHTTP(w, r)
					return
				}
			}

			writeError(
				w,
				http.StatusForbidden,
				"forbidden",
				"you do not have permission to perform this action",
				nil,
			)
		})
	}
}

func AuthenticatedUserFromContext(
	ctx context.Context,
) (domain.AuthenticatedUser, bool) {
	user, ok := ctx.Value(authenticatedUserContextKey).(domain.AuthenticatedUser)

	return user, ok
}

func extractBearerToken(header string) (string, error) {
	parts := strings.Fields(header)

	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", errors.New("invalid authorization header")
	}

	if parts[1] == "" {
		return "", errors.New("empty bearer token")
	}

	return parts[1], nil
}
