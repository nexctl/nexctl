package model

import "time"

// File stores metadata for managed files.
type File struct {
	ID          int64     `db:"id" json:"id"`
	FileName    string    `db:"file_name" json:"file_name"`
	StoragePath string    `db:"storage_path" json:"storage_path"`
	SizeBytes   int64     `db:"size_bytes" json:"size_bytes"`
	SHA256      string    `db:"sha256" json:"sha256"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
}
