package model

import "time"

// Task stores a control-plane task definition.
type Task struct {
	ID         int64     `db:"id" json:"id"`
	TaskType   string    `db:"task_type" json:"task_type"`
	ScopeType  string    `db:"scope_type" json:"scope_type"`
	ScopeValue string    `db:"scope_value" json:"scope_value"`
	Status     string    `db:"status" json:"status"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
	UpdatedAt  time.Time `db:"updated_at" json:"updated_at"`
}

// TaskExecution stores a task execution result for a node.
type TaskExecution struct {
	ID         int64     `db:"id" json:"id"`
	TaskID     int64     `db:"task_id" json:"task_id"`
	NodeID     int64     `db:"node_id" json:"node_id"`
	Status     string    `db:"status" json:"status"`
	Output     string    `db:"output" json:"output"`
	StartedAt  time.Time `db:"started_at" json:"started_at"`
	FinishedAt time.Time `db:"finished_at" json:"finished_at"`
}
