package router

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/nexctl/nexctl/server/internal/alert"
	"github.com/nexctl/nexctl/server/internal/api/handler"
	apimiddleware "github.com/nexctl/nexctl/server/internal/api/middleware"
	"github.com/nexctl/nexctl/server/internal/audit"
	"github.com/nexctl/nexctl/server/internal/auth"
	"github.com/nexctl/nexctl/server/internal/config"
	"github.com/nexctl/nexctl/server/internal/filemgr"
	"github.com/nexctl/nexctl/server/internal/node"
	"github.com/nexctl/nexctl/server/internal/runtime"
	"github.com/nexctl/nexctl/server/internal/task"
	"github.com/nexctl/nexctl/server/internal/upgrade"
	"github.com/nexctl/nexctl/server/internal/ws"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// New creates the HTTP router.
func New(cfg config.Config, logger *zap.Logger, authService *auth.Service, nodeService *node.Service, runtimeService *runtime.Service, taskService *task.Service, fileService *filemgr.Service, upgradeService *upgrade.Service, alertService *alert.Service, auditService *audit.Service, wsService *ws.Service, db *sql.DB, rdb *redis.Client) http.Handler {
	authHandler := handler.NewAuthHandler(authService)
	nodeHandler := handler.NewNodeHandler(nodeService, runtimeService)
	healthHandler := handler.NewHealthHandler(db, rdb)
	moduleHandler := handler.NewModuleHandler(taskService, fileService, upgradeService, alertService, auditService)
	wsHandler := handler.NewWSHandler(nodeService, wsService, logger, cfg.App.WebSocketAllowedOrigins)

	authMiddleware := apimiddleware.NewAuthMiddleware(cfg.Auth)

	r := chi.NewRouter()
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.Timeout(30 * time.Second))
	r.Use(apimiddleware.Logging(logger))
	r.Use(apimiddleware.Recover(logger))

	r.Get("/healthz", healthHandler.Healthz)

	r.Route("/api/v1", func(api chi.Router) {
		loginRL := apimiddleware.RateLimitFunc(60*time.Second/20, 8, 4096)
		registerRL := apimiddleware.RateLimitFunc(60*time.Second/40, 15, 4096)
		api.Post("/auth/login", loginRL(authHandler.Login))
		api.Post("/agents/register", registerRL(nodeHandler.Register))
		api.Get("/agents/ws", wsHandler.AgentConnect)

		api.Group(func(protected chi.Router) {
			protected.Use(authMiddleware.RequireLogin)

			protected.Get("/me", nodeHandler.CurrentUser)

			protected.Group(func(nodeRoutes chi.Router) {
				nodeRoutes.Use(authMiddleware.RequirePermission("nodes:read"))
				nodeRoutes.Get("/nodes", nodeHandler.List)
				nodeRoutes.Get("/nodes/{nodeID}", nodeHandler.Detail)
			})

			protected.Group(func(nodeWrite chi.Router) {
				nodeWrite.Use(authMiddleware.RequirePermission("nodes:write"))
				nodeWrite.Post("/nodes", nodeHandler.CreatePending)
				nodeWrite.Post("/nodes/{nodeID}/runtime-state", nodeHandler.UpdateRuntimeState)
				nodeWrite.Delete("/nodes/{nodeID}", nodeHandler.Delete)
			})

			protected.Group(func(mod chi.Router) {
				mod.Use(authMiddleware.RequirePermission("modules:read"))
				mod.Get("/tasks", moduleHandler.ListTasks)
				mod.Get("/files", moduleHandler.ListFiles)
				mod.Get("/upgrades/releases", moduleHandler.ListReleases)
				mod.Get("/alerts/rules", moduleHandler.ListAlertRules)
				mod.Get("/alerts/events", moduleHandler.ListAlertEvents)
				mod.Get("/audit/logs", moduleHandler.ListAuditLogs)
			})
		})
	})

	return r
}
