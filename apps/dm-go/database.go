package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// EventStore handles persistence of game events and snapshots
type EventStore struct {
	db *sql.DB
}

// Ensure EventStore implements EventStoreInterface
var _ EventStoreInterface = (*EventStore)(nil)

// NewEventStore creates a new event store
func NewEventStore(dbPath string) (*EventStore, error) {
	if dbPath == "" {
		dbPath = ":memory:"
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	store := &EventStore{db: db}

	if err := store.initSchema(); err != nil {
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return store, nil
}

func (es *EventStore) initSchema() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS events (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			session_id TEXT NOT NULL,
			round INTEGER NOT NULL,
			event_data TEXT NOT NULL,
			timestamp INTEGER NOT NULL DEFAULT (unixepoch())
		)`,
		`CREATE TABLE IF NOT EXISTS snapshots (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			session_id TEXT NOT NULL,
			round INTEGER NOT NULL,
			state_data TEXT NOT NULL,
			timestamp INTEGER NOT NULL DEFAULT (unixepoch())
		)`,
		`CREATE TABLE IF NOT EXISTS sessions (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			status TEXT NOT NULL DEFAULT 'active',
			created_at INTEGER NOT NULL DEFAULT (unixepoch()),
			updated_at INTEGER NOT NULL DEFAULT (unixepoch())
		)`,
		`CREATE INDEX IF NOT EXISTS idx_events_session_round ON events(session_id, round)`,
		`CREATE INDEX IF NOT EXISTS idx_snapshots_session_round ON snapshots(session_id, round)`,
	}

	for _, query := range queries {
		if _, err := es.db.Exec(query); err != nil {
			return fmt.Errorf("failed to execute query %q: %w", query, err)
		}
	}

	return nil
}

// CreateSession creates a new game session
func (es *EventStore) CreateSession(sessionID, name string) error {
	_, err := es.db.Exec(
		"INSERT INTO sessions (id, name, status) VALUES (?, ?, ?)",
		sessionID, name, "active",
	)
	return err
}

// AppendEvents appends events to a session
func (es *EventStore) AppendEvents(sessionID string, round int, events []Event) error {
	if len(events) == 0 {
		return nil
	}

	tx, err := es.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare("INSERT INTO events (session_id, round, event_data) VALUES (?, ?, ?)")
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, event := range events {
		eventData, err := json.Marshal(event)
		if err != nil {
			return fmt.Errorf("failed to marshal event: %w", err)
		}

		_, err = stmt.Exec(sessionID, round, string(eventData))
		if err != nil {
			return fmt.Errorf("failed to insert event: %w", err)
		}
	}

	return tx.Commit()
}

// SaveSnapshot saves a game state snapshot
func (es *EventStore) SaveSnapshot(sessionID string, round int, state State) error {
	stateData, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	_, err = es.db.Exec(
		"INSERT INTO snapshots (session_id, round, state_data) VALUES (?, ?, ?)",
		sessionID, round, string(stateData),
	)
	return err
}

// GetEvents retrieves events for a session from a given round
func (es *EventStore) GetEvents(sessionID string, fromRound int) ([]Event, error) {
	rows, err := es.db.Query(
		"SELECT event_data FROM events WHERE session_id = ? AND round >= ? ORDER BY round, id",
		sessionID, fromRound,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query events: %w", err)
	}
	defer rows.Close()

	var events []Event
	for rows.Next() {
		var eventData string
		if err := rows.Scan(&eventData); err != nil {
			return nil, fmt.Errorf("failed to scan event: %w", err)
		}

		var event Event
		if err := json.Unmarshal([]byte(eventData), &event); err != nil {
			return nil, fmt.Errorf("failed to unmarshal event: %w", err)
		}

		events = append(events, event)
	}

	return events, rows.Err()
}

// GetLatestSnapshot retrieves the most recent snapshot for a session
func (es *EventStore) GetLatestSnapshot(sessionID string) (*State, error) {
	var stateData string
	err := es.db.QueryRow(
		"SELECT state_data FROM snapshots WHERE session_id = ? ORDER BY round DESC LIMIT 1",
		sessionID,
	).Scan(&stateData)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get latest snapshot: %w", err)
	}

	var state State
	if err := json.Unmarshal([]byte(stateData), &state); err != nil {
		return nil, fmt.Errorf("failed to unmarshal state: %w", err)
	}

	return &state, nil
}

// GetSnapshotAtRound retrieves a snapshot at a specific round
func (es *EventStore) GetSnapshotAtRound(sessionID string, round int) (*State, error) {
	var stateData string
	err := es.db.QueryRow(
		"SELECT state_data FROM snapshots WHERE session_id = ? AND round <= ? ORDER BY round DESC LIMIT 1",
		sessionID, round,
	).Scan(&stateData)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get snapshot at round: %w", err)
	}

	var state State
	if err := json.Unmarshal([]byte(stateData), &state); err != nil {
		return nil, fmt.Errorf("failed to unmarshal state: %w", err)
	}

	return &state, nil
}

// UpdateSessionStatus updates the status of a session
func (es *EventStore) UpdateSessionStatus(sessionID, status string) error {
	_, err := es.db.Exec(
		"UPDATE sessions SET status = ?, updated_at = ? WHERE id = ?",
		status, time.Now().Unix(), sessionID,
	)
	return err
}

// Close closes the database connection
func (es *EventStore) Close() error {
	return es.db.Close()
}