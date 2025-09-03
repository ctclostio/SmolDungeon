package main

import (
	"encoding/json"
	"fmt"
	"time"
)

// MemoryEventStore is an in-memory implementation for testing
type MemoryEventStore struct {
	events    []Event
	snapshots []Snapshot
	sessions  []Session
}

// Ensure MemoryEventStore implements EventStoreInterface
var _ EventStoreInterface = (*MemoryEventStore)(nil)

// Snapshot represents a game state snapshot
type Snapshot struct {
	ID        int    `json:"id"`
	SessionID string `json:"sessionId"`
	Round     int    `json:"round"`
	StateData string `json:"stateData"`
	Timestamp int64  `json:"timestamp"`
}

// Session represents a game session
type Session struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Status    string `json:"status"`
	CreatedAt int64  `json:"createdAt"`
	UpdatedAt int64  `json:"updatedAt"`
}

// NewMemoryEventStore creates a new in-memory event store
func NewMemoryEventStore() *MemoryEventStore {
	return &MemoryEventStore{
		events:    []Event{},
		snapshots: []Snapshot{},
		sessions:  []Session{},
	}
}

// CreateSession creates a new game session
func (mes *MemoryEventStore) CreateSession(sessionID, name string) error {
	session := Session{
		ID:        sessionID,
		Name:      name,
		Status:    "active",
		CreatedAt: time.Now().Unix(),
		UpdatedAt: time.Now().Unix(),
	}
	mes.sessions = append(mes.sessions, session)
	return nil
}

// AppendEvents appends events to a session
func (mes *MemoryEventStore) AppendEvents(sessionID string, round int, events []Event) error {
	for _, event := range events {
		event.ID = fmt.Sprintf("%s-%d-%d", sessionID, round, len(mes.events))
		mes.events = append(mes.events, event)
	}
	return nil
}

// SaveSnapshot saves a game state snapshot
func (mes *MemoryEventStore) SaveSnapshot(sessionID string, round int, state State) error {
	stateData, err := json.Marshal(state)
	if err != nil {
		return err
	}

	snapshot := Snapshot{
		ID:        len(mes.snapshots),
		SessionID: sessionID,
		Round:     round,
		StateData: string(stateData),
		Timestamp: time.Now().Unix(),
	}
	mes.snapshots = append(mes.snapshots, snapshot)
	return nil
}

// GetEvents retrieves events for a session from a given round
func (mes *MemoryEventStore) GetEvents(sessionID string, fromRound int) ([]Event, error) {
	var result []Event
	for _, event := range mes.events {
		if event.ID != "" && len(event.ID) > len(sessionID) && event.ID[:len(sessionID)] == sessionID {
			result = append(result, event)
		}
	}
	return result, nil
}

// GetLatestSnapshot retrieves the most recent snapshot for a session
func (mes *MemoryEventStore) GetLatestSnapshot(sessionID string) (*State, error) {
	var latest *Snapshot
	for i := range mes.snapshots {
		if mes.snapshots[i].SessionID == sessionID {
			if latest == nil || mes.snapshots[i].Round > latest.Round {
				latest = &mes.snapshots[i]
			}
		}
	}

	if latest == nil {
		return nil, nil
	}

	var state State
	if err := json.Unmarshal([]byte(latest.StateData), &state); err != nil {
		return nil, err
	}

	return &state, nil
}

// GetSnapshotAtRound retrieves a snapshot at a specific round
func (mes *MemoryEventStore) GetSnapshotAtRound(sessionID string, round int) (*State, error) {
	for _, snapshot := range mes.snapshots {
		if snapshot.SessionID == sessionID && snapshot.Round <= round {
			var state State
			if err := json.Unmarshal([]byte(snapshot.StateData), &state); err != nil {
				return nil, err
			}
			return &state, nil
		}
	}
	return nil, nil
}

// UpdateSessionStatus updates the status of a session
func (mes *MemoryEventStore) UpdateSessionStatus(sessionID, status string) error {
	for i := range mes.sessions {
		if mes.sessions[i].ID == sessionID {
			mes.sessions[i].Status = status
			mes.sessions[i].UpdatedAt = time.Now().Unix()
			return nil
		}
	}
	return fmt.Errorf("session not found: %s", sessionID)
}

// Close is a no-op for memory store
func (mes *MemoryEventStore) Close() error {
	return nil
}