package terminal

import "context"

// Service is the web-terminal placeholder for future phases.
type Service struct{}

// NewService creates a terminal service.
func NewService() *Service {
	return &Service{}
}

// Ping is a minimal callable method for dependency wiring and testing.
func (s *Service) Ping(context.Context) error {
	return nil
}
