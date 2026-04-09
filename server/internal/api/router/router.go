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
	nodeHandler := handler.NewNodeHandler(nodeService, runtimeService, wsService)
	healthHandler := handler.NewHealthHandler(db, rdb)
	moduleHandler := handler.NewModuleHandler(taskService, fileService, upgradeService, alertService, auditService)
	wsHandler := handler.NewWSHandler(nodeService, wsService, taskService, logger, cfg.App.WebSocketAllowedOrigins)
	terminalWSHandler := handler.NewTerminalWSHandler(cfg.Auth, wsService, logger, cfg.App.WebSocketAllowedOrigins)

	authMiddleware := apimiddleware.NewAuthMiddleware(cfg.Auth)

	r := chi.NewRouter()
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	// 根路由不设短 Timeout，避免浏览器终端 WebSocket 长连接被中断；短超时仅挂在下方纯 HTTP API 子树。
	r.Use(apimiddleware.Logging(logger))
	r.Use(apimiddleware.Recover(logger))

	r.Get("/healthz", healthHandler.Healthz)

	r.Route("/api/v1", func(api chi.Router) {
		loginRL := apimiddleware.RateLimitFunc(60*time.Second/20, 8, 4096)
		api.Post("/auth/login", loginRL(authHandler.Login))
		api.Get("/agents/ws", wsHandler.AgentConnect)
		api.Get("/nodes/{nodeID}/terminal/ws", terminalWSHandler.ServeWS)

		api.Group(func(protected chi.Router) {
			protected.Use(chimiddleware.Timeout(30 * time.Second))
			protected.Use(authMiddleware.RequireLogin)

			protected.Get("/me", nodeHandler.CurrentUser)

			// 所有 /nodes 路由挂在同一 Route 子树下，避免多个 Group 并列时 DELETE 未正确注册（表现为 404）。
			protected.Route("/nodes", func(nr chi.Router) {
				nr.With(authMiddleware.RequirePermission("nodes:read")).Get("/", nodeHandler.List)
				nr.With(authMiddleware.RequirePermission("nodes:read")).Get("/{nodeID}", nodeHandler.Detail)
				nr.With(authMiddleware.RequirePermission("nodes:write")).Post("/", nodeHandler.CreatePending)
				nr.With(authMiddleware.RequirePermission("nodes:read")).Get("/{nodeID}/agent-credentials", nodeHandler.AgentCredentials)
				nr.With(authMiddleware.RequirePermission("nodes:write")).Post("/{nodeID}/upgrade-agent", nodeHandler.TriggerAgentUpgrade)
				nr.With(authMiddleware.RequirePermission("nodes:write")).Post("/{nodeID}/file-op", nodeHandler.FileOp)
				nr.With(authMiddleware.RequirePermission("nodes:write")).Post("/{nodeID}/runtime-state", nodeHandler.UpdateRuntimeState)
				nr.With(authMiddleware.RequirePermission("nodes:write")).Delete("/{nodeID}", nodeHandler.Delete)
			})

			protected.Route("/task-schedules", func(sr chi.Router) {
				sr.With(authMiddleware.RequirePermission("modules:read")).Get("/", moduleHandler.ListTaskSchedules)
				sr.With(authMiddleware.RequirePermission("modules:write")).Post("/", moduleHandler.CreateTaskSchedule)
			})

			protected.Route("/tasks", func(tr chi.Router) {
				tr.With(authMiddleware.RequirePermission("modules:read")).Get("/", moduleHandler.ListTasks)
				tr.With(authMiddleware.RequirePermission("modules:write")).Post("/", moduleHandler.CreateTask)
				tr.With(authMiddleware.RequirePermission("modules:read")).Get("/{taskID}", moduleHandler.GetTask)
			})

			protected.Group(func(mod chi.Router) {
				mod.Use(authMiddleware.RequirePermission("modules:read"))
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
