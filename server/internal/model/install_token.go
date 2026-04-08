package model

import "time"

// InstallToken authorizes new agent registration.
type InstallToken struct {
	ID          int64      `db:"id" json:"id"`
	Token       string     `db:"token" json:"token"`
	Description string     `db:"description" json:"description"`
	MaxUses     int        `db:"max_uses" json:"max_uses"`
	UsedCount   int        `db:"used_count" json:"used_count"`
	ExpiresAt   *time.Time `db:"expires_at" json:"expires_at"`
	CreatedAt   time.Time  `db:"created_at" json:"created_at"`
}
