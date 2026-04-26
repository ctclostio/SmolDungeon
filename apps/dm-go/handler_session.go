package main

import (
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
)

// handleCreateSession creates a new game session
func handleCreateSession(c *fiber.Ctx) error {
	var req struct {
		ScenarioName string `json:"scenarioName"`
	}

	if err := c.BodyParser(&req); err != nil {
		// Allow empty body
		req.ScenarioName = ""
	}

	// Use session service to create session
	sessionID, state, err := sessionService.CreateSession(req.ScenarioName)
	if err != nil {
		log.Printf("Failed to create session: %v", err)
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	// If the game starts with an AI turn, process it immediately
	dummyResolution := Resolution{
		State:  state,
		Events: []Event{},
		Logs:   []string{},
	}
	finalResolution := gameService.ProcessAITurns(sessionID, dummyResolution)

	return c.JSON(fiber.Map{
		"sessionId": sessionID,
		"state":     finalResolution.State,
	})
}

// handleGetSession retrieves a session by ID
func handleGetSession(c *fiber.Ctx) error {
	sessionID := c.Params("sessionId")

	state, err := sessionService.GetSession(sessionID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Session not found"})
	}

	return c.JSON(fiber.Map{
		"sessionId": sessionID,
		"state":     state,
	})
}

// handleGetSessionState returns just the game state for a session
func handleGetSessionState(c *fiber.Ctx) error {
	sessionID := c.Params("sessionId")

	state, err := sessionService.GetSession(sessionID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Session not found"})
	}

	return c.JSON(state)
}

// handleGetScenariosJSON returns available scenarios as JSON array
func handleGetScenariosJSON(c *fiber.Ctx) error {
	scenarios, err := scenarioService.GetAvailableScenarios()
	if err != nil {
		log.Printf("Failed to get scenarios: %v", err)
		return c.Status(500).JSON(fiber.Map{"error": "Failed to load scenarios"})
	}

	return c.JSON(scenarios)
}

// handleHealthCheck returns server health status
func handleHealthCheck(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status":    "ok",
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

// handleSessionsOverview returns active sessions count
func handleSessionsOverview(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"activeSessions": stateManager.GetStateCount(),
		"totalSessions":  stateManager.GetStateCount(),
	})
}
