package main

import (
	"log"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

// handleWebSocket handles WebSocket connections for real-time game updates
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

		log.Printf("Received WebSocket message: %v", msg)

		// Check message type
		msgType, ok := msg["type"].(string)
		if !ok || msgType != "action" {
			continue
		}

		// Get action from message
		actionStr, ok := msg["action"].(string)
		if !ok {
			continue
		}

		// Get current state
		state, exists := stateManager.GetState(sessionID)
		if !exists {
			c.WriteJSON(fiber.Map{"error": "Session not found"})
			continue
		}

		// Get current character
		currentChar := GetCurrentCharacter(state)
		if currentChar == nil {
			c.WriteJSON(fiber.Map{"error": "No current character"})
			continue
		}

		// Create action based on message
		action := createActionFromWebSocket(actionStr, currentChar, state)
		if action.Kind == "" {
			c.WriteJSON(fiber.Map{"error": "Unknown action"})
			continue
		}
		if err := validatePlayerAction(state, action); err != nil {
			c.WriteJSON(fiber.Map{"error": "Invalid action: " + err.Error()})
			continue
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

		log.Printf("Applied action %s for session %s: %s", actionStr, sessionID, strings.Join(resolution.Logs, "; "))

		finalResolution := gameService.ProcessAITurns(sessionID, resolution)

		// Send response back via WebSocket
		c.WriteJSON(fiber.Map{
			"success": true,
			"logs":    finalResolution.Logs,
			"state":   finalResolution.State,
		})

		// Broadcast update to all clients
		broadcastGameUpdate(sessionID, finalResolution.State)
	}

	// Clean up on disconnect
	clientsMutex.Lock()
	delete(clients, sessionID)
	clientsMutex.Unlock()

	log.Printf("WebSocket client disconnected for session %s", sessionID)
}

// createActionFromWebSocket creates an action from a WebSocket message
func createActionFromWebSocket(actionStr string, currentChar *Character, state State) Action {
	return createActionFromClientAction(actionStr, currentChar, state)
}

// broadcastGameUpdate broadcasts game state update to WebSocket clients
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
