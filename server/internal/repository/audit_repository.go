package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/nexctl/nexctl/server/internal/model"
)

// AuditRepository defines audit log data access.
type AuditRepository interface {
	Create(ctx context.Context, log *model.AuditLog) error
}

// MySQLAuditRepository is the MySQL audit repository.
type MySQLAuditRepository struct {
	db *sql.DB
}

// NewAuditRepository creates an audit repository.
func NewAuditRepository(db *sql.DB) *MySQLAuditRepository {
	return &MySQLAuditRepository{db: db}
}

// Create inserts a new audit log entry.
func (r *MySQLAuditRepository) Create(ctx context.Context, log *model.AuditLog) error {
	result, err := r.db.ExecContext(ctx, `INSERT INTO audit_logs (actor_type, actor_id, actor_name, action, resource_type, resource_id, detail, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, NOW())`, log.ActorType, log.ActorID, log.ActorName, log.Action, log.ResourceType, log.ResourceID, log.Detail)
	if err != nil {
		return fmt.Errorf("create audit log: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("get audit log id: %w", err)
	}
	log.ID = id
	return nil
}
