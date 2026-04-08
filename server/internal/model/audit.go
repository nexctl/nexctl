package model

import "time"

// AuditLog stores an auditable control-plane action.
type AuditLog struct {
	ID           int64     `db:"id" json:"id"`
	ActorType    string    `db:"actor_type" json:"actor_type"`
	ActorID      string    `db:"actor_id" json:"actor_id"`
	ActorName    string    `db:"actor_name" json:"actor_name"`
	Action       string    `db:"action" json:"action"`
	ResourceType string    `db:"resource_type" json:"resource_type"`
	ResourceID   string    `db:"resource_id" json:"resource_id"`
	Detail       string    `db:"detail" json:"detail"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
}
