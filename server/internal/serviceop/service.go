package serviceop

import "context"

// Service is the service-operation placeholder for future phases.
type Service struct{}

// NewService creates a service-operation service.
func NewService() *Service {
	return &Service{}
}

// Ping is a minimal callable method for dependency wiring and testing.
func (s *Service) Ping(context.Context) error {
	return nil
}
