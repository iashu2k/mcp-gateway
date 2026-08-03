package httpapi

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/iashu2k/mcp-gateway/backend/internal/auth"
	"github.com/iashu2k/mcp-gateway/backend/internal/config"
	"github.com/iashu2k/mcp-gateway/backend/internal/domain"
	"github.com/iashu2k/mcp-gateway/backend/internal/executor"
	"github.com/iashu2k/mcp-gateway/backend/internal/repository"
	"github.com/iashu2k/mcp-gateway/backend/internal/service"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
)

func NewRouter(
	logger *slog.Logger,
	db *pgxpool.Pool,
	cfg config.Config,
) http.Handler {
	// -------------------------------------------------------------------------
	// Repositories
	// -------------------------------------------------------------------------

	serverRepository := repository.NewServerRepository(db)
	toolRepository := repository.NewToolRepository(db)
	userRepository := repository.NewUserRepository(db)
	invocationRepository := repository.NewInvocationRepository(db)

	// -------------------------------------------------------------------------
	// Services
	// -------------------------------------------------------------------------

	serverService := service.NewServerService(serverRepository)

	toolService := service.NewToolService(
		toolRepository,
		serverRepository,
	)

	tokenService := auth.NewTokenService(
		cfg.JWTSecret,
		cfg.JWTIssuer,
		time.Duration(cfg.JWTTTLMinutes)*time.Minute,
	)

	authService := service.NewAuthService(
		userRepository,
		tokenService,
	)

	schemaValidator := service.NewJSONSchemaValidator()
	toolExecutor := executor.NewRouterExecutor(
		executor.NewMockExecutor(),
		executor.NewGitHubExecutor(cfg.GitHubToken),
	)

	invocationService := service.NewInvocationService(
		invocationRepository,
		serverRepository,
		toolRepository,
		schemaValidator,
		toolExecutor,
	)

	// -------------------------------------------------------------------------
	// HTTP handlers
	// -------------------------------------------------------------------------

	authHandler := NewAuthHandler(authService)
	serverHandler := NewServerHandler(serverService)
	toolHandler := NewToolHandler(toolService)
	invocationHandler := NewInvocationHandler(invocationService)

	// -------------------------------------------------------------------------
	// Router and global middleware
	// -------------------------------------------------------------------------

	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Recoverer)
	router.Use(requestLogger(logger))
	router.Use(middleware.Timeout(15 * time.Second))
	// router.Use(middleware.StripSlashes)

	router.Get("/health", Handler{DB: db}.Health)

	// -------------------------------------------------------------------------
	// API v1 routes
	// -------------------------------------------------------------------------

	router.Route("/api/v1", func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			writeJSON(w, http.StatusOK, map[string]string{
				"message": "MCP Gateway API",
			})
		})

		r.Post("/auth/login", authHandler.Login)

		r.Group(func(r chi.Router) {
			r.Use(RequireAuthentication(tokenService))

			r.Get("/auth/me", authHandler.Me)

			r.Route("/servers", func(r chi.Router) {
				// Catalog read routes are available to every authenticated role:
				// admin, developer, and viewer.
				r.Get("/", serverHandler.List)
				r.Get("/{serverID}", serverHandler.GetByID)
				r.Get("/{serverID}/tools/", toolHandler.ListByServerID)
				r.Get("/{serverID}/tools/{toolID}", toolHandler.GetByID)

				// Phase 4: low-risk tool invocation.
				// Admin and developer can invoke; viewer cannot.
				r.With(
					RequireRoles(
						domain.RoleAdmin,
						domain.RoleDeveloper,
					),
				).Post(
					"/{serverID}/tools/{toolID}/invoke",
					invocationHandler.Invoke,
				)

				// Catalog mutation routes are restricted to administrators.
				r.Group(func(r chi.Router) {
					r.Use(RequireRoles(domain.RoleAdmin))

					r.Post("/", serverHandler.Create)
					r.Patch("/{serverID}", serverHandler.Update)
					r.Delete("/{serverID}", serverHandler.Delete)

					r.Post("/{serverID}/tools/", toolHandler.Create)
					r.Patch(
						"/{serverID}/tools/{toolID}",
						toolHandler.Update,
					)
					r.Delete(
						"/{serverID}/tools/{toolID}",
						toolHandler.Delete,
					)
				})
			})
		})
	})

	return router
}

func requestLogger(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			next.ServeHTTP(ww, r)

			status := ww.Status()
			if status == 0 {
				status = http.StatusOK
			}

			logger.Info("http request completed",
				"request_id", middleware.GetReqID(r.Context()),
				"method", r.Method,
				"path", r.URL.Path,
				"status", status,
				"bytes", ww.BytesWritten(),
				"duration_ms", time.Since(start).Milliseconds(),
			)
		})
	}
}
