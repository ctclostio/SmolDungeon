package main

import (
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/websocket/v2"
	"github.com/google/uuid"
	"gopkg.in/yaml.v3"
)

var (
	eventStore     EventStoreInterface
	llmClient      *LLMClient
	stateManager   *StateManager
	templateEngine *TemplateEngine
	clients        = make(map[string]*websocket.Conn)
	clientsMutex   sync.RWMutex
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
	// Check if we should run in demo mode (no SQLite)
	if len(os.Args) > 1 && os.Args[1] == "demo" {
		fmt.Println("🎮 Starting SmolDungeon in DEMO mode (no SQLite required)")
		DemoServer()
		return
	}

	// Normal mode with SQLite
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
	// Use SQLite database for persistence
	var err error
	eventStore, err = NewEventStore(dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	log.Printf("Using SQLite database for persistence")

	// Initialize state manager for thread-safe state access
	stateManager = NewStateManager()
	log.Printf("Initialized state manager")

	// Load existing sessions from DB
	es := eventStore.(*EventStore) // Cast to access db
	rows, err := es.db.Query("SELECT id FROM sessions WHERE status = 'active'")
	if err != nil {
		log.Printf("Failed to load sessions: %v", err)
	} else {
		defer rows.Close()
		loadedCount := 0
		for rows.Next() {
			var sessionID string
			if err := rows.Scan(&sessionID); err != nil {
				log.Printf("Scan error: %v", err)
				continue
			}
			if snapshot, err := eventStore.GetLatestSnapshot(sessionID); err == nil && snapshot != nil {
				stateManager.SetState(sessionID, *snapshot)
				loadedCount++
			} else {
				log.Printf("Failed to load snapshot for session %s: %v", sessionID, err)
			}
		}
		log.Printf("Loaded %d active sessions from DB", loadedCount)
	}

	// Initialize template engine for Go-based web frontend
	templateEngine, err = NewTemplateEngine()
	if err != nil {
		log.Fatalf("Failed to initialize template engine: %v", err)
	}
	log.Printf("Initialized template engine")

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
			"activeSessions": stateManager.GetStateCount(),
			"totalSessions":  stateManager.GetStateCount(),
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

	// WebSocket endpoint for real-time game
	app.Get("/ws/:sessionId", websocket.New(handleWebSocket))

	// Web routes for the game interface
	app.Get("/", handleHomePage)
	app.Get("/scenarios", handleScenariosPage)
	app.Get("/game/:sessionId", handleGamePage)
	app.Post("/game/:sessionId/action", handleGameAction)
	app.Post("/game/start", handleStartGame)
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

	stateManager.SetState(sessionID, resolution.State)

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

	stateManager.SetState(req.SessionID, req.State)

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

	state, exists := stateManager.GetState(sessionID)
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

// WebSocket handler for real-time game updates
func handleWebSocket(c *websocket.Conn) {
	sessionID := c.Params("sessionId")

	// Register client
	clientsMutex.Lock()
	clients[sessionID] = c
	clientsMutex.Unlock()

	log.Printf("WebSocket client connected for session %s", sessionID)

	// Handle WebSocket messages
	for {
		var msg map[string]interface{}
		err := c.ReadJSON(&msg)
		if err != nil {
			log.Printf("WebSocket read error: %v", err)
			break
		}

		// Handle incoming messages (e.g., ping, etc.)
		log.Printf("Received WebSocket message: %v", msg)
	}

	// Clean up on disconnect
	clientsMutex.Lock()
	delete(clients, sessionID)
	clientsMutex.Unlock()

	log.Printf("WebSocket client disconnected for session %s", sessionID)
}

// Broadcast game state update to WebSocket clients
func broadcastGameUpdate(sessionID string, state State) {
	clientsMutex.RLock()
	conn, exists := clients[sessionID]
	clientsMutex.RUnlock()

	if exists {
		err := conn.WriteJSON(fiber.Map{
			"type":  "game_update",
			"state": state,
		})
		if err != nil {
			log.Printf("WebSocket broadcast error: %v", err)
		}
	}
}

// Game page handler - serves the HTML interface using Go templates
func handleGamePage(c *fiber.Ctx) error {
	sessionID := c.Params("sessionId")

	state, exists := stateManager.GetState(sessionID)
	if !exists {
		return c.Status(404).SendString("Session not found")
	}

	currentChar := GetCurrentCharacter(state)
	isPlayerTurn := currentChar != nil && currentChar.IsPlayer

	html, err := templateEngine.RenderGamePage(state, sessionID, isPlayerTurn)
	if err != nil {
		log.Printf("Template render error: %v", err)
		return c.Status(500).SendString("Internal server error")
	}

	c.Set("Content-Type", "text/html")
	return c.SendString(html)
}

// Render characters as HTML
func renderCharacters(state State) string {
	var html strings.Builder
	for _, char := range state.Characters {
		class := "character"
		if char.IsPlayer {
			class += " player"
		} else {
			class += " enemy"
		}
		html.WriteString(fmt.Sprintf(`<div class="%s">
	           <strong>%s</strong><br>
	           HP: %d/%d<br>
	           Position: (%d, %d)
	       </div>`, class, char.Name, char.Stats.HP, char.Stats.MaxHP, char.Position.X, char.Position.Y))
	}
	return html.String()
}

// Game action handler
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

	state, exists := stateManager.GetState(sessionID)
	if !exists {
		return c.Status(404).JSON(fiber.Map{"error": "Session not found"})
	}

	// Get current character
	currentChar := GetCurrentCharacter(state)
	if currentChar == nil {
		return c.Status(400).JSON(fiber.Map{"error": "No current character"})
	}

	// Create action based on request
	var action Action
	switch req.Action {
	case "attack":
		// For simplicity, attack the first enemy
		var targetID ID
		for _, char := range state.Characters {
			if !char.IsPlayer && char.Stats.HP > 0 {
				targetID = char.ID
				break
			}
		}
		if targetID == "" {
			return c.Status(400).JSON(fiber.Map{"error": "No valid target"})
		}

		// Use first weapon
		var weaponID ID
		if len(currentChar.Weapons) > 0 {
			weaponID = currentChar.Weapons[0].ID
		}

		action = Action{
			Kind:     "Attack",
			Attacker: currentChar.ID,
			Target:   targetID,
			Weapon:   weaponID,
		}

	case "defend":
		action = Action{
			Kind:  "Defend",
			Actor: currentChar.ID,
		}

	case "flee":
		action = Action{
			Kind:  "Flee",
			Actor: currentChar.ID,
		}

	default:
		return c.Status(400).JSON(fiber.Map{"error": "Unknown action"})
	}

	// Apply the action
	seed := time.Now().UnixNano()
	resolution := ApplyAction(state, action, seed)

	// Update state
	newState := resolution.State
	stateManager.SetState(sessionID, newState)

	// Persist to database
	if err := eventStore.AppendEvents(sessionID, newState.Round, resolution.Events); err != nil {
		log.Printf("Failed to append events: %v", err)
	}

	if newState.Round > state.Round {
		if err := eventStore.SaveSnapshot(sessionID, newState.Round, newState); err != nil {
			log.Printf("Failed to save snapshot: %v", err)
		}
	}

	log.Printf("Applied action %s for session %s: %s", req.Action, sessionID, strings.Join(resolution.Logs, "; "))

	// Broadcast update to WebSocket clients
	broadcastGameUpdate(sessionID, newState)

	return c.JSON(fiber.Map{"success": true, "logs": resolution.Logs})
}

// Home page handler - serves the home page using Go templates
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

// Scenarios page handler - serves the scenarios page using Go templates
func handleScenariosPage(c *fiber.Ctx) error {
	scenarios, err := GetAvailableScenarios()
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

// Start game handler
func handleStartGame(c *fiber.Ctx) error {
	scenarioName := c.FormValue("scenario")
	if scenarioName == "" {
		return c.Status(400).SendString("Scenario name is required")
	}

	// Load scenario
	scenarioPath := fmt.Sprintf("../../scenarios/%s.yaml", scenarioName)
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

	// Redirect to game page
	return c.Redirect(fmt.Sprintf("/game/%s", sessionID))
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

// LoadScenario loads a scenario from a YAML file
func LoadScenario(filename string) (*Scenario, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read scenario file: %w", err)
	}

	var scenario Scenario
	if err := yaml.Unmarshal(data, &scenario); err != nil {
		return nil, fmt.Errorf("failed to parse scenario YAML: %w", err)
	}

	return &scenario, nil
}

// ConvertScenarioToState converts a scenario to an initial game state
func ConvertScenarioToState(scenario *Scenario, seed int64) State {
	rng := NewSeededRNG(seed)

	// Convert players
	players := make([]Character, len(scenario.Players))
	for i, p := range scenario.Players {
		players[i] = convertScenarioCharacterToCharacter(p, true)
	}

	// Convert enemies
	enemies := make([]Character, len(scenario.Enemies))
	for i, e := range scenario.Enemies {
		enemies[i] = convertScenarioCharacterToCharacter(e, false)
	}

	// Combine all characters
	allCharacters := append(players, enemies...)

	// Create turn order based on initiative
	type charWithInit struct {
		id         ID
		initiative int
	}

	initiatives := make([]charWithInit, len(allCharacters))
	for i, char := range allCharacters {
		initiative := char.Stats.Speed + rng.RollD20()
		initiatives[i] = charWithInit{id: char.ID, initiative: initiative}
	}

	// Sort by initiative descending
	sort.Slice(initiatives, func(i, j int) bool {
		return initiatives[i].initiative > initiatives[j].initiative
	})

	turnOrder := make([]ID, len(initiatives))
	for i, init := range initiatives {
		turnOrder[i] = init.id
	}

	return State{
		Round:       1,
		Characters:  allCharacters,
		TurnOrder:   turnOrder,
		CurrentTurn: 0,
		IsComplete:  false,
	}
}

// convertScenarioCharacterToCharacter converts a scenario character to a game character
func convertScenarioCharacterToCharacter(sc ScenarioCharacter, isPlayer bool) Character {
	char := Character{
		ID:               NewID(),
		Name:             sc.Name,
		IsPlayer:         isPlayer,
		AbilityCooldowns: make(map[string]int),
	}

	// Convert stats
	char.Stats = Stat{
		HP:      sc.Stats.HP,
		MaxHP:   sc.Stats.MaxHP,
		Attack:  sc.Stats.Attack,
		Defense: sc.Stats.Defense,
		Speed:   sc.Stats.Speed,
	}

	// Convert position
	char.Position = Position{
		X: sc.Position.X,
		Y: sc.Position.Y,
	}

	// Convert weapons
	char.Weapons = make([]Weapon, len(sc.Weapons))
	for i, w := range sc.Weapons {
		char.Weapons[i] = Weapon{
			ID:       NewID(),
			Name:     w.Name,
			Damage:   w.Damage,
			Accuracy: w.Accuracy,
		}
	}

	// Convert abilities
	char.Abilities = make([]Ability, len(sc.Abilities))
	for i, a := range sc.Abilities {
		char.Abilities[i] = Ability{
			ID:       NewID(),
			Name:     a.Name,
			Cooldown: a.Cooldown,
			Effect:   a.Effect,
			Power:    a.Power,
		}
	}

	// Convert items
	char.Items = make([]Item, len(sc.Items))
	for i, item := range sc.Items {
		char.Items[i] = Item{
			ID:     NewID(),
			Name:   item.Name,
			Type:   item.Type,
			Effect: item.Effect,
		}
	}

	return char
}

// GetAvailableScenarios returns a list of available scenario files
func GetAvailableScenarios() ([]string, error) {
	files, err := os.ReadDir("../../scenarios")
	if err != nil {
		return nil, fmt.Errorf("failed to read scenarios directory: %w", err)
	}

	var scenarios []string
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".yaml") {
			scenarios = append(scenarios, strings.TrimSuffix(file.Name(), ".yaml"))
		}
	}

	return scenarios, nil
}

// Helper functions for HTML rendering
func renderCombatMap(state State) string {
	var html strings.Builder

	// Create 5x5 grid centered on (0,0)
	for y := -2; y <= 2; y++ {
		for x := -2; x <= 2; x++ {
			html.WriteString(`<div class="character">`)

			// Find character at this position
			var charAtPos *Character
			for i := range state.Characters {
				if state.Characters[i].Position.X == x && state.Characters[i].Position.Y == y {
					charAtPos = &state.Characters[i]
					break
				}
			}

			if charAtPos != nil {
				class := "character"
				if charAtPos.IsPlayer {
					class += " player"
				} else {
					class += " enemy"
				}

				if charAtPos.Stats.HP == 0 {
					class += " dead"
				} else if charAtPos.Stats.HP < charAtPos.Stats.MaxHP/3 {
					class += " low-health"
				}

				html.WriteString(fmt.Sprintf(`<div>%s</div>`, charAtPos.Name))
				html.WriteString(fmt.Sprintf(`<div class="health-bar"><div class="health-fill" style="width: %d%%"></div></div>`,
					(charAtPos.Stats.HP*100)/charAtPos.Stats.MaxHP))
			} else {
				html.WriteString(`<div>·</div>`)
			}

			html.WriteString(`</div>`)
		}
	}

	return html.String()
}

func renderTurnIndicator(state State) string {
	currentChar := GetCurrentCharacter(state)
	if currentChar == nil {
		return "Unknown Turn"
	}
	return fmt.Sprintf("%s's Turn", currentChar.Name)
}

func renderActionButtons(state State, isPlayerTurn bool) string {
	if !isPlayerTurn {
		return `<div id="action-buttons" style="display: none;"></div>`
	}

	return `
	<div id="action-buttons">
		<div class="action-buttons">
			<button class="btn btn-attack" onclick="sendAction('attack')">⚔️ Attack</button>
			<button class="btn btn-defend" onclick="sendAction('defend')">🛡️ Defend</button>
			<button class="btn btn-ability" onclick="sendAction('ability')">✨ Ability</button>
			<button class="btn btn-item" onclick="sendAction('item')">🎒 Use Item</button>
			<button class="btn btn-flee" onclick="sendAction('flee')">🏃 Flee</button>
		</div>
	</div>`
}

func renderStateJSON(state State) string {
	// Simple JSON representation for JavaScript
	return fmt.Sprintf(`{"round":%d,"currentTurn":%d,"isComplete":%t}`,
		state.Round, state.CurrentTurn, state.IsComplete)
}
