package main

import (
	"fmt"
	"sync"
	"testing"
)

func TestStateManager_ConcurrentAccess(t *testing.T) {
	sm := NewStateManager()

	// Test concurrent writes
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			sessionID := fmt.Sprintf("session-%d", id)
			state := State{
				Round:      1,
				Characters: []Character{},
				TurnOrder:  []ID{},
			}
			sm.SetState(sessionID, state)
		}(i)
	}

	wg.Wait()

	// Verify all states were set
	if count := sm.GetStateCount(); count != 100 {
		t.Errorf("Expected 100 states, got %d", count)
	}

	// Test concurrent reads
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			sessionID := fmt.Sprintf("session-%d", id)
			state, exists := sm.GetState(sessionID)
			if !exists {
				t.Errorf("State not found for session %s", sessionID)
			}
			if state.Round != 1 {
				t.Errorf("Expected round 1, got %d", state.Round)
			}
		}(i)
	}

	wg.Wait()
}

func TestStateManager_DeleteState(t *testing.T) {
	sm := NewStateManager()

	// Add a state
	sessionID := "test-session"
	state := State{
		Round:      1,
		Characters: []Character{},
		TurnOrder:  []ID{},
	}
	sm.SetState(sessionID, state)

	// Verify it exists
	if _, exists := sm.GetState(sessionID); !exists {
		t.Error("State should exist")
	}

	// Delete it
	sm.DeleteState(sessionID)

	// Verify it's gone
	if _, exists := sm.GetState(sessionID); exists {
		t.Error("State should not exist after deletion")
	}
}

func TestStateManager_GetAllStates(t *testing.T) {
	sm := NewStateManager()

	// Add multiple states
	for i := 0; i < 5; i++ {
		sessionID := fmt.Sprintf("session-%d", i)
		state := State{
			Round:      i + 1,
			Characters: []Character{},
			TurnOrder:  []ID{},
		}
		sm.SetState(sessionID, state)
	}

	// Get all states
	allStates := sm.GetAllStates()

	if len(allStates) != 5 {
		t.Errorf("Expected 5 states, got %d", len(allStates))
	}

	// Verify each state
	for i := 0; i < 5; i++ {
		sessionID := fmt.Sprintf("session-%d", i)
		state, exists := allStates[sessionID]
		if !exists {
			t.Errorf("State not found for session %s", sessionID)
		}
		if state.Round != i+1 {
			t.Errorf("Expected round %d, got %d", i+1, state.Round)
		}
	}
}
