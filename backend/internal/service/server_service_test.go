package service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/iashu2k/mcp-gateway/backend/internal/domain"
	"github.com/iashu2k/mcp-gateway/backend/internal/repository"
)

type fakeServerStore struct {
	createFn func(
		ctx context.Context,
		server domain.MCPServer,
	) (domain.MCPServer, error)
}

func (f fakeServerStore) Create(
	ctx context.Context,
	server domain.MCPServer,
) (domain.MCPServer, error) {
	return f.createFn(ctx, server)
}

func (f fakeServerStore) List(
	context.Context,
) ([]domain.MCPServer, error) {
	return nil, nil
}

func (f fakeServerStore) GetByID(
	context.Context,
	uuid.UUID,
) (domain.MCPServer, error) {
	return domain.MCPServer{}, repository.ErrServerNotFound
}

func (f fakeServerStore) Update(
	context.Context,
	domain.MCPServer,
) (domain.MCPServer, error) {
	return domain.MCPServer{}, nil
}

func (f fakeServerStore) Delete(context.Context, uuid.UUID) error {
	return nil
}

func TestCreateServerUsesDefaultTransportAndActiveStatus(t *testing.T) {
	store := fakeServerStore{
		createFn: func(
			_ context.Context,
			server domain.MCPServer,
		) (domain.MCPServer, error) {
			if server.TransportType != domain.TransportStreamableHTTP {
				t.Fatalf(
					"expected transport %q, got %q",
					domain.TransportStreamableHTTP,
					server.TransportType,
				)
			}

			if server.Status != domain.ServerStatusActive {
				t.Fatalf(
					"expected status %q, got %q",
					domain.ServerStatusActive,
					server.Status,
				)
			}

			server.ID = uuid.New()
			return server, nil
		},
	}

	service := NewServerService(store)

	server, err := service.Create(context.Background(), domain.CreateMCPServerRequest{
		Name:        "github-mcp",
		Description: "GitHub tools",
		BaseURL:     "http://localhost:3001",
		OwnerTeam:   "developer-platform",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if server.ID == uuid.Nil {
		t.Fatal("expected generated server ID")
	}
}

func TestCreateServerRejectsInvalidURL(t *testing.T) {
	store := fakeServerStore{
		createFn: func(
			_ context.Context,
			server domain.MCPServer,
		) (domain.MCPServer, error) {
			return server, nil
		},
	}

	service := NewServerService(store)

	_, err := service.Create(context.Background(), domain.CreateMCPServerRequest{
		Name:        "bad-server",
		Description: "Invalid URL test",
		BaseURL:     "not-a-url",
		OwnerTeam:   "developer-platform",
	})

	var validationError ValidationError
	if !errors.As(err, &validationError) {
		t.Fatalf("expected ValidationError, got %v", err)
	}
}

func TestCreateServerRequiresName(t *testing.T) {
	store := fakeServerStore{
		createFn: func(
			_ context.Context,
			server domain.MCPServer,
		) (domain.MCPServer, error) {
			return server, nil
		},
	}

	service := NewServerService(store)

	_, err := service.Create(context.Background(), domain.CreateMCPServerRequest{
		Description: "Missing name",
		BaseURL:     "http://localhost:3001",
		OwnerTeam:   "developer-platform",
	})

	var validationError ValidationError
	if !errors.As(err, &validationError) {
		t.Fatalf("expected ValidationError, got %v", err)
	}
}
