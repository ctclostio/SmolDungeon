package main

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func setupGameActionTest(t *testing.T) *fiber.App {
	t.Helper()

	stateManager = NewStateManager()
	eventStore = NewMemoryEventStore()
	aiDecisionMaker = NewAIDecisionMaker()
	gameService = NewGameService(eventStore, stateManager, aiDecisionMaker)
	sessionService = NewSessionService(eventStore, stateManager, "../../scenarios")

	var err error
	templateEngine, err = NewTemplateEngine()
	if err != nil {
		t.Fatalf("Failed to create template engine: %v", err)
	}

	app := fiber.New()
	app.Post("/game/:sessionId/action", handleGameAction)
	return app
}

func postGameAction(t *testing.T, app *fiber.App, sessionID, action string) (int, map[string]interface{}) {
	t.Helper()

	body, _ := json.Marshal(map[string]string{"action": action})
	req := httptest.NewRequest("POST", "/game/"+sessionID+"/action", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	return resp.StatusCode, result
}

func TestIntegration_PlayerActionProcessesAITurns(t *testing.T) {
	app := setupGameActionTest(t)

	player := createTestCharacter(true, "TestPlayer")
	enemy := createTestCharacter(false, "TestEnemy")
	state := State{
		Round:       1,
		Characters:  []Character{player, enemy},
		TurnOrder:   []ID{player.ID, enemy.ID},
		CurrentTurn: 0,
		IsComplete:  false,
	}
	state.CharacterMap = buildCharacterMap(&state)

	sessionID := "test-player-action"
	stateManager.SetState(sessionID, state)

	status, result := postGameAction(t, app, sessionID, "attack")
	if status != 200 {
		t.Fatalf("Expected status 200, got %d: %v", status, result)
	}

	if success, _ := result["success"].(bool); !success {
		t.Fatalf("Expected successful response, got %v", result)
	}

	updatedState, exists := stateManager.GetState(sessionID)
	if !exists {
		t.Fatal("State not found in manager")
	}

	updatedEnemy := GetCharacterByID(updatedState, enemy.ID)
	if updatedEnemy == nil {
		t.Fatal("Could not find enemy in updated state")
	}
	if updatedEnemy.Stats.HP >= enemy.Stats.HP {
		t.Fatalf("Expected enemy to take damage from player action")
	}

	if !updatedState.IsComplete {
		current := GetCurrentCharacter(updatedState)
		if current == nil {
			t.Fatal("Expected current character after AI processing")
		}
		if !current.IsPlayer {
			t.Fatalf("Expected AI turns to resolve back to a player turn, got %s", current.Name)
		}
	}
}

func TestIntegration_PublicEndpointRejectsEnemyTurn(t *testing.T) {
	app := setupGameActionTest(t)

	player := createTestCharacter(true, "TestPlayer")
	enemy := createTestCharacter(false, "TestEnemy")
	state := State{
		Round:       1,
		Characters:  []Character{player, enemy},
		TurnOrder:   []ID{enemy.ID, player.ID},
		CurrentTurn: 0,
		IsComplete:  false,
	}
	state.CharacterMap = buildCharacterMap(&state)

	sessionID := "test-enemy-action-rejected"
	stateManager.SetState(sessionID, state)

	status, result := postGameAction(t, app, sessionID, "attack")
	if status != 400 {
		t.Fatalf("Expected status 400, got %d: %v", status, result)
	}

	errText, _ := result["error"].(string)
	if !strings.Contains(errText, "not a player turn") {
		t.Fatalf("Expected not-a-player-turn error, got %q", errText)
	}

	updatedState, exists := stateManager.GetState(sessionID)
	if !exists {
		t.Fatal("State not found in manager")
	}

	updatedPlayer := GetCharacterByID(updatedState, player.ID)
	if updatedPlayer == nil {
		t.Fatal("Could not find player in updated state")
	}
	if updatedPlayer.Stats.HP != player.Stats.HP {
		t.Fatalf("Rejected enemy action should not damage player: %d -> %d", player.Stats.HP, updatedPlayer.Stats.HP)
	}
}
