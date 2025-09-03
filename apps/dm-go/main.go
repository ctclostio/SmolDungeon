package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/google/uuid"
)

var (
	eventStore    EventStoreInterface
	llmClient     *LLMClient
	currentStates = make(map[string]State)
)

// EventStoreInterface defines the interface for event stores
type EventStoreInterface interface {
	CreateSession(sessionID, name string) error
	AppendEvents(sessionID string, round int, events []Event) error
	SaveSnapshot(sessionID string, round int, state State) error
	GetEvents(sessionID string, fromRound int) ([]Event, error)
	GetLatestSnapshot(sessionID string) (*State, error)
	GetSnapshotAtRound(sessionID string, round int) (*State, error)
	UpdateSessionStatus(sessionID, status string) error
	Close() error
}

func main() {
	// Configuration
	dbPath := getEnv("DB_PATH", "./dm-server.db")
	port := getEnv("PORT", "3000")
	llmConfig := LLMConfig{
		// Remote model settings
		BaseURL:     getEnv("LLM_BASE_URL", ""),
		APIKey:      getEnv("LLM_API_KEY", "dummy-key"),
		Model:       getEnv("LLM_MODEL", "gpt-3.5-turbo"),
		MaxTokens:   getEnvInt("LLM_MAX_TOKENS", 150),
		Temperature: getEnvFloat("LLM_TEMPERATURE", 0.7),

		// Local model settings
		LocalEnabled:     getEnvBool("LLM_LOCAL_ENABLED", false),
		LocalBaseURL:     getEnv("LLM_LOCAL_BASE_URL", "http://localhost:8000/v1"),
		LocalModel:       getEnv("LLM_LOCAL_MODEL", "gpt-oss-20b"),
		LocalMaxTokens:   getEnvInt("LLM_LOCAL_MAX_TOKENS", 200),
		LocalTemperature: getEnvFloat("LLM_LOCAL_TEMPERATURE", 0.8),

		// Model selection
		PreferredModel: getEnv("LLM_PREFERRED_MODEL", "auto"), // "remote", "local", or "auto"
	}

	// Initialize components
	// For testing without CGO, use memory store
	eventStore = NewMemoryEventStore()
	log.Printf("Using in-memory event store for testing")

	llmClient = NewLLMClient(llmConfig)

	// Setup Fiber app
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			log.Printf("Error: %v", err)
			return c.Status(500).JSON(fiber.Map{
				"error": "Internal server error",
			})
		},
	})

	// Middleware
	app.Use(logger.New())
	app.Use(cors.New())

	// Routes
	setupRoutes(app)

	// Health check
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":    "ok",
			"timestamp": time.Now().Format(time.RFC3339),
		})
	})

	// Sessions overview
	app.Get("/sessions", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"activeSessions": len(currentStates),
			"totalSessions":  len(currentStates),
		})
	})

	// Start server
	log.Printf("DM Server starting on port %s", port)
	log.Printf("Database: %s", dbPath)
	log.Println("Available endpoints:")
	log.Println("  POST /tools/get_state_summary")
	log.Println("  POST /tools/roll_check")
	log.Println("  POST /tools/apply_action")
	log.Println("  POST /llm/generate_narration")
	log.Println("  POST /llm/generate_combat_description")
	log.Println("  GET  /health")
	log.Println("  GET  /sessions")
	log.Println("  POST /sessions")
	log.Println("  GET  /sessions/:sessionId")

	if llmConfig.LocalEnabled {
		log.Printf("Local LLM Model: %s at %s", llmConfig.LocalModel, llmConfig.LocalBaseURL)
	}
	log.Printf("LLM Preferred Model: %s", llmConfig.PreferredModel)

	log.Fatal(app.Listen(":" + port))
}

func setupRoutes(app *fiber.App) {
	// Tools endpoints
	app.Post("/tools/get_state_summary", handleGetStateSummary)
	app.Post("/tools/roll_check", handleRollCheck)
	app.Post("/tools/apply_action", handleApplyAction)

	// LLM endpoints
	app.Post("/llm/generate_narration", handleGenerateNarration)
	app.Post("/llm/generate_combat_description", handleGenerateCombatDescription)

	// Session management
	app.Post("/sessions", handleCreateSession)
	app.Get("/sessions/:sessionId", handleGetSession)
}

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

func handleRollCheck(c *fiber.Ctx) error {
	var req RollCheck
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	sessionID := c.Get("session-id")
	modifier := 0

	if sessionID != "" {
		if state, exists := currentStates[sessionID]; exists {
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

func handleApplyAction(c *fiber.Ctx) error {
	var req struct {
		State  State  `json:"state"`
		Action Action `json:"action"`
		Seed   int64  `json:"seed"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if req.Seed == 0 {
		return c.Status(400).JSON(fiber.Map{"error": "State, action, and seed are required"})
	}

	resolution := ApplyAction(req.State, req.Action, req.Seed)

	sessionID := c.Get("session-id")
	if sessionID == "" {
		sessionID = uuid.New().String()
	}

	currentStates[sessionID] = resolution.State

	// Persist to database
	if err := eventStore.AppendEvents(sessionID, resolution.State.Round, resolution.Events); err != nil {
		log.Printf("Failed to append events: %v", err)
	}

	if resolution.State.Round > req.State.Round {
		if err := eventStore.SaveSnapshot(sessionID, resolution.State.Round, resolution.State); err != nil {
			log.Printf("Failed to save snapshot: %v", err)
		}
	}

	return c.JSON(resolution)
}

func handleCreateSession(c *fiber.Ctx) error {
	var req struct {
		SessionID string `json:"sessionId"`
		State     State  `json:"state"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if req.SessionID == "" || req.State.Round == 0 {
		return c.Status(400).JSON(fiber.Map{"error": "Session ID and state are required"})
	}

	currentStates[req.SessionID] = req.State

	if err := eventStore.CreateSession(req.SessionID, "Session "+req.SessionID); err != nil {
		log.Printf("Failed to create session: %v", err)
		return c.Status(500).JSON(fiber.Map{"error": "Failed to create session"})
	}

	if err := eventStore.SaveSnapshot(req.SessionID, req.State.Round, req.State); err != nil {
		log.Printf("Failed to save initial snapshot: %v", err)
	}

	return c.JSON(fiber.Map{
		"success":   true,
		"sessionId": req.SessionID,
	})
}

func handleGetSession(c *fiber.Ctx) error {
	sessionID := c.Params("sessionId")

	state, exists := currentStates[sessionID]
	if !exists {
		return c.Status(404).JSON(fiber.Map{"error": "Session not found"})
	}

	return c.JSON(fiber.Map{
		"sessionId": sessionID,
		"state":     state,
	})
}

func handleGenerateNarration(c *fiber.Ctx) error {
	var req struct {
		State    State    `json:"state"`
		Events   []string `json:"events"`
		Context  string   `json:"context,omitempty"`
		UseLocal bool     `json:"useLocal,omitempty"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	narration, err := llmClient.GenerateNarrationWithModel(req.State, req.Events, req.Context, req.UseLocal)
	if err != nil {
		log.Printf("Narration generation failed: %v", err)
		return c.Status(500).JSON(fiber.Map{"error": "Failed to generate narration"})
	}

	return c.JSON(fiber.Map{
		"narration": narration,
		"model":     req.UseLocal && llmClient.config.LocalEnabled,
	})
}

func handleGenerateCombatDescription(c *fiber.Ctx) error {
	var req struct {
		State    State      `json:"state"`
		Action   Action     `json:"action"`
		Events   []string   `json:"events"`
		Attacker *Character `json:"attacker,omitempty"`
		Target   *Character `json:"target,omitempty"`
		UseLocal bool       `json:"useLocal,omitempty"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Create enhanced context for combat description
	context := fmt.Sprintf("Combat Action: %s", req.Action.Kind)
	if req.Attacker != nil && req.Target != nil {
		context += fmt.Sprintf(" - %s attacks %s", req.Attacker.Name, req.Target.Name)
	}

	narration, err := llmClient.GenerateNarrationWithModel(req.State, req.Events, context, req.UseLocal)
	if err != nil {
		log.Printf("Combat description generation failed: %v", err)
		return c.Status(500).JSON(fiber.Map{"error": "Failed to generate combat description"})
	}

	return c.JSON(fiber.Map{
		"description": narration,
		"action":      req.Action.Kind,
		"model":       req.UseLocal && llmClient.config.LocalEnabled,
	})
}

// Helper functions
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvFloat(key string, defaultValue float32) float32 {
	if value := os.Getenv(key); value != "" {
		if floatValue, err := strconv.ParseFloat(value, 32); err == nil {
			return float32(floatValue)
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}
