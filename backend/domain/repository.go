package domain

// Repository defines the interface for data persistence
// This interface is defined in the domain layer, and implemented in the infrastructure layer
type Repository interface {
	GetTodaySession(date string) *WorkSession
	GetSessionsByMonth(yearMonth string) ([]*WorkSession, error)
	SaveSession(session *WorkSession) error
	GetConfig() (*WorkConfig, error)
	SaveConfig(config *WorkConfig) error
	Close() error
}
