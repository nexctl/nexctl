package app

import (
	"context"
	"time"

	"github.com/nexctl/nexctl/server/internal/config"
	"go.uber.org/zap"
)

// StatusNodeRepository updates node status based on heartbeat freshness.
type StatusNodeRepository interface {
	MarkTimedOutNodes(ctx context.Context, unstableBefore, offlineBefore time.Time) error
}

// StatusManager periodically recalculates node online state.
type StatusManager struct {
	cfg    config.NodeConfig
	nodes  StatusNodeRepository
	logger *zap.Logger
}

// NewStatusManager creates a node status manager.
func NewStatusManager(cfg config.NodeConfig, nodes StatusNodeRepository, logger *zap.Logger) *StatusManager {
	return &StatusManager{cfg: cfg, nodes: nodes, logger: logger}
}

// Run starts the status management loop.
func (m *StatusManager) Run(ctx context.Context) {
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			now := time.Now().UTC()
			unstableBefore := now.Add(-time.Duration(m.cfg.UnstableTimeoutSeconds) * time.Second)
			offlineBefore := now.Add(-time.Duration(m.cfg.HeartbeatTimeoutSeconds) * time.Second)
			if err := m.nodes.MarkTimedOutNodes(ctx, unstableBefore, offlineBefore); err != nil {
				m.logger.Warn("mark timed out nodes", zap.Error(err))
			}
		}
	}
}
