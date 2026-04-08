package node

import "github.com/nexctl/nexctl/server/internal/model"

// RegisterRequest is the node registration request body.
type RegisterRequest struct {
	InstallToken     string `json:"install_token,omitempty"`
	EnrollmentToken  string `json:"enrollment_token,omitempty"`
	NodeKey          string `json:"node_key"`
	Name             string `json:"name"`
	Hostname        string `json:"hostname"`
	Platform        string `json:"platform"`
	PlatformVersion string `json:"platform_version"`
	Arch            string `json:"arch"`
	PrivateIP       string `json:"private_ip,omitempty"`
	PublicIP        string `json:"public_ip,omitempty"`
	AgentVersion    string `json:"agent_version"`
}

// CreatePendingNodeRequest is the console "add node" body.
type CreatePendingNodeRequest struct {
	Name            string `json:"name"`
	ExpiresInHours  int    `json:"expires_in_hours"`
}

// CreatePendingNodeResponse returns the one-time enrollment token for agent bootstrap.
type CreatePendingNodeResponse struct {
	ID                  int64  `json:"id"`
	Name                string `json:"name"`
	Status              string `json:"status"`
	EnrollmentToken     string `json:"enrollment_token"`
	EnrollmentExpiresAt string `json:"enrollment_expires_at,omitempty"`
}

// RegisterResponse is the node registration response body.
type RegisterResponse struct {
	NodeID      int64  `json:"node_id"`
	AgentID     string `json:"agent_id"`
	AgentSecret string `json:"agent_secret"`
	WSURL       string `json:"ws_url"`
}

// ListItem is the node list response item.
type ListItem struct {
	ID              int64                   `json:"id"`
	Name            string                  `json:"name"`
	Status          string                  `json:"status"`
	Hostname        string                  `json:"hostname"`
	Platform        string                  `json:"platform"`
	Arch            string                  `json:"arch"`
	AgentVersion    string                  `json:"agent_version"`
	LastHeartbeatAt string                  `json:"last_heartbeat_at"`
	Labels          []string                `json:"labels"`
	RuntimeState    *model.NodeRuntimeState `json:"runtime_state,omitempty"`
}

// ListResponse is the node list response body.
type ListResponse struct {
	Items []*ListItem `json:"items"`
}

// DetailResponse is the node detail response body.
type DetailResponse struct {
	ID               int64                   `json:"id"`
	Name             string                  `json:"name"`
	Status           string                  `json:"status"`
	Hostname         string                  `json:"hostname"`
	Platform         string                  `json:"platform"`
	PlatformVersion  string                  `json:"platform_version"`
	Arch             string                  `json:"arch"`
	AgentVersion     string                  `json:"agent_version"`
	NodeKey          string                  `json:"node_key"`
	LastHeartbeatAt  string                  `json:"last_heartbeat_at"`
	LastOnlineAt     string                  `json:"last_online_at"`
	Labels           []string                `json:"labels"`
	RuntimeState     *model.NodeRuntimeState `json:"runtime_state,omitempty"`
	Services         []ServiceItem           `json:"services"`
	RecentTasks      []TaskItem              `json:"recent_tasks"`
	Alerts           []AlertItem             `json:"alerts"`
	ShortTermMetrics []MetricPoint           `json:"short_term_metrics"`
}

// ServiceItem is the node detail service row.
type ServiceItem struct {
	Name        string `json:"name"`
	Status      string `json:"status"`
	StartupType string `json:"startup_type"`
}

// TaskItem is the node detail recent task row.
type TaskItem struct {
	ID        int64  `json:"id"`
	Type      string `json:"type"`
	Status    string `json:"status"`
	Target    string `json:"target"`
	CreatedAt string `json:"created_at"`
}

// AlertItem is the node detail recent alert row.
type AlertItem struct {
	ID        int64  `json:"id"`
	Severity  string `json:"severity"`
	Summary   string `json:"summary"`
	CreatedAt string `json:"created_at"`
}

// MetricPoint is a short-term runtime point.
type MetricPoint struct {
	Time   string  `json:"time"`
	CPU    float64 `json:"cpu"`
	Memory float64 `json:"memory"`
	Disk   float64 `json:"disk"`
}
