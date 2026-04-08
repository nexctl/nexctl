package model

import "time"

// AlertRule stores a threshold or event rule.
type AlertRule struct {
	ID         int64     `db:"id" json:"id"`
	Name       string    `db:"name" json:"name"`
	RuleType   string    `db:"rule_type" json:"rule_type"`
	TargetExpr string    `db:"target_expr" json:"target_expr"`
	Enabled    bool      `db:"enabled" json:"enabled"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
}

// AlertEvent stores a rule evaluation result.
type AlertEvent struct {
	ID          int64     `db:"id" json:"id"`
	RuleID      int64     `db:"rule_id" json:"rule_id"`
	NodeID      int64     `db:"node_id" json:"node_id"`
	Severity    string    `db:"severity" json:"severity"`
	Status      string    `db:"status" json:"status"`
	Summary     string    `db:"summary" json:"summary"`
	TriggeredAt time.Time `db:"triggered_at" json:"triggered_at"`
}
