package model

import "time"

// NodeRuntimeState stores the latest runtime snapshot for a node.
type NodeRuntimeState struct {
	ID            int64     `db:"id" json:"id"`
	NodeID        int64     `db:"node_id" json:"node_id"`
	CPUPercent    float64   `db:"cpu_percent" json:"cpu_percent"`
	MemoryPercent float64   `db:"memory_percent" json:"memory_percent"`
	DiskPercent   float64   `db:"disk_percent" json:"disk_percent"`
	NetworkRxBps  uint64    `db:"network_rx_bps" json:"network_rx_bps"`
	NetworkTxBps  uint64    `db:"network_tx_bps" json:"network_tx_bps"`
	Load1         float64   `db:"load_1" json:"load_1"`
	Load5         float64   `db:"load_5" json:"load_5"`
	Load15        float64   `db:"load_15" json:"load_15"`
	UptimeSeconds uint64    `db:"uptime_seconds" json:"uptime_seconds"`
	ProcessCount  uint32    `db:"process_count" json:"process_count"`
	ReportedAt    time.Time `db:"reported_at" json:"reported_at"`
	UpdatedAt     time.Time `db:"updated_at" json:"updated_at"`
}
