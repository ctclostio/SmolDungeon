package main

import (
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
)

// handleGetStateSummary returns a summary of the game state
func handleGetStateSummary(c *fiber.Ctx) error {
	var req struct {
		State State `json:"state"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if req.State.Round == 0 { // Basic validation
		return c.Status(400).JSON(fiber.Map{"error": "State is required"})
	}

	summary := GetStateSummary(req.State)
	return c.JSON(fiber.Map{"summary": summary})
}

// handleRollCheck handles dice roll checks
func handleRollCheck(c *fiber.Ctx) error {
	var req RollCheck
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Validate roll check request
	if req.Actor == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Actor ID is required"})
	}
	validTypes := map[string]bool{
		"attack": true, "defense": true, "skill": true, "save": true,
	}
	if !validTypes[req.Type] {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid roll type. Must be one of: attack, defense, skill, save",
		})
	}
	if req.DC < 0 {
		return c.Status(400).JSON(fiber.Map{"error": "DC cannot be negative"})
	}

	sessionID := c.Get("session-id")
	modifier := 0

	if sessionID != "" {
		if state, exists := stateManager.GetState(sessionID); exists {
			character := GetCharacterByID(state, req.Actor)
			if character != nil {
				switch req.Type {
				case "attack":
					modifier = character.Stats.Attack
				case "defense":
					modifier = character.Stats.Defense
				case "skill", "save":
					modifier = character.Stats.Speed / 2
				}
			}
		}
	}

	seed := time.Now().UnixNano()
	rng := NewSeededRNG(seed)
	roll := rng.RollD20()
	total := roll + modifier
	success := total >= req.DC

	result := RollResult{
		Roll:     roll,
		Modifier: modifier,
		Total:    total,
		Success:  success,
	}

	return c.JSON(result)
}

// handleApplyAction applies a game action and processes AI turns
func handleApplyAction(c *fiber.Ctx) error {
	var req struct {
		SessionID string `json:"sessionId"`
		Action    Action `json:"action"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Validate sessionId
	if req.SessionID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "sessionId is required"})
	}

	// Get current state from state manager
	state, exists := stateManager.GetState(req.SessionID)
	if !exists {
		return c.Status(404).JSON(fiber.Map{"error": "Session not found"})
	}

	// Validate action
	if req.Action.Kind == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid action: kind is required"})
	}

	// Validate action has required actor
	if req.Action.Actor == "" && req.Action.Attacker == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid action: actor or attacker is required"})
	}

	// Generate seed
	seed := time.Now().UnixNano()

	// Apply action
	resolution := ApplyAction(state, req.Action, seed)

	// Update state
	stateManager.SetState(req.SessionID, resolution.State)

	// Persist to database
	if err := gameService.PersistResolution(req.SessionID, resolution, state.Round); err != nil {
		log.Printf("Failed to persist resolution: %v", err)
	}

	// Broadcast update to WebSocket clients
	broadcastGameUpdate(req.SessionID, resolution.State)

	// Auto-execute AI turns until it's a player's turn or game ends
	log.Printf("Processing AI turns after player action (current turn: %d)", resolution.State.CurrentTurn)
	currentChar := GetCurrentCharacter(resolution.State)
	if currentChar != nil {
		log.Printf("Current character after player action: %s (IsPlayer: %v)", currentChar.Name, currentChar.IsPlayer)
	}
	finalResolution := gameService.ProcessAITurns(req.SessionID, resolution)
	log.Printf("AI processing complete. Final current turn: %d", finalResolution.State.CurrentTurn)

	// Broadcast final state after AI turns
	broadcastGameUpdate(req.SessionID, finalResolution.State)

	return c.JSON(finalResolution)
}
