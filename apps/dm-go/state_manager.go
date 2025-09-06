package main

import (
	"sync"
)

// StateManager provides thread-safe access to game states
type StateManager struct {
	mu     sync.RWMutex
	states map[string]State
}

// NewStateManager creates a new state manager
func NewStateManager() *StateManager {
	return &StateManager{
		states: make(map[string]State),
	}
}

// GetState retrieves a state by session ID
func (sm *StateManager) GetState(sessionID string) (State, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	state, exists := sm.states[sessionID]
	return state, exists
}

// SetState sets a state for a session ID
func (sm *StateManager) SetState(sessionID string, state State) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.states[sessionID] = state
}

// DeleteState removes a state by session ID
func (sm *StateManager) DeleteState(sessionID string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	delete(sm.states, sessionID)
}

// GetAllStates returns a copy of all states
func (sm *StateManager) GetAllStates() map[string]State {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	result := make(map[string]State)
	for k, v := range sm.states {
		result[k] = v
	}
	return result
}

// GetStateCount returns the number of active states
func (sm *StateManager) GetStateCount() int {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return len(sm.states)
}
