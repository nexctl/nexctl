package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/nexctl/nexctl/server/internal/model"
)

// InstallTokenRepository defines install token data access.
type InstallTokenRepository interface {
	FindByToken(ctx context.Context, token string) (*model.InstallToken, error)
	IncrementUsedCount(ctx context.Context, id int64) error
}

// MySQLInstallTokenRepository is the MySQL install token repository.
type MySQLInstallTokenRepository struct {
	db *sql.DB
}

// NewInstallTokenRepository creates an install token repository.
func NewInstallTokenRepository(db *sql.DB) *MySQLInstallTokenRepository {
	return &MySQLInstallTokenRepository{db: db}
}

// FindByToken finds an install token record by token string.
func (r *MySQLInstallTokenRepository) FindByToken(ctx context.Context, token string) (*model.InstallToken, error) {
	const query = `SELECT id, token, description, max_uses, used_count, expires_at, created_at FROM install_tokens WHERE token = ? LIMIT 1`
	var item model.InstallToken
	var expiresAt sql.NullTime
	err := r.db.QueryRowContext(ctx, query, token).Scan(&item.ID, &item.Token, &item.Description, &item.MaxUses, &item.UsedCount, &expiresAt, &item.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("find install token: %w", err)
	}
	if expiresAt.Valid {
		t := expiresAt.Time.UTC()
		item.ExpiresAt = &t
	}
	return &item, nil
}

// IncrementUsedCount increments the used count for an install token.
func (r *MySQLInstallTokenRepository) IncrementUsedCount(ctx context.Context, id int64) error {
	if _, err := r.db.ExecContext(ctx, `UPDATE install_tokens SET used_count = used_count + 1 WHERE id = ?`, id); err != nil {
		return fmt.Errorf("increment install token use count: %w", err)
	}
	return nil
}

// IsUsable checks whether the install token is still valid.
func IsUsable(token *model.InstallToken, now time.Time) bool {
	if token == nil {
		return false
	}
	if token.MaxUses > 0 && token.UsedCount >= token.MaxUses {
		return false
	}
	if token.ExpiresAt != nil && token.ExpiresAt.Before(now) {
		return false
	}
	return true
}
