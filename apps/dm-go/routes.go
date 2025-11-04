package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

// setupRoutes configures all application routes
func setupRoutes(app *fiber.App) {
	// Health check
	app.Get("/health", handleHealthCheck)

	// Sessions overview
	app.Get("/sessions", handleSessionsOverview)

	// Tools endpoints
	app.Post("/tools/get_state_summary", handleGetStateSummary)
	app.Post("/tools/roll_check", handleRollCheck)
	app.Post("/tools/apply_action", handleApplyAction)

	// LLM endpoints
	app.Post("/llm/generate_narration", handleGenerateNarration)
	app.Post("/llm/generate_combat_description", handleGenerateCombatDescription)

	// Session management (JSON API for React frontend)
	app.Post("/sessions", handleCreateSession)
	app.Get("/sessions/:sessionId", handleGetSession)
	app.Get("/sessions/:sessionId/state", handleGetSessionState)

	// Scenarios (JSON API for React frontend)
	app.Get("/scenarios", handleGetScenariosJSON)

	// WebSocket endpoint for real-time game
	app.Get("/ws/:sessionId", websocket.New(handleWebSocket))

	// Web routes for the game interface (Go template frontend)
	app.Get("/", handleHomePage)
	app.Get("/scenarios-page", handleScenariosPage)
	app.Get("/game/:sessionId", handleGamePage)
	app.Post("/game/:sessionId/action", handleGameAction)
	app.Post("/game/start", handleStartGame)
}
