package alert

import "context"

// Service is the alert placeholder for future phases.
type Service struct{}

// NewService creates an alert service.
func NewService() *Service {
	return &Service{}
}

// Ping is a minimal callable method for dependency wiring and testing.
func (s *Service) Ping(context.Context) error {
	return nil
}

// ListRules returns the reserved alert-rule list contract for future implementation.
func (s *Service) ListRules(context.Context) (*RuleListResponse, error) {
	return &RuleListResponse{Items: []RuleItem{}}, nil
}

// ListEvents returns the reserved alert-event list contract for future implementation.
func (s *Service) ListEvents(context.Context) (*EventListResponse, error) {
	return &EventListResponse{Items: []EventItem{}}, nil
}
