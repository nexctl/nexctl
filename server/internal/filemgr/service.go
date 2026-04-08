package filemgr

import "context"

// Service is the file management placeholder for future phases.
type Service struct{}

// NewService creates a file management service.
func NewService() *Service {
	return &Service{}
}

// Ping is a minimal callable method for dependency wiring and testing.
func (s *Service) Ping(context.Context) error {
	return nil
}

// List returns the reserved file list contract for future implementation.
func (s *Service) List(context.Context) (*ListResponse, error) {
	return &ListResponse{Items: []ListItem{}}, nil
}
