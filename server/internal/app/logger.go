package app

import "go.uber.org/zap"

// newLogger creates a zap logger for the configured environment.
func newLogger(env string) (*zap.Logger, error) {
	if env == "production" {
		return zap.NewProduction()
	}
	return zap.NewDevelopment()
}
