package health

import "context"

// DatabasePinger defines the interface for checking database connectivity.
type DatabasePinger interface {
	Ping(ctx context.Context) error
}

// Service aggregates application and database health status.
type Service struct {
	AppName string
	AppEnv  string
	DB      DatabasePinger
}

// Status represents the health check result data.
type Status struct {
	Service  string `json:"service"`
	Env      string `json:"env"`
	App      string `json:"app"`
	Database string `json:"database"`
}

// NewService creates a new health check service.
func NewService(appName string, appEnv string, db DatabasePinger) *Service {
	return &Service{
		AppName: appName,
		AppEnv:  appEnv,
		DB:      db,
	}
}

// Check performs health checks and returns the status along with a boolean
// indicating whether the system is fully healthy.
func (s *Service) Check(ctx context.Context) (Status, bool) {
	status := Status{
		Service: s.AppName,
		Env:     s.AppEnv,
		App:     "ok",
	}

	if s.DB == nil {
		status.Database = "unavailable"
		return status, false
	}

	if err := s.DB.Ping(ctx); err != nil {
		status.Database = "unavailable"
		return status, false
	}

	status.Database = "ok"
	return status, true
}
