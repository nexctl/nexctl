package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/nexctl/nexctl/server/internal/model"
)

const nodeSelectCols = `id, agent_id, agent_secret, node_key, name, hostname, platform, platform_version, arch, agent_version, status, last_heartbeat_at, last_online_at, created_at, updated_at, enrollment_token_hash, enrollment_expires_at`

// NodeRepository defines node data access.
type NodeRepository interface {
	Create(ctx context.Context, node *model.Node) error
	CreatePendingEnrollment(ctx context.Context, node *model.Node, enrollmentTokenHash string, enrollmentExpiresAt *time.Time) error
	GetByID(ctx context.Context, id int64) (*model.Node, error)
	GetByEnrollmentTokenHash(ctx context.Context, hash string) (*model.Node, error)
	GetByAgentCredential(ctx context.Context, agentID, agentSecret string) (*model.Node, error)
	List(ctx context.Context) ([]*model.Node, error)
	UpdateHeartbeat(ctx context.Context, nodeID int64, seenAt time.Time, status string) error
	// UpdateAgentMeta 由 Agent runtime_state 上报，刷新控制台展示的 OS/架构/版本与 hostname。
	UpdateAgentMeta(ctx context.Context, nodeID int64, hostname, platform, platformVersion, arch, agentVersion string) error
	// SetPendingEnrollmentToken replaces enrollment hash/expiry for a row that is still pending (awaiting agent).
	SetPendingEnrollmentToken(ctx context.Context, nodeID int64, enrollmentTokenHash string, enrollmentExpiresAt *time.Time) error
	CompleteEnrollment(ctx context.Context, node *model.Node) error
	MarkTimedOutNodes(ctx context.Context, unstableBefore, offlineBefore time.Time) error
	DeleteByID(ctx context.Context, id int64) error
}

// MySQLNodeRepository is the MySQL node repository.
type MySQLNodeRepository struct {
	db *sql.DB
}

// NewNodeRepository creates a node repository.
func NewNodeRepository(db *sql.DB) *MySQLNodeRepository {
	return &MySQLNodeRepository{db: db}
}

// Create creates a node record.
func (r *MySQLNodeRepository) Create(ctx context.Context, node *model.Node) error {
	const query = `INSERT INTO nodes (agent_id, agent_secret, node_key, name, hostname, platform, platform_version, arch, agent_version, status, last_heartbeat_at, last_online_at, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())`
	result, err := r.db.ExecContext(ctx, query, node.AgentID, node.AgentSecret, node.NodeKey, node.Name, node.Hostname, node.Platform, node.PlatformVersion, node.Arch, node.AgentVersion, node.Status, node.LastHeartbeatAt, node.LastOnlineAt)
	if err != nil {
		return fmt.Errorf("create node: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("get node id: %w", err)
	}
	node.ID = id
	return nil
}

// CreatePendingEnrollment inserts a pre-created node row awaiting agent enrollment_token.
func (r *MySQLNodeRepository) CreatePendingEnrollment(ctx context.Context, node *model.Node, enrollmentTokenHash string, enrollmentExpiresAt *time.Time) error {
	const query = `INSERT INTO nodes (agent_id, agent_secret, node_key, name, hostname, platform, platform_version, arch, agent_version, status, last_heartbeat_at, last_online_at, enrollment_token_hash, enrollment_expires_at, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW(), ?, ?, NOW(), NOW())`
	result, err := r.db.ExecContext(ctx, query, node.AgentID, node.AgentSecret, node.NodeKey, node.Name, node.Hostname, node.Platform, node.PlatformVersion, node.Arch, node.AgentVersion, node.Status, enrollmentTokenHash, enrollmentExpiresAt)
	if err != nil {
		return fmt.Errorf("create pending node: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("get pending node id: %w", err)
	}
	node.ID = id
	return nil
}

// GetByID returns a node by ID.
func (r *MySQLNodeRepository) GetByID(ctx context.Context, id int64) (*model.Node, error) {
	return r.getOne(ctx, `SELECT `+nodeSelectCols+` FROM nodes WHERE id = ?`, id)
}

// GetByEnrollmentTokenHash returns a node awaiting enrollment with the given token hash.
func (r *MySQLNodeRepository) GetByEnrollmentTokenHash(ctx context.Context, hash string) (*model.Node, error) {
	return r.getOne(ctx, `SELECT `+nodeSelectCols+` FROM nodes WHERE enrollment_token_hash = ? AND (enrollment_expires_at IS NULL OR enrollment_expires_at > UTC_TIMESTAMP())`, hash)
}

// GetByAgentCredential returns a node by agent credential.
func (r *MySQLNodeRepository) GetByAgentCredential(ctx context.Context, agentID, agentSecret string) (*model.Node, error) {
	return r.getOne(ctx, `SELECT `+nodeSelectCols+` FROM nodes WHERE agent_id = ? AND agent_secret = ?`, agentID, agentSecret)
}

// List returns all nodes.
func (r *MySQLNodeRepository) List(ctx context.Context) ([]*model.Node, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT `+nodeSelectCols+` FROM nodes ORDER BY updated_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("list nodes: %w", err)
	}
	defer rows.Close()
	var items []*model.Node
	for rows.Next() {
		item, err := scanNode(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

// UpdateHeartbeat updates node heartbeat timestamps and status.
func (r *MySQLNodeRepository) UpdateHeartbeat(ctx context.Context, nodeID int64, seenAt time.Time, status string) error {
	if _, err := r.db.ExecContext(ctx, `UPDATE nodes SET status = ?, last_heartbeat_at = ?, last_online_at = ?, updated_at = NOW() WHERE id = ?`, status, seenAt, seenAt, nodeID); err != nil {
		return fmt.Errorf("update node heartbeat: %w", err)
	}
	return nil
}

// UpdateAgentMeta updates host identity fields reported by the agent.
func (r *MySQLNodeRepository) UpdateAgentMeta(ctx context.Context, nodeID int64, hostname, platform, platformVersion, arch, agentVersion string) error {
	if _, err := r.db.ExecContext(ctx, `UPDATE nodes SET hostname = ?, platform = ?, platform_version = ?, arch = ?, agent_version = ?, updated_at = NOW() WHERE id = ?`,
		hostname, platform, platformVersion, arch, agentVersion, nodeID); err != nil {
		return fmt.Errorf("update node agent meta: %w", err)
	}
	return nil
}

// DeleteByID removes a node row. Related node_runtime_states rows are removed by ON DELETE CASCADE.
func (r *MySQLNodeRepository) DeleteByID(ctx context.Context, id int64) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM nodes WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete node: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete node rows affected: %w", err)
	}
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// SetPendingEnrollmentToken updates enrollment fields for a pending node (e.g. re-issue install token).
func (r *MySQLNodeRepository) SetPendingEnrollmentToken(ctx context.Context, nodeID int64, enrollmentTokenHash string, enrollmentExpiresAt *time.Time) error {
	res, err := r.db.ExecContext(ctx, `UPDATE nodes SET enrollment_token_hash = ?, enrollment_expires_at = ?, updated_at = NOW() WHERE id = ? AND status = ?`, enrollmentTokenHash, enrollmentExpiresAt, nodeID, model.NodeStatusPending)
	if err != nil {
		return fmt.Errorf("set pending enrollment token: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("set pending enrollment rows: %w", err)
	}
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// CompleteEnrollment finalizes a pending node after agent presents a valid enrollment_token.
func (r *MySQLNodeRepository) CompleteEnrollment(ctx context.Context, n *model.Node) error {
	res, err := r.db.ExecContext(ctx, `
UPDATE nodes SET
  agent_id = ?, agent_secret = ?, node_key = ?, hostname = ?, platform = ?, platform_version = ?, arch = ?, agent_version = ?,
  status = ?, last_heartbeat_at = ?, last_online_at = ?,
  enrollment_token_hash = NULL, enrollment_expires_at = NULL,
  updated_at = NOW()
WHERE id = ? AND enrollment_token_hash IS NOT NULL`,
		n.AgentID, n.AgentSecret, n.NodeKey, n.Hostname, n.Platform, n.PlatformVersion, n.Arch, n.AgentVersion,
		n.Status, n.LastHeartbeatAt, n.LastOnlineAt, n.ID)
	if err != nil {
		return fmt.Errorf("complete enrollment: %w", err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("complete enrollment rows: %w", err)
	}
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// MarkTimedOutNodes marks nodes unstable or offline based on heartbeat thresholds.
func (r *MySQLNodeRepository) MarkTimedOutNodes(ctx context.Context, unstableBefore, offlineBefore time.Time) error {
	if _, err := r.db.ExecContext(ctx, `UPDATE nodes SET status = 'unstable', updated_at = NOW() WHERE last_heartbeat_at < ? AND last_heartbeat_at >= ? AND status NOT IN ('offline', 'pending')`, unstableBefore, offlineBefore); err != nil {
		return fmt.Errorf("mark unstable nodes: %w", err)
	}
	if _, err := r.db.ExecContext(ctx, `UPDATE nodes SET status = 'offline', updated_at = NOW() WHERE last_heartbeat_at < ? AND status NOT IN ('offline', 'pending')`, offlineBefore); err != nil {
		return fmt.Errorf("mark offline nodes: %w", err)
	}
	return nil
}

func (r *MySQLNodeRepository) getOne(ctx context.Context, query string, args ...any) (*model.Node, error) {
	item, err := scanNode(r.db.QueryRowContext(ctx, query, args...))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return item, nil
}

type nodeScanner interface{ Scan(dest ...any) error }

func scanNode(scanner nodeScanner) (*model.Node, error) {
	var item model.Node
	var enrollHash sql.NullString
	var enrollExp sql.NullTime
	if err := scanner.Scan(&item.ID, &item.AgentID, &item.AgentSecret, &item.NodeKey, &item.Name, &item.Hostname, &item.Platform, &item.PlatformVersion, &item.Arch, &item.AgentVersion, &item.Status, &item.LastHeartbeatAt, &item.LastOnlineAt, &item.CreatedAt, &item.UpdatedAt, &enrollHash, &enrollExp); err != nil {
		return nil, fmt.Errorf("scan node: %w", err)
	}
	if enrollHash.Valid {
		item.EnrollmentTokenHash = enrollHash.String
	}
	if enrollExp.Valid {
		t := enrollExp.Time.UTC()
		item.EnrollmentExpiresAt = &t
	}
	return &item, nil
}
