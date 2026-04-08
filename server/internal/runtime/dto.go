package runtime

import "time"

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
