package main

import (
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/websocket/v2"

	"github.com/smol-dungeon/dm-go/middleware"
)

// Global dependencies (minimized)
var (
	eventStore      EventStoreInterface
	llmClient       *LLMClient
	stateManager    *StateManager
	templateEngine  *TemplateEngine
	aiDecisionMaker *AIDecisionMaker
	scenarioDir     string

	// Services
	gameService     *GameService
	sessionService  *SessionService
	scenarioService *ScenarioService

	// WebSocket clients
	clients      = make(map[string]*websocket.Conn)
	clientsMutex sync.RWMutex
)

func main() {
	// Check for demo mode
	if len(os.Args) > 1 && os.Args[1] == "demo" {
		fmt.Println("🎮 Starting SmolDungeon in DEMO mode (no SQLite required)")
		DemoServer()
		return
	}

	// Load configuration
	config := loadConfig()

	// Initialize core components
	if err := initializeComponents(config); err != nil {
		log.Fatalf("Failed to initialize components: %v", err)
	}

	// Initialize services
	initializeServices()

	// Load existing sessions
	loadActiveSessions()

	// Setup Fiber app
	app := setupFiberApp()

	// Start server
	logServerInfo(config.Port, config.LLMConfig)
	log.Fatal(app.Listen(":" + config.Port))
}

// Config holds application configuration
type Config struct {
	DBPath      string
	Port        string
	ScenarioDir string
	LLMConfig   LLMConfig
}

// loadConfig loads configuration from environment variables
func loadConfig() Config {
	return Config{
		DBPath:      getEnv("DB_PATH", "./dm-server.db"),
		Port:        getEnv("PORT", "3000"),
		ScenarioDir: getEnv("SCENARIO_DIR", "../../scenarios"),
		LLMConfig: LLMConfig{
			BaseURL:          getEnv("LLM_BASE_URL", ""),
			APIKey:           getEnv("LLM_API_KEY", "dummy-key"),
			Model:            getEnv("LLM_MODEL", "gpt-3.5-turbo"),
			MaxTokens:        getEnvInt("LLM_MAX_TOKENS", 150),
			Temperature:      getEnvFloat("LLM_TEMPERATURE", 0.7),
			LocalEnabled:     getEnvBool("LLM_LOCAL_ENABLED", false),
			LocalBaseURL:     getEnv("LLM_LOCAL_BASE_URL", "http://localhost:8000/v1"),
			LocalModel:       getEnv("LLM_LOCAL_MODEL", "gpt-oss-20b"),
			LocalMaxTokens:   getEnvInt("LLM_LOCAL_MAX_TOKENS", 200),
			LocalTemperature: getEnvFloat("LLM_LOCAL_TEMPERATURE", 0.8),
			PreferredModel:   getEnv("LLM_PREFERRED_MODEL", "auto"),
		},
	}
}

// initializeComponents initializes core application components
func initializeComponents(config Config) error {
	var err error

	// Initialize event store
	eventStore, err = NewEventStore(config.DBPath)
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	log.Printf("Using SQLite database for persistence")

	// Initialize state manager
	stateManager = NewStateManager()
	log.Printf("Initialized state manager")

	// Initialize template engine
	templateEngine, err = NewTemplateEngine()
	if err != nil {
		return fmt.Errorf("failed to initialize template engine: %w", err)
	}
	log.Printf("Initialized template engine")

	// Initialize LLM client
	llmClient = NewLLMClient(config.LLMConfig)

	// Initialize AI decision maker
	aiDecisionMaker = NewAIDecisionMaker()
	log.Printf("Initialized AI decision maker")

	// Set scenario directory
	scenarioDir = config.ScenarioDir

	return nil
}

// initializeServices creates service instances
func initializeServices() {
	gameService = NewGameService(eventStore, stateManager, aiDecisionMaker)
	sessionService = NewSessionService(eventStore, stateManager, scenarioDir)
	scenarioService = NewScenarioService(scenarioDir)
}

// loadActiveSessions loads active sessions from the database
func loadActiveSessions() {
	es := eventStore.(*EventStore)
	rows, err := es.db.Query("SELECT id FROM sessions WHERE status = 'active'")
	if err != nil {
		log.Printf("Failed to load sessions: %v", err)
		return
	}
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

// setupFiberApp configures and returns a Fiber application
func setupFiberApp() *fiber.App {
	app := fiber.New(fiber.Config{
		ErrorHandler: middleware.ErrorHandler,
	})

	// Middleware
	app.Use(logger.New())
	app.Use(middleware.CORS())
	app.Use(middleware.RateLimit())

	// Routes
	setupRoutes(app)

	return app
}

// logServerInfo logs server startup information
func logServerInfo(port string, llmConfig LLMConfig) {
	log.Printf("DM Server starting on port %s", port)
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
}
