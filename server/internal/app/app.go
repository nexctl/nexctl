package app

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/nexctl/nexctl/server/internal/alert"
	"github.com/nexctl/nexctl/server/internal/api/router"
	"github.com/nexctl/nexctl/server/internal/audit"
	"github.com/nexctl/nexctl/server/internal/auth"
	"github.com/nexctl/nexctl/server/internal/config"
	"github.com/nexctl/nexctl/server/internal/filemgr"
	"github.com/nexctl/nexctl/server/internal/node"
	"github.com/nexctl/nexctl/server/internal/repository"
	"github.com/nexctl/nexctl/server/internal/runtime"
	"github.com/nexctl/nexctl/server/internal/task"
	"github.com/nexctl/nexctl/server/internal/upgrade"
	"github.com/nexctl/nexctl/server/internal/ws"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// App is the NexCtl server application.
type App struct {
	cfg           config.Config
	logger        *zap.Logger
	server        *http.Server
	db            *sql.DB
	rdb           *redis.Client
	statusManager *StatusManager
}

// New creates a new server application.
func New(ctx context.Context, configPath string) (*App, error) {
	cfg, err := config.Load(configPath)
	if err != nil {
		return nil, err
	}
	logger, err := newLogger(cfg.App.Env)
	if err != nil {
		return nil, err
	}
	db, err := repository.NewMySQL(ctx, cfg.MySQL)
	if err != nil {
		return nil, err
	}
	rdb, err := repository.NewRedis(ctx, cfg.Redis)
	if err != nil {
		return nil, err
	}

	userRepo := repository.NewUserRepository(db)
	installTokenRepo := repository.NewInstallTokenRepository(db)
	nodeRepo := repository.NewNodeRepository(db)
	runtimeRepo := repository.NewRuntimeStateRepository(db, rdb, cfg.Node.RuntimePointsTTLSeconds, cfg.Node.RuntimePointsMaxCount)
	auditRepo := repository.NewAuditRepository(db)
	sessionCache := repository.NewNodeSessionCache(rdb)

	auditService := audit.NewService(auditRepo, logger)
	authService := auth.NewService(cfg.Auth, userRepo, auditService)
	runtimeService := runtime.NewService(runtimeRepo, nodeRepo)
	nodeService := node.NewService(cfg.Node, installTokenRepo, nodeRepo, runtimeRepo, auditService, cfg.App.ExternalURL)
	taskService := task.NewService()
	fileService := filemgr.NewService()
	upgradeService := upgrade.NewService()
	alertService := alert.NewService()
	wsService := ws.NewService(cfg.Node, nodeRepo, runtimeService, sessionCache, auditService, logger)

	httpServer := &http.Server{
		Addr:              cfg.App.ListenAddr,
		Handler:           router.New(cfg, logger, authService, nodeService, runtimeService, taskService, fileService, upgradeService, alertService, auditService, wsService, db, rdb),
		ReadHeaderTimeout: 10 * time.Second,
	}

	return &App{
		cfg:           cfg,
		logger:        logger,
		server:        httpServer,
		db:            db,
		rdb:           rdb,
		statusManager: NewStatusManager(cfg.Node, nodeRepo, logger),
	}, nil
}

// Run starts the HTTP server and background tasks.
func (a *App) Run(ctx context.Context) error {
	errCh := make(chan error, 1)
	go a.statusManager.Run(ctx)
	go func() {
		a.logger.Info("server starting", zap.String("addr", a.cfg.App.ListenAddr))
		if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- fmt.Errorf("listen and serve: %w", err)
		}
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), a.cfg.ShutdownTimeout())
		defer cancel()
		_ = a.server.Shutdown(shutdownCtx)
		_ = a.rdb.Close()
		return a.db.Close()
	case err := <-errCh:
		return err
	}
}
