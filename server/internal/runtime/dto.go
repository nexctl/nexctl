package runtime

import (
	"strings"
	"time"
)

// UpdateStateRequest is the current runtime state update request body.
type UpdateStateRequest struct {
	CPUPercent    float64 `json:"cpu_percent"`
	MemoryPercent float64 `json:"memory_percent"`
	DiskPercent   float64 `json:"disk_percent"`
	NetworkRxBps  uint64  `json:"network_rx_bps"`
	NetworkTxBps  uint64  `json:"network_tx_bps"`
	Load1         float64 `json:"load_1"`
	Load5         float64 `json:"load_5"`
	Load15        float64 `json:"load_15"`
	UptimeSeconds uint64  `json:"uptime_seconds"`
	ProcessCount  uint32  `json:"process_count"`
	Timestamp     string  `json:"timestamp,omitempty"`
	// Agent 上报的机器信息与版本（与 WebSocket runtime_state 一致）
	Hostname        string `json:"hostname,omitempty"`
	Platform        string `json:"platform,omitempty"`
	PlatformVersion string `json:"platform_version,omitempty"`
	Arch            string `json:"arch,omitempty"`
	AgentVersion    string `json:"agent_version,omitempty"`
}

// HasAgentMeta 为 true 时写入 nodes 表的 hostname/platform/arch/agent_version 等展示字段。
func (r UpdateStateRequest) HasAgentMeta() bool {
	return strings.TrimSpace(r.Platform) != "" || strings.TrimSpace(r.Arch) != "" || strings.TrimSpace(r.AgentVersion) != "" ||
		strings.TrimSpace(r.Hostname) != "" || strings.TrimSpace(r.PlatformVersion) != ""
}

// ReportedAt parses the optional payload timestamp or falls back to now.
func (r UpdateStateRequest) ReportedAt() time.Time {
	if r.Timestamp == "" {
		return time.Now().UTC()
	}
	if parsed, err := time.Parse(time.RFC3339, r.Timestamp); err == nil {
		return parsed.UTC()
	}
	return time.Now().UTC()
}
