package main

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/websocket/v2"
)

// DemoServer runs a simple demo server without SQLite dependencies
func DemoServer() {
	fmt.Println("🚀 Starting SmolDungeon Demo Server...")
	fmt.Println("🎮 Game will be available at: http://localhost:3000")
	fmt.Println("📊 Health check at: http://localhost:3000/health")
	fmt.Println("🎯 Scenarios at: http://localhost:3000/scenarios")

	// Initialize configuration
	scenarioDir = getEnv("SCENARIO_DIR", "../../scenarios")

	// Use memory store instead of SQLite for demo
	eventStore = NewMemoryEventStore()
	stateManager = NewStateManager()

	var err error
	templateEngine, err = NewTemplateEngine()
	if err != nil {
		log.Fatalf("Failed to create template engine: %v", err)
	}

	llmClient = NewLLMClient(LLMConfig{
		APIKey:      "demo-key",
		Model:       "gpt-3.5-turbo",
		MaxTokens:   150,
		Temperature: 0.7,
	})

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

	// CORS configuration
	app.Use(cors.New(cors.Config{
		AllowOrigins: "http://localhost:3000,http://localhost:5173",
		AllowMethods: "GET,POST,PUT,DELETE",
		AllowHeaders: "Origin,Content-Type,Accept,session-id",
		AllowCredentials: true,
	}))

	// Rate limiting
	app.Use(limiter.New(limiter.Config{
		Max:        30,
		Expiration: 1 * time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(429).JSON(fiber.Map{
				"error": "Too many requests. Please try again later.",
			})
		},
	}))

	// Routes
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":    "ok",
			"timestamp": time.Now().Format(time.RFC3339),
			"message":   "SmolDungeon Demo Server is running!",
		})
	})

	app.Get("/", handleHomePage)
	app.Get("/scenarios", handleScenariosPage)
	app.Get("/game/:sessionId", handleGamePage)
	app.Post("/game/start", handleStartGameDemo)
	app.Post("/game/:sessionId/action", handleGameAction)

	// WebSocket route for real-time updates
	app.Get("/ws/:sessionId", websocket.New(handleWebSocket))

	// Create a demo session on startup
	createDemoSession()

	log.Println("✅ Demo server initialized successfully!")
	log.Println("🌐 Opening your browser to http://localhost:3000")

	// Start server
	log.Fatal(app.Listen(":3000"))
}

// createDemoSession creates a sample game for immediate testing
func createDemoSession() {
	// Create a simple demo scenario
	player := Character{
		ID:   NewID(),
		Name: "Demo Hero",
		Stats: Stat{
			HP: 30, MaxHP: 30, Attack: 6, Defense: 4, Speed: 3,
		},
		Position: Position{X: 0, Y: 0},
		Weapons: []Weapon{
			{ID: NewID(), Name: "Steel Sword", Damage: 8, Accuracy: 85},
		},
		Abilities: []Ability{
			{ID: NewID(), Name: "Power Strike", Cooldown: 3, Effect: "damage", Power: 12},
		},
		Items: []Item{
			{ID: NewID(), Name: "Health Potion", Type: "consumable", Effect: "heal 20 HP"},
		},
		AbilityCooldowns: make(map[string]int),
		IsPlayer:         true,
	}

	goblin := Character{
		ID:   NewID(),
		Name: "Demo Goblin",
		Stats: Stat{
			HP: 15, MaxHP: 15, Attack: 4, Defense: 2, Speed: 5,
		},
		Position: Position{X: 1, Y: 1},
		Weapons: []Weapon{
			{ID: NewID(), Name: "Rusty Dagger", Damage: 5, Accuracy: 75},
		},
		Abilities: []Ability{
			{ID: NewID(), Name: "Sneaky Strike", Cooldown: 4, Effect: "damage", Power: 8},
		},
		Items:            []Item{},
		AbilityCooldowns: make(map[string]int),
		IsPlayer:         false,
	}

	// Create initial state
	seed := time.Now().UnixNano()
	state := CreateInitialState([]Character{player}, []Character{goblin}, seed)

	// Create demo session
	sessionID := "demo-session"
	stateManager.SetState(sessionID, state)

	log.Printf("✅ Created demo session: %s", sessionID)
	log.Printf("🎮 Game ready with Hero vs Goblin")
}

// handleStartGameDemo handles game start for demo
func handleStartGameDemo(c *fiber.Ctx) error {
	scenarioName := c.FormValue("scenario")
	if scenarioName == "" {
		scenarioName = "goblin-ambush"
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
		// Use demo scenario instead
		return c.Redirect("/game/demo-session")
	}

	// Create initial game state
	seed := time.Now().UnixNano()
	state := ConvertScenarioToState(scenario, seed)

	// Create session
	sessionID := "demo-" + scenarioName
	stateManager.SetState(sessionID, state)

	// Save to memory store
	if err := eventStore.CreateSession(sessionID, scenario.Name); err != nil {
		log.Printf("Failed to create session: %v", err)
	}

	if err := eventStore.SaveSnapshot(sessionID, state.Round, state); err != nil {
		log.Printf("Failed to save initial snapshot: %v", err)
	}

	log.Printf("🎯 Started new game: %s (%s)", sessionID, scenario.Name)
	return c.Redirect(fmt.Sprintf("/game/%s", sessionID))
}
