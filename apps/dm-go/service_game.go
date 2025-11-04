package main

import (
	"fmt"
	"log"
	"time"
)

// GameService handles game business logic
type GameService struct {
	eventStore      EventStoreInterface
	stateManager    *StateManager
	aiDecisionMaker *AIDecisionMaker
}

// NewGameService creates a new game service
func NewGameService(
	eventStore EventStoreInterface,
	stateManager *StateManager,
	aiDecisionMaker *AIDecisionMaker,
) *GameService {
	return &GameService{
		eventStore:      eventStore,
		stateManager:    stateManager,
		aiDecisionMaker: aiDecisionMaker,
	}
}

// ProcessAITurns automatically executes consecutive AI turns until a player turn or game end
func (s *GameService) ProcessAITurns(sessionID string, initialResolution Resolution) Resolution {
	currentResolution := initialResolution
	maxAITurns := 10 // Safety limit to prevent infinite loops

	for i := 0; i < maxAITurns; i++ {
		// Check if game is complete
		if currentResolution.State.IsComplete {
			log.Printf("Game complete, stopping AI turns")
			break
		}

		// Get current character
		currentChar := GetCurrentCharacter(currentResolution.State)
		if currentChar == nil {
			log.Printf("No current character, stopping AI turns")
			break
		}

		// Skip dead characters
		if currentChar.Stats.HP <= 0 {
			log.Printf("Skipping dead character %s", currentChar.Name)
			// Advance turn and continue
			dummyAction := Action{
				Kind:  "Defend",
				Actor: currentChar.ID,
			}
			seed := time.Now().UnixNano()
			skipResolution := ApplyAction(currentResolution.State, dummyAction, seed)
			currentResolution.State = skipResolution.State
			s.stateManager.SetState(sessionID, skipResolution.State)
			continue
		}

		// Check if it's a player's turn
		if currentChar.IsPlayer {
			log.Printf("Player's turn (%s), stopping AI execution", currentChar.Name)
			break
		}

		// It's an AI turn - decide and execute action
		log.Printf("AI turn for %s (HP: %d/%d)", currentChar.Name, currentChar.Stats.HP, currentChar.Stats.MaxHP)

		aiAction := s.aiDecisionMaker.DecideAction(currentResolution.State, currentChar.ID)

		// Generate new seed for AI action
		seed := time.Now().UnixNano()

		// Apply AI action
		aiResolution := ApplyAction(currentResolution.State, aiAction, seed)

		// Append AI logs to the overall logs
		currentResolution.Logs = append(currentResolution.Logs, aiResolution.Logs...)
		currentResolution.Events = append(currentResolution.Events, aiResolution.Events...)
		currentResolution.State = aiResolution.State

		// Update state in manager
		s.stateManager.SetState(sessionID, aiResolution.State)

		// Persist AI action
		if err := s.eventStore.AppendEvents(sessionID, aiResolution.State.Round, aiResolution.Events); err != nil {
			log.Printf("Failed to append AI events: %v", err)
		}

		if aiResolution.State.Round > currentResolution.State.Round {
			if err := s.eventStore.SaveSnapshot(sessionID, aiResolution.State.Round, aiResolution.State); err != nil {
				log.Printf("Failed to save AI snapshot: %v", err)
			}
		}

		// Broadcast AI action to WebSocket clients
		broadcastGameUpdate(sessionID, aiResolution.State)

		// Small delay for better UX (so players can see AI actions)
		time.Sleep(100 * time.Millisecond)
	}

	return currentResolution
}

// PersistResolution saves a resolution to the event store
func (s *GameService) PersistResolution(sessionID string, resolution Resolution, previousRound int) error {
	if err := s.eventStore.AppendEvents(sessionID, resolution.State.Round, resolution.Events); err != nil {
		return fmt.Errorf("failed to append events: %w", err)
	}

	if resolution.State.Round > previousRound {
		if err := s.eventStore.SaveSnapshot(sessionID, resolution.State.Round, resolution.State); err != nil {
			return fmt.Errorf("failed to save snapshot: %w", err)
		}
	}

	return nil
}
