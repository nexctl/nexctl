package upgrade

import "context"

// Service is the agent upgrade placeholder for future phases.
type Service struct{}

// NewService creates an upgrade service.
func NewService() *Service {
	return &Service{}
}

// Ping is a minimal callable method for dependency wiring and testing.
func (s *Service) Ping(context.Context) error {
	return nil
}

// ListReleases returns the reserved release list contract for future implementation.
func (s *Service) ListReleases(context.Context) (*ReleaseListResponse, error) {
	return &ReleaseListResponse{Items: []ReleaseItem{}}, nil
}
