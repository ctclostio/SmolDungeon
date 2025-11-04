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

		// Send response back via WebSocket
		c.WriteJSON(fiber.Map{
			"success": true,
			"logs":    resolution.Logs,
			"state":   newState,
		})

		// Broadcast update to all clients
		broadcastGameUpdate(sessionID, newState)
	}

	// Clean up on disconnect
	clientsMutex.Lock()
	delete(clients, sessionID)
	clientsMutex.Unlock()

	log.Printf("WebSocket client disconnected for session %s", sessionID)
}

// createActionFromWebSocket creates an action from a WebSocket message
func createActionFromWebSocket(actionStr string, currentChar *Character, state State) Action {
	var action Action

	switch actionStr {
	case "attack":
		// Target the opposite team
		var targetID ID
		targetIsPlayer := !currentChar.IsPlayer
		for _, char := range state.Characters {
			if char.IsPlayer == targetIsPlayer && char.Stats.HP > 0 {
				targetID = char.ID
				break
			}
		}
		if targetID == "" {
			return Action{} // Invalid action
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
	}

	return action
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
