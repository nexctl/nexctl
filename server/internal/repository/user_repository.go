package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/nexctl/nexctl/server/internal/model"
)

// UserRepository defines user data access.
type UserRepository interface {
	FindByUsername(ctx context.Context, username string) (*model.User, string, error)
}

// MySQLUserRepository is the MySQL user repository implementation.
type MySQLUserRepository struct {
	db *sql.DB
}

// NewUserRepository creates a user repository.
func NewUserRepository(db *sql.DB) *MySQLUserRepository {
	return &MySQLUserRepository{db: db}
}

// FindByUsername returns a user with its primary role code.
func (r *MySQLUserRepository) FindByUsername(ctx context.Context, username string) (*model.User, string, error) {
	const query = `
SELECT u.id, u.username, u.password_hash, u.display_name, u.status, u.created_at, u.updated_at, COALESCE(ro.code, '')
FROM users u
LEFT JOIN user_roles ur ON ur.user_id = u.id
LEFT JOIN roles ro ON ro.id = ur.role_id
WHERE u.username = ?
LIMIT 1`

	var user model.User
	var roleCode string
	err := r.db.QueryRowContext(ctx, query, username).Scan(
		&user.ID, &user.Username, &user.PasswordHash, &user.DisplayName, &user.Status, &user.CreatedAt, &user.UpdatedAt, &roleCode,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, "", nil
		}
		return nil, "", fmt.Errorf("find user by username: %w", err)
	}
	return &user, roleCode, nil
}
