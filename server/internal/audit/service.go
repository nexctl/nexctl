package audit

import (
	"context"

	"github.com/nexctl/nexctl/server/internal/model"
	"github.com/nexctl/nexctl/server/internal/repository"
	"go.uber.org/zap"
)

// RecordInput is the input for creating an audit log entry.
type RecordInput struct {
	ActorType    string
	ActorID      string
	ActorName    string
	Action       string
	ResourceType string
	ResourceID   string
	Detail       string
}

// Service records auditable actions.
type Service struct {
	repo   repository.AuditRepository
	logger *zap.Logger
}

// NewService creates an audit service.
func NewService(repo repository.AuditRepository, logger *zap.Logger) *Service {
	return &Service{repo: repo, logger: logger}
}

// Record stores an audit entry.
func (s *Service) Record(ctx context.Context, input RecordInput) error {
	entry := &model.AuditLog{
		ActorType:    input.ActorType,
		ActorID:      input.ActorID,
		ActorName:    input.ActorName,
		Action:       input.Action,
		ResourceType: input.ResourceType,
		ResourceID:   input.ResourceID,
		Detail:       input.Detail,
	}
	if err := s.repo.Create(ctx, entry); err != nil {
		s.logger.Warn("write audit log", zap.Error(err), zap.String("action", input.Action))
		return err
	}
	return nil
}

// List returns the reserved audit-log list contract for future implementation.
func (s *Service) List(context.Context) (*ListResponse, error) {
	return &ListResponse{Items: []LogItem{}}, nil
}
