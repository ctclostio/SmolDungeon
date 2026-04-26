package main

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// handleHomePage serves the home page using Go templates
func handleHomePage(c *fiber.Ctx) error {
	html, err := templateEngine.RenderHomePage()
	if err != nil {
		log.Printf("Home template render error: %v", err)
		// Fallback to simple HTML if templates fail
		return c.SendString(`
<!DOCTYPE html>
<html>
<head>
    <title>SmolDungeon - Server Running</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; background: #f5f5f5; }
        .container { max-width: 800px; margin: 0 auto; background: white; padding: 30px; border-radius: 10px; }
        h1 { color: #2c3e50; text-align: center; }
        .button { padding: 15px 25px; background: #007bff; color: white; text-decoration: none; border-radius: 8px; margin: 10px; display: inline-block; }
    </style>
</head>
<body>
    <div class="container">
        <h1>🚀 SmolDungeon Server</h1>
        <p style="text-align: center;">Template engine error, but server is running!</p>
        <div style="text-align: center;">
            <a href="/game/demo-session" class="button">🎮 Play Demo Game</a>
            <a href="/scenarios" class="button">🎯 Choose Scenario</a>
        </div>
    </div>
</body>
</html>`)
	}

	c.Set("Content-Type", "text/html")
	return c.SendString(html)
}

// handleScenariosPage serves the scenarios page using Go templates
func handleScenariosPage(c *fiber.Ctx) error {
	scenarios, err := scenarioService.GetAvailableScenarios()
	if err != nil {
		log.Printf("Failed to get scenarios: %v", err)
		scenarios = []string{}
	}

	html, err := templateEngine.RenderScenariosPage(scenarios)
	if err != nil {
		log.Printf("Scenarios template render error: %v", err)
		return c.Status(500).SendString("Internal server error")
	}

	c.Set("Content-Type", "text/html")
	return c.SendString(html)
}

// handleGamePage serves the HTML interface using Go templates
func handleGamePage(c *fiber.Ctx) error {
	sessionID := c.Params("sessionId")

	state, err := sessionService.GetSession(sessionID)
	if err != nil {
		return c.Status(404).SendString("Session not found")
	}

	currentChar := GetCurrentCharacter(state)
	if currentChar != nil && !currentChar.IsPlayer {
		resolution := Resolution{
			State:  state,
			Events: []Event{},
			Logs:   []string{},
		}
		finalResolution := gameService.ProcessAITurns(sessionID, resolution)
		state = finalResolution.State
		currentChar = GetCurrentCharacter(state)
	}
	isPlayerTurn := currentChar != nil && currentChar.IsPlayer

	html, err := templateEngine.RenderGamePage(state, sessionID, isPlayerTurn)
	if err != nil {
		log.Printf("Template render error: %v", err)
		return c.Status(500).SendString("Internal server error")
	}

	c.Set("Content-Type", "text/html")
	return c.SendString(html)
}

// handleGameAction handles game actions from the HTML interface
func handleGameAction(c *fiber.Ctx) error {
	sessionID := c.Params("sessionId")

	var req struct {
		Action  string `json:"action"`
		Target  string `json:"target,omitempty"`
		Weapon  string `json:"weapon,omitempty"`
		Ability string `json:"ability,omitempty"`
		Item    string `json:"item,omitempty"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	state, err := sessionService.GetSession(sessionID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Session not found"})
	}

	// Get current character
	currentChar := GetCurrentCharacter(state)
	if currentChar == nil {
		return c.Status(400).JSON(fiber.Map{"error": "No current character"})
	}

	// Create action based on request
	action := createActionFromHTTP(req.Action, currentChar, state)
	if action.Kind == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Unknown action"})
	}
	if err := validatePlayerAction(state, action); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid action: " + err.Error()})
	}

	// Apply the action
	seed := time.Now().UnixNano()
	resolution := ApplyAction(state, action, seed)

	// Update state
	newState := resolution.State
	stateManager.SetState(sessionID, newState)

	// Persist to database
	if err := gameService.PersistResolution(sessionID, resolution, state.Round); err != nil {
		log.Printf("Failed to persist resolution: %v", err)
	}

	log.Printf("Applied action %s for session %s: %s", req.Action, sessionID, strings.Join(resolution.Logs, "; "))

	// Broadcast update to WebSocket clients
	broadcastGameUpdate(sessionID, newState)

	finalResolution := gameService.ProcessAITurns(sessionID, resolution)
	broadcastGameUpdate(sessionID, finalResolution.State)

	return c.JSON(fiber.Map{
		"success": true,
		"logs":    finalResolution.Logs,
		"state":   finalResolution.State,
	})
}

// handleStartGame creates a new game session and redirects to game page
func handleStartGame(c *fiber.Ctx) error {
	scenarioName := c.FormValue("scenario")
	if scenarioName == "" {
		return c.Status(400).SendString("Scenario name is required")
	}

	// Validate scenario name to prevent path traversal
	if strings.Contains(scenarioName, "..") || strings.Contains(scenarioName, "/") || strings.Contains(scenarioName, "\\") {
		return c.Status(400).SendString("Invalid scenario name")
	}

	// Load scenario
	scenarioPath := filepath.Join(scenarioDir, scenarioName+".yaml")
	scenario, err := LoadScenario(scenarioPath)
	if err != nil {
		log.Printf("Failed to load scenario %s: %v", scenarioName, err)
		return c.Status(500).SendString("Failed to load scenario")
	}

	// Create initial game state
	seed := time.Now().UnixNano()
	state := ConvertScenarioToState(scenario, seed)

	// Create session
	sessionID := uuid.New().String()
	stateManager.SetState(sessionID, state)

	// Save to database
	if err := eventStore.CreateSession(sessionID, scenario.Name); err != nil {
		log.Printf("Failed to create session: %v", err)
	}

	if err := eventStore.SaveSnapshot(sessionID, state.Round, state); err != nil {
		log.Printf("Failed to save initial snapshot: %v", err)
	}

	dummyResolution := Resolution{
		State:  state,
		Events: []Event{},
		Logs:   []string{},
	}
	finalResolution := gameService.ProcessAITurns(sessionID, dummyResolution)
	stateManager.SetState(sessionID, finalResolution.State)

	// Redirect to game page
	return c.Redirect(fmt.Sprintf("/game/%s", sessionID))
}

// createActionFromHTTP creates an action from an HTTP request
func createActionFromHTTP(actionStr string, currentChar *Character, state State) Action {
	return createActionFromClientAction(actionStr, currentChar, state)
}
