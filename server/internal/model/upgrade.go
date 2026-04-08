package model

import "time"

// AgentRelease stores an agent release artifact.
type AgentRelease struct {
	ID         int64     `db:"id" json:"id"`
	Version    string    `db:"version" json:"version"`
	Channel    string    `db:"channel" json:"channel"`
	PackageURL string    `db:"package_url" json:"package_url"`
	Checksum   string    `db:"checksum" json:"checksum"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
}
