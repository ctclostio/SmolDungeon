package main

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

// SessionService handles session management business logic
type SessionService struct {
	eventStore   EventStoreInterface
	stateManager *StateManager
	scenarioDir  string
}

// NewSessionService creates a new session service
func NewSessionService(
	eventStore EventStoreInterface,
	stateManager *StateManager,
	scenarioDir string,
) *SessionService {
	return &SessionService{
		eventStore:   eventStore,
		stateManager: stateManager,
		scenarioDir:  scenarioDir,
	}
}

// CreateSession creates a new game session from a scenario
func (s *SessionService) CreateSession(scenarioName string) (string, State, error) {
	// Default scenario if not provided
	if scenarioName == "" {
		scenarioName = "goblin-ambush"
	}

	// Validate scenario name to prevent path traversal
	if strings.Contains(scenarioName, "..") || strings.Contains(scenarioName, "/") || strings.Contains(scenarioName, "\\") {
		return "", State{}, fmt.Errorf("invalid scenario name")
	}

	// Load scenario
	scenarioPath := filepath.Join(s.scenarioDir, scenarioName+".yaml")
	scenario, err := LoadScenario(scenarioPath)
	if err != nil {
		return "", State{}, fmt.Errorf("scenario not found: %w", err)
	}

	// Create initial game state
	sessionID := uuid.New().String()
	seed := time.Now().UnixNano()
	state := ConvertScenarioToState(scenario, seed)

	// Store state
	s.stateManager.SetState(sessionID, state)

	// Persist to database
	if err := s.eventStore.CreateSession(sessionID, scenario.Name); err != nil {
		return "", State{}, fmt.Errorf("failed to create session: %w", err)
	}

	if err := s.eventStore.SaveSnapshot(sessionID, state.Round, state); err != nil {
		log.Printf("Failed to save initial snapshot: %v", err)
	}

	log.Printf("Created new session %s with scenario %s", sessionID, scenario.Name)

	return sessionID, state, nil
}

// GetSession retrieves a session by ID
func (s *SessionService) GetSession(sessionID string) (State, error) {
	state, exists := s.stateManager.GetState(sessionID)
	if !exists {
		return State{}, fmt.Errorf("session not found")
	}
	return state, nil
}
