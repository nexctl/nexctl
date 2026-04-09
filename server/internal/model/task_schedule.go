package model

import (
	"database/sql"
	"time"
)

// TaskSchedule 计划任务（CRON 周期执行）。
type TaskSchedule struct {
	ID           int64        `db:"id" json:"id"`
	Name         string       `db:"name" json:"name"`
	CronExpr     string       `db:"cron_expr" json:"cron_expr"`
	TaskType     string       `db:"task_type" json:"task_type"`
	ScopeType    string       `db:"scope_type" json:"scope_type"`
	ScopeValue   string       `db:"scope_value" json:"scope_value"`
	Detail       string       `db:"detail" json:"detail"`
	Enabled      bool         `db:"enabled" json:"enabled"`
	OperatorID   int64        `db:"operator_id" json:"operator_id"`
	OperatorName string       `db:"operator_name" json:"operator_name"`
	NextRunAt    time.Time    `db:"next_run_at" json:"next_run_at"`
	LastRunAt    sql.NullTime `db:"last_run_at" json:"last_run_at,omitempty"`
	CreatedAt    time.Time    `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time    `db:"updated_at" json:"updated_at"`
}
