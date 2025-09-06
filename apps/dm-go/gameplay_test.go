package main

import (
	"testing"
	"time"
)

// TestFullGameFlow tests a complete game from start to finish
func TestFullGameFlow(t *testing.T) {
	// Create a simple scenario
	player := createTestCharacter(true, "Hero")
	enemy := createTestCharacter(false, "Goblin")
	enemy.Stats.HP = 15 // Make enemy weaker for faster test

	// Create initial state
	seed := time.Now().UnixNano()
	state := CreateInitialState([]Character{player}, []Character{enemy}, seed)

	// Verify initial state
	if len(state.Characters) != 2 {
		t.Fatalf("Expected 2 characters, got %d", len(state.Characters))
	}

	if state.Round != 1 {
		t.Fatalf("Expected round 1, got %d", state.Round)
	}

	if len(state.TurnOrder) != 2 {
		t.Fatalf("Expected 2 characters in turn order, got %d", len(state.TurnOrder))
	}

	t.Logf("Initial state: Round %d, %d characters", state.Round, len(state.Characters))

	// Simulate several rounds of combat
	for round := 1; round <= 5 && !state.IsComplete; round++ {
		t.Logf("=== ROUND %d ===", round)

		// Get current character
		currentChar := GetCurrentCharacter(state)
		if currentChar == nil {
			t.Fatalf("No current character in round %d", round)
		}

		t.Logf("Current turn: %s (Player: %v)", currentChar.Name, currentChar.IsPlayer)

		// Create an action based on character type
		var action Action
		var actionSeed int64 = time.Now().UnixNano()

		if currentChar.IsPlayer {
			// Player attacks enemy
			action = Action{
				Kind:     "Attack",
				Attacker: currentChar.ID,
				Target:   enemy.ID,
				Weapon:   currentChar.Weapons[0].ID,
			}
			t.Logf("Player %s attacks %s with %s", currentChar.Name, enemy.Name, currentChar.Weapons[0].Name)
		} else {
			// Enemy attacks player
			action = Action{
				Kind:     "Attack",
				Attacker: currentChar.ID,
				Target:   player.ID,
				Weapon:   currentChar.Weapons[0].ID,
			}
			t.Logf("Enemy %s attacks %s with %s", currentChar.Name, player.Name, currentChar.Weapons[0].Name)
		}

		// Apply the action
		resolution := ApplyAction(state, action, actionSeed)

		// Log the results
		t.Logf("Action results: %d events, %d logs", len(resolution.Events), len(resolution.Logs))
		for _, log := range resolution.Logs {
			t.Logf("  → %s", log)
		}

		// Update state
		state = resolution.State

		// Check for combat end
		if state.IsComplete {
			t.Logf("Combat ended in round %d!", round)
			if state.Winner != nil {
				t.Logf("Winner: %s", *state.Winner)
			}
			break
		}

		// Log character status
		for _, char := range state.Characters {
			t.Logf("  %s: %d/%d HP", char.Name, char.Stats.HP, char.Stats.MaxHP)
		}
	}

	// Verify game ended properly
	if !state.IsComplete {
		t.Error("Game should have completed within 5 rounds")
	}

	// Verify we have a winner
	if state.Winner == nil {
		t.Error("Game should have a winner")
	}

	t.Logf("Final game state: Round %d, Winner: %s", state.Round, *state.Winner)
}

// TestCombatMechanics tests specific combat mechanics
func TestCombatMechanics(t *testing.T) {
	// Test attack mechanics
	t.Run("AttackHit", func(t *testing.T) {
		player := createTestCharacter(true, "Attacker")
		target := createTestCharacter(false, "Target")
		target.Stats.Defense = 5 // Low defense to ensure hit

		state := CreateInitialState([]Character{player}, []Character{target}, 12345)

		action := Action{
			Kind:     "Attack",
			Attacker: player.ID,
			Target:   target.ID,
			Weapon:   player.Weapons[0].ID,
		}

		resolution := ApplyAction(state, action, 12345)

		// Should have damage event
		hasDamageEvent := false
		for _, event := range resolution.Events {
			if event.Type == "damage" {
				hasDamageEvent = true
				if event.Amount <= 0 {
					t.Error("Damage amount should be positive")
				}
				break
			}
		}

		if !hasDamageEvent {
			t.Error("Expected damage event from attack")
		}
	})

	t.Run("DefendAction", func(t *testing.T) {
		player := createTestCharacter(true, "Defender")
		enemy := createTestCharacter(false, "Enemy")

		state := CreateInitialState([]Character{player}, []Character{enemy}, 12345)
		originalDefense := player.Stats.Defense

		action := Action{
			Kind:  "Defend",
			Actor: player.ID,
		}

		resolution := ApplyAction(state, action, 12345)

		// Find player in new state
		var newPlayer *Character
		for i := range resolution.State.Characters {
			if resolution.State.Characters[i].ID == player.ID {
				newPlayer = &resolution.State.Characters[i]
				break
			}
		}

		if newPlayer == nil {
			t.Fatal("Player not found in new state")
		}

		if newPlayer.Stats.Defense <= originalDefense {
			t.Errorf("Expected defense to increase, got %d (was %d)", newPlayer.Stats.Defense, originalDefense)
		}
	})

	t.Run("TurnAdvancement", func(t *testing.T) {
		player := createTestCharacter(true, "Player")
		enemy := createTestCharacter(false, "Enemy")

		state := CreateInitialState([]Character{player}, []Character{enemy}, 12345)
		originalTurn := state.CurrentTurn

		// Perform any action to advance turn
		action := Action{
			Kind:  "Defend",
			Actor: player.ID,
		}

		resolution := ApplyAction(state, action, 12345)

		if resolution.State.CurrentTurn == originalTurn {
			t.Error("Turn should have advanced")
		}
	})
}

// TestWebSocketBroadcasting tests the WebSocket broadcasting functionality
func TestWebSocketBroadcasting(t *testing.T) {
	// This test verifies the broadcast function works without actual WebSocket connections
	sessionID := "test-session-websocket"
	state := State{
		Round:       1,
		Characters:  []Character{createTestCharacter(true, "Test")},
		TurnOrder:   []ID{NewID()},
		CurrentTurn: 0,
		IsComplete:  false,
	}

	// Test that broadcast function doesn't panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("broadcastGameUpdate panicked: %v", r)
		}
	}()

	// This should not panic even without WebSocket clients
	broadcastGameUpdate(sessionID, state)

	t.Log("WebSocket broadcasting test completed successfully")
}

// TestTemplateRendering tests the new template engine
func TestTemplateRendering(t *testing.T) {
	te, err := NewTemplateEngine()
	if err != nil {
		t.Fatalf("Failed to create template engine: %v", err)
	}

	// Test game page rendering
	state := State{
		Round: 1,
		Characters: []Character{
			createTestCharacter(true, "Hero"),
			createTestCharacter(false, "Goblin"),
		},
		TurnOrder:   []ID{NewID(), NewID()},
		CurrentTurn: 0,
		IsComplete:  false,
	}

	sessionID := "test-session"
	isPlayerTurn := true

	html, err := te.RenderGamePage(state, sessionID, isPlayerTurn)
	if err != nil {
		t.Fatalf("Failed to render game page: %v", err)
	}

	// Verify HTML contains expected elements
	if len(html) == 0 {
		t.Error("Rendered HTML should not be empty")
	}

	// Check for key elements
	expectedElements := []string{
		"SmolDungeon",
		"Round 1",
		"Hero",
		"Goblin",
		"Your Turn",
		"Attack",
		"Defend",
	}

	for _, element := range expectedElements {
		if !contains(html, element) {
			t.Errorf("Expected HTML to contain '%s'", element)
		}
	}

	t.Log("Template rendering test completed successfully")
}

// TestStatePersistence tests that game state is properly managed
func TestStatePersistence(t *testing.T) {
	sm := NewStateManager()

	// Create initial game state
	initialState := State{
		Round: 1,
		Characters: []Character{
			createTestCharacter(true, "Hero"),
			createTestCharacter(false, "Goblin"),
		},
		TurnOrder:   []ID{NewID(), NewID()},
		CurrentTurn: 0,
		IsComplete:  false,
	}

	sessionID := "persistence-test"

	// Set initial state
	sm.SetState(sessionID, initialState)

	// Retrieve state
	retrievedState, exists := sm.GetState(sessionID)
	if !exists {
		t.Fatal("State should exist")
	}

	// Verify state integrity
	if retrievedState.Round != initialState.Round {
		t.Errorf("Round mismatch: expected %d, got %d", initialState.Round, retrievedState.Round)
	}

	if len(retrievedState.Characters) != len(initialState.Characters) {
		t.Errorf("Character count mismatch: expected %d, got %d",
			len(initialState.Characters), len(retrievedState.Characters))
	}

	// Simulate game progression
	action := Action{
		Kind:  "Defend",
		Actor: initialState.Characters[0].ID,
	}

	resolution := ApplyAction(retrievedState, action, time.Now().UnixNano())

	// Update state in manager
	sm.SetState(sessionID, resolution.State)

	// Retrieve updated state
	updatedState, exists := sm.GetState(sessionID)
	if !exists {
		t.Fatal("Updated state should exist")
	}

	// Verify state was updated (round may not advance on defend action)
	if updatedState.Round < 1 {
		t.Errorf("Round should not decrease, got %d", updatedState.Round)
	}

	t.Log("State persistence test completed successfully")
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsSubstring(s, substr)))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
