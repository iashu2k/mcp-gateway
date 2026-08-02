package httpapi

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/iashu2k/mcp-gateway/backend/internal/auth"
	"github.com/iashu2k/mcp-gateway/backend/internal/config"
	"github.com/iashu2k/mcp-gateway/backend/internal/domain"
	"github.com/iashu2k/mcp-gateway/backend/internal/repository"
	"github.com/iashu2k/mcp-gateway/backend/internal/service"
	"github.com/jackc/pgx/v5/pgxpool"
)

func NewRouter(
	logger *slog.Logger,
	db *pgxpool.Pool,
	cfg config.Config,
) http.Handler {
	serverRepository := repository.NewServerRepository(db)
	toolRepository := repository.NewToolRepository(db)
	userRepository := repository.NewUserRepository(db)

	serverService := service.NewServerService(serverRepository)
	toolService := service.NewToolService(toolRepository, serverRepository)

	tokenService := auth.NewTokenService(
		cfg.JWTSecret,
		cfg.JWTIssuer,
		time.Duration(cfg.JWTTTLMinutes)*time.Minute,
	)
	authService := service.NewAuthService(userRepository, tokenService)

	serverHandler := NewServerHandler(serverService)
	toolHandler := NewToolHandler(toolService)
	authHandler := NewAuthHandler(authService)

	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Recoverer)
	router.Use(requestLogger(logger))
	router.Use(middleware.Timeout(15 * time.Second))

	router.Get("/health", Handler{DB: db}.Health)

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
				r.Get("/", serverHandler.List)
				r.Get("/{serverID}", serverHandler.GetByID)

				r.Get("/{serverID}/tools/", toolHandler.ListByServerID)
				r.Get("/{serverID}/tools/{toolID}", toolHandler.GetByID)

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

			next.ServeHTTP(w, r)

			logger.Info("http request completed",
				"request_id", middleware.GetReqID(r.Context()),
				"method", r.Method,
				"path", r.URL.Path,
				"duration_ms", time.Since(start).Milliseconds(),
			)
		})
	}
}
