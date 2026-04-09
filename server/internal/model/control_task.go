package model

import (
	"database/sql"
	"time"
)

// ControlTask 控制面任务执行实例（手动或计划触发）。
type ControlTask struct {
	ID           int64          `db:"id" json:"id"`
	ScheduleID   sql.NullInt64  `db:"schedule_id" json:"schedule_id,omitempty"`
	TaskType     string         `db:"task_type" json:"task_type"`
	ScopeType    string         `db:"scope_type" json:"scope_type"`
	ScopeValue   string         `db:"scope_value" json:"scope_value"`
	Status       string         `db:"status" json:"status"`
	Progress     int            `db:"progress" json:"progress"`
	OperatorID   int64          `db:"operator_id" json:"operator_id"`
	OperatorName string         `db:"operator_name" json:"operator_name"`
	Payload      sql.NullString `db:"payload" json:"payload,omitempty"`
	Detail       string         `db:"detail" json:"detail"`
	Output       string         `db:"output" json:"output"`
	CreatedAt    time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time      `db:"updated_at" json:"updated_at"`
	FinishedAt   sql.NullTime   `db:"finished_at" json:"finished_at,omitempty"`
}
