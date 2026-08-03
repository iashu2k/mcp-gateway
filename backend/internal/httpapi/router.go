package httpapi

import (
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/iashu2k/mcp-gateway/backend/internal/auth"
	"github.com/iashu2k/mcp-gateway/backend/internal/config"
	"github.com/iashu2k/mcp-gateway/backend/internal/domain"
	"github.com/iashu2k/mcp-gateway/backend/internal/executor"
	"github.com/iashu2k/mcp-gateway/backend/internal/observability"
	"github.com/iashu2k/mcp-gateway/backend/internal/repository"
	"github.com/iashu2k/mcp-gateway/backend/internal/service"
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

	invocationHistoryService := service.NewInvocationHistoryService(invocationRepository)

	// -------------------------------------------------------------------------
	// HTTP handlers
	// -------------------------------------------------------------------------

	authHandler := NewAuthHandler(authService)
	serverHandler := NewServerHandler(serverService)
	toolHandler := NewToolHandler(toolService)
	invocationHandler := NewInvocationHandler(invocationService)
	invocationHistoryHandler := NewInvocationHistoryHandler(invocationHistoryService)

	// -------------------------------------------------------------------------
	// Router and global middleware
	// -------------------------------------------------------------------------

	router := chi.NewRouter()
	// CORS middleware (must be first)
	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173", "http://localhost:3000"},
		AllowedMethods:   []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300, // 5 minutes
	}))


	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Recoverer)
	router.Use(requestLogger(logger))
	router.Use(middleware.Timeout(15 * time.Second))

	router.Get("/health", Handler{DB: db}.Health)

	// Metrics endpoint (public for Prometheus scraping)
	router.Get("/metrics", promhttp.Handler().ServeHTTP)

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

			// Phase 6: Invocation history
			r.Route("/invocations", func(r chi.Router) {
				r.Get("/", invocationHistoryHandler.List)
				r.Get("/{invocationID}", invocationHistoryHandler.GetByID)
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

			duration := time.Since(start)

			// Structured log
			logger.Info("http request completed",
				"request_id", middleware.GetReqID(r.Context()),
				"method", r.Method,
				"path", r.URL.Path,
				"status", status,
				"bytes", ww.BytesWritten(),
				"duration_ms", duration.Milliseconds(),
			)

			// Prometheus metrics
			observability.HTTPRequestsTotal.WithLabelValues(
				r.Method,
				normalizePath(r.URL.Path),
				strconv.Itoa(status),
			).Inc()

			observability.HTTPRequestDuration.WithLabelValues(
				r.Method,
				normalizePath(r.URL.Path),
			).Observe(duration.Seconds())
		})
	}
}

// normalizePath reduces path cardinality for metrics by replacing UUIDs with placeholders.
func normalizePath(path string) string {
	// Replace UUIDs with placeholder to avoid high cardinality in metrics
	// Pattern: /api/v1/servers/{uuid}/tools/{uuid}/invoke -> /api/v1/servers/:id/tools/:id/invoke
	segments := splitPath(path)
	for i, seg := range segments {
		if isUUID(seg) {
			segments[i] = ":id"
		}
	}
	return joinPath(segments)
}

func splitPath(path string) []string {
	var segments []string
	current := ""
	for _, c := range path {
		if c == '/' {
			if current != "" {
				segments = append(segments, current)
				current = ""
			}
		} else {
			current += string(c)
		}
	}
	if current != "" {
		segments = append(segments, current)
	}
	return segments
}

func joinPath(segments []string) string {
	if len(segments) == 0 {
		return "/"
	}
	result := ""
	for _, seg := range segments {
		result += "/" + seg
	}
	return result
}

func isUUID(s string) bool {
	if len(s) != 36 {
		return false
	}
	for i, c := range s {
		switch i {
		case 8, 13, 18, 23:
			if c != '-' {
				return false
			}
		default:
			if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
				return false
			}
		}
	}
	return true
}