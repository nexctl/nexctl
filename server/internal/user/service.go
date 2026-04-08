package user

import (
	"context"

	"github.com/nexctl/nexctl/server/internal/model"
	"github.com/nexctl/nexctl/server/internal/repository"
)

// Service provides user-related business operations.
type Service struct {
	users repository.UserRepository
}

// NewService creates a user service.
func NewService(users repository.UserRepository) *Service {
	return &Service{users: users}
}

// FindByUsername returns a user by username.
func (s *Service) FindByUsername(ctx context.Context, username string) (*model.User, string, error) {
	return s.users.FindByUsername(ctx, username)
}
