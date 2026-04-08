package model

import "time"

const (
	// NodeStatusPending means the node was pre-created in console and awaits enrollment_token from agent.
	NodeStatusPending = "pending"
	// NodeStatusOnline means the node session is healthy.
	NodeStatusOnline = "online"
	// NodeStatusUnstable means the node heartbeat is late.
	NodeStatusUnstable = "unstable"
	// NodeStatusOffline means the node heartbeat timed out.
	NodeStatusOffline = "offline"
)

// Node stores the managed node identity and access secret.
type Node struct {
	ID              int64     `db:"id" json:"id"`
	AgentID         string    `db:"agent_id" json:"agent_id"`
	AgentSecret     string    `db:"agent_secret" json:"-"`
	NodeKey         string    `db:"node_key" json:"node_key"`
	Name            string    `db:"name" json:"name"`
	Hostname        string    `db:"hostname" json:"hostname"`
	Platform        string    `db:"platform" json:"platform"`
	PlatformVersion string    `db:"platform_version" json:"platform_version"`
	Arch            string    `db:"arch" json:"arch"`
	AgentVersion    string    `db:"agent_version" json:"agent_version"`
	Status          string    `db:"status" json:"status"`
	LastHeartbeatAt time.Time `db:"last_heartbeat_at" json:"last_heartbeat_at"`
	LastOnlineAt    time.Time `db:"last_online_at" json:"last_online_at"`
	CreatedAt       time.Time `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time `db:"updated_at" json:"updated_at"`
	// EnrollmentTokenHash is SHA256(enrollment_token) hex; empty when not awaiting enrollment.
	EnrollmentTokenHash string `db:"enrollment_token_hash" json:"-"`
	EnrollmentExpiresAt *time.Time `db:"enrollment_expires_at" json:"enrollment_expires_at,omitempty"`
}

// NodeLabel stores a label assignment for a node.
type NodeLabel struct {
	ID         int64     `db:"id" json:"id"`
	NodeID     int64     `db:"node_id" json:"node_id"`
	LabelKey   string    `db:"label_key" json:"label_key"`
	LabelValue string    `db:"label_value" json:"label_value"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
}
