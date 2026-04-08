package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/nexctl/nexctl/server/internal/model"
	"github.com/redis/go-redis/v9"
)

// RuntimeStateRepository defines node runtime state access.
type RuntimeStateRepository interface {
	Upsert(ctx context.Context, state *model.NodeRuntimeState) error
	GetByNodeID(ctx context.Context, nodeID int64) (*model.NodeRuntimeState, error)
	// DeleteForNode removes Redis short-term metric keys for a node (MySQL row may be removed via FK CASCADE).
	DeleteForNode(ctx context.Context, nodeID int64) error
}

// MySQLRuntimeStateRepository is the runtime state repository with Redis cache.
type MySQLRuntimeStateRepository struct {
	db             *sql.DB
	rdb            *redis.Client
	pointsTTL      time.Duration
	pointsMaxCount int64
}

// NewRuntimeStateRepository creates a runtime state repository.
func NewRuntimeStateRepository(db *sql.DB, rdb *redis.Client, ttlSeconds, maxCount int) *MySQLRuntimeStateRepository {
	return &MySQLRuntimeStateRepository{db: db, rdb: rdb, pointsTTL: time.Duration(ttlSeconds) * time.Second, pointsMaxCount: int64(maxCount)}
}

// Upsert upserts the current runtime state and appends a short-term Redis point.
func (r *MySQLRuntimeStateRepository) Upsert(ctx context.Context, state *model.NodeRuntimeState) error {
	const query = `INSERT INTO node_runtime_states (node_id, cpu_percent, memory_percent, disk_percent, network_rx_bps, network_tx_bps, load_1, load_5, load_15, uptime_seconds, process_count, reported_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW()) ON DUPLICATE KEY UPDATE cpu_percent = VALUES(cpu_percent), memory_percent = VALUES(memory_percent), disk_percent = VALUES(disk_percent), network_rx_bps = VALUES(network_rx_bps), network_tx_bps = VALUES(network_tx_bps), load_1 = VALUES(load_1), load_5 = VALUES(load_5), load_15 = VALUES(load_15), uptime_seconds = VALUES(uptime_seconds), process_count = VALUES(process_count), reported_at = VALUES(reported_at), updated_at = NOW()`
	if _, err := r.db.ExecContext(ctx, query, state.NodeID, state.CPUPercent, state.MemoryPercent, state.DiskPercent, state.NetworkRxBps, state.NetworkTxBps, state.Load1, state.Load5, state.Load15, state.UptimeSeconds, state.ProcessCount, state.ReportedAt); err != nil {
		return fmt.Errorf("upsert runtime state: %w", err)
	}
	return r.appendShortTermPoint(ctx, state)
}

// GetByNodeID returns the latest runtime state by node ID.
func (r *MySQLRuntimeStateRepository) GetByNodeID(ctx context.Context, nodeID int64) (*model.NodeRuntimeState, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, node_id, cpu_percent, memory_percent, disk_percent, network_rx_bps, network_tx_bps, load_1, load_5, load_15, uptime_seconds, process_count, reported_at, updated_at FROM node_runtime_states WHERE node_id = ?`, nodeID)
	var item model.NodeRuntimeState
	if err := row.Scan(&item.ID, &item.NodeID, &item.CPUPercent, &item.MemoryPercent, &item.DiskPercent, &item.NetworkRxBps, &item.NetworkTxBps, &item.Load1, &item.Load5, &item.Load15, &item.UptimeSeconds, &item.ProcessCount, &item.ReportedAt, &item.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get runtime state: %w", err)
	}
	return &item, nil
}

// DeleteForNode deletes Redis keys tied to a node (runtime points + online marker; formats align with appendShortTermPoint / NodeSessionCache).
func (r *MySQLRuntimeStateRepository) DeleteForNode(ctx context.Context, nodeID int64) error {
	pointsKey := fmt.Sprintf("node:%d:runtime_points", nodeID)
	onlineKey := fmt.Sprintf("node:%d:online", nodeID)
	pipe := r.rdb.TxPipeline()
	pipe.Del(ctx, pointsKey)
	pipe.Del(ctx, onlineKey)
	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("delete redis node keys: %w", err)
	}
	return nil
}

func (r *MySQLRuntimeStateRepository) appendShortTermPoint(ctx context.Context, state *model.NodeRuntimeState) error {
	key := fmt.Sprintf("node:%d:runtime_points", state.NodeID)
	raw, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("marshal runtime point: %w", err)
	}
	pipe := r.rdb.TxPipeline()
	pipe.LPush(ctx, key, raw)
	pipe.LTrim(ctx, key, 0, r.pointsMaxCount-1)
	pipe.Expire(ctx, key, r.pointsTTL)
	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("write runtime points to redis: %w", err)
	}
	return nil
}
