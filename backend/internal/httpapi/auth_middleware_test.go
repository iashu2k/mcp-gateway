package httpapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/iashu2k/mcp-gateway/backend/internal/domain"
)

func TestExtractBearerToken(t *testing.T) {
	token, err := extractBearerToken("Bearer example-token")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if token != "example-token" {
		t.Fatalf("expected token %q, got %q", "example-token", token)
	}
}

func TestExtractBearerTokenRejectsInvalidHeader(t *testing.T) {
	_, err := extractBearerToken("Basic example-token")
	if err == nil {
		t.Fatal("expected invalid authorization header error")
	}
}

func TestRequireRolesRejectsUnauthorizedRole(t *testing.T) {
	handler := RequireRoles(domain.RoleAdmin)(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		}),
	)

	request := httptest.NewRequest(http.MethodGet, "/admin", nil)
	request = withAuthenticatedUser(
		request,
		domain.AuthenticatedUser{
			Role: domain.RoleViewer,
		},
	)

	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusForbidden,
			recorder.Code,
		)
	}
}

func withAuthenticatedUser(
	request *http.Request,
	user domain.AuthenticatedUser,
) *http.Request {
	ctx := request.Context()
	ctx = context.WithValue(ctx, authenticatedUserContextKey, user)

	return request.WithContext(ctx)
}
