package main

// EventStoreInterface defines the interface for event stores
type EventStoreInterface interface {
	CreateSession(sessionID, name string) error
	AppendEvents(sessionID string, round int, events []Event) error
	SaveSnapshot(sessionID string, round int, state State) error
	GetEvents(sessionID string, fromRound int) ([]Event, error)
	GetLatestSnapshot(sessionID string) (*State, error)
	GetSnapshotAtRound(sessionID string, round int) (*State, error)
	UpdateSessionStatus(sessionID, status string) error
	Close() error
}
