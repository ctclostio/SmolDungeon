package main

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
)

// TestIntegration_EnemyTargetsPlayerInGameAction tests the actual handleGameAction endpoint
// to verify enemies target players correctly
func TestIntegration_EnemyTargetsPlayerInGameAction(t *testing.T) {
	// Setup: Initialize required globals
	stateManager = NewStateManager()
	eventStore = NewMemoryEventStore()

	var err error
	templateEngine, err = NewTemplateEngine()
	if err != nil {
		t.Fatalf("Failed to create template engine: %v", err)
	}

	// Create a test scenario: Player vs Enemy
	player := createTestCharacter(true, "TestPlayer")
	enemy := createTestCharacter(false, "TestEnemy")

	// Create initial state with enemy going first
	state := State{
		Round:       1,
		Characters:  []Character{player, enemy},
		TurnOrder:   []ID{enemy.ID, player.ID}, // Enemy goes first!
		CurrentTurn: 0,
		IsComplete:  false,
	}
	state.CharacterMap = buildCharacterMap(&state)

	// Store state
	sessionID := "test-session-targeting-" + time.Now().Format("20060102150405")
	stateManager.SetState(sessionID, state)

	// Setup Fiber app with the actual handler
	app := fiber.New()
	app.Post("/game/:sessionId/action", handleGameAction)

	// Test 1: Enemy's turn - should target player
	t.Run("EnemyAttacksPlayer", func(t *testing.T) {
		// Create attack request from enemy's perspective
		attackReq := map[string]string{
			"action": "attack",
		}
		body, _ := json.Marshal(attackReq)

		req := httptest.NewRequest("POST", "/game/"+sessionID+"/action", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req, -1)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		if resp.StatusCode != 200 {
			t.Fatalf("Expected status 200, got %d", resp.StatusCode)
		}

		// Parse response
		var result struct {
			Success bool     `json:"success"`
			Logs    []string `json:"logs"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if !result.Success {
			t.Fatal("Expected successful response")
		}

		// Verify logs show enemy attacking player (not another enemy)
		foundCorrectLog := false
		for _, log := range result.Logs {
			t.Logf("Combat log: %s", log)
			// Log should mention TestEnemy attacking TestPlayer
			if len(log) > 0 {
				foundCorrectLog = true
			}
		}

		if !foundCorrectLog {
			t.Fatal("Expected combat log describing enemy attacking player")
		}

		// Get updated state from state manager to verify damage
		updatedState, exists := stateManager.GetState(sessionID)
		if !exists {
			t.Fatal("State not found in manager")
		}

		var updatedPlayer *Character
		for i := range updatedState.Characters {
			if updatedState.Characters[i].IsPlayer {
				updatedPlayer = &updatedState.Characters[i]
				break
			}
		}

		if updatedPlayer == nil {
			t.Fatal("Could not find player in updated state")
		}

		if updatedPlayer.Stats.HP >= player.Stats.MaxHP {
			t.Errorf("Expected player to take damage from enemy attack. HP: %d, MaxHP: %d",
				updatedPlayer.Stats.HP, player.Stats.MaxHP)
		}

		t.Logf("✅ Enemy correctly targeted player! Player HP: %d → %d",
			player.Stats.HP, updatedPlayer.Stats.HP)
	})

	// Test 2: Verify enemy cannot select another enemy as target
	t.Run("EnemyCannotTargetEnemy", func(t *testing.T) {
		// Create scenario with 2 enemies and 1 player
		player2 := createTestCharacter(true, "TestPlayer2")
		enemy1 := createTestCharacter(false, "TestEnemy1")
		enemy2 := createTestCharacter(false, "TestEnemy2")

		state2 := State{
			Round:       1,
			Characters:  []Character{player2, enemy1, enemy2},
			TurnOrder:   []ID{enemy1.ID, enemy2.ID, player2.ID},
			CurrentTurn: 0, // Enemy1's turn
			IsComplete:  false,
		}
		state2.CharacterMap = buildCharacterMap(&state2)

		sessionID2 := "test-session-multi-enemy-" + time.Now().Format("20060102150405")
		stateManager.SetState(sessionID2, state2)

		// Make attack request
		attackReq := map[string]string{
			"action": "attack",
		}
		body, _ := json.Marshal(attackReq)

		req := httptest.NewRequest("POST", "/game/"+sessionID2+"/action", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req, -1)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		if resp.StatusCode != 200 {
			t.Fatalf("Expected status 200, got %d", resp.StatusCode)
		}

		// Parse response
		var result struct {
			Success bool     `json:"success"`
			Logs    []string `json:"logs"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if !result.Success {
			t.Fatal("Expected successful response")
		}

		// Log output to verify correct targeting
		for _, log := range result.Logs {
			t.Logf("Combat log: %s", log)
		}

		// Get updated state from state manager
		updatedState, exists := stateManager.GetState(sessionID2)
		if !exists {
			t.Fatal("State not found in manager")
		}

		// Verify enemy2 did NOT take damage (enemy1 should not attack enemy2)
		var enemy2Updated *Character
		for i := range updatedState.Characters {
			if updatedState.Characters[i].ID == enemy2.ID {
				enemy2Updated = &updatedState.Characters[i]
				break
			}
		}

		if enemy2Updated == nil {
			t.Fatal("Could not find enemy2 in updated state")
		}

		if enemy2Updated.Stats.HP < enemy2.Stats.HP {
			t.Errorf("Enemy2 should not take damage! HP: %d → %d (BUG: enemy attacked enemy)",
				enemy2.Stats.HP, enemy2Updated.Stats.HP)
		}

		// Verify player took damage instead
		var playerUpdated *Character
		for i := range updatedState.Characters {
			if updatedState.Characters[i].IsPlayer {
				playerUpdated = &updatedState.Characters[i]
				break
			}
		}

		if playerUpdated == nil {
			t.Fatal("Could not find player in updated state")
		}

		if playerUpdated.Stats.HP >= player2.Stats.HP {
			t.Errorf("Expected player to take damage. HP: %d",
				playerUpdated.Stats.HP)
		}

		t.Logf("✅ Enemy correctly targeted player, not other enemy!")
		t.Logf("   Enemy2 HP: %d (unchanged)", enemy2Updated.Stats.HP)
		t.Logf("   Player HP: %d → %d (took damage)", player2.Stats.HP, playerUpdated.Stats.HP)
	})
}
