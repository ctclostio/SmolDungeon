package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
)

// Color codes for terminal output
const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
	ColorPurple = "\033[35m"
	ColorCyan   = "\033[36m"
	ColorWhite  = "\033[37m"
	ColorBold   = "\033[1m"
)

type Character struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Stats    Stat   `json:"stats"`
	Position struct {
		X int `json:"x"`
		Y int `json:"y"`
	} `json:"position"`
	Weapons          []Weapon          `json:"weapons"`
	Abilities        []Ability         `json:"abilities"`
	Items            []interface{}     `json:"items"`
	AbilityCooldowns map[string]int    `json:"abilityCooldowns"`
	IsPlayer         bool              `json:"isPlayer"`
}

type Stat struct {
	HP      int `json:"hp"`
	MaxHP   int `json:"maxHp"`
	Attack  int `json:"attack"`
	Defense int `json:"defense"`
	Speed   int `json:"speed"`
}

type Weapon struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Damage   int    `json:"damage"`
	Accuracy int    `json:"accuracy"`
}

type Ability struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Cooldown int    `json:"cooldown"`
	Effect   string `json:"effect"`
	Power    int    `json:"power"`
}

type GameState struct {
	Round       int         `json:"round"`
	Characters  []Character `json:"characters"`
	TurnOrder   []string    `json:"turnOrder"`
	CurrentTurn int         `json:"currentTurn"`
	IsComplete  bool        `json:"isComplete"`
	Winner      *string     `json:"winner,omitempty"`
}

type Action struct {
	Kind     string `json:"kind"`
	Attacker string `json:"attacker,omitempty"`
	Target   string `json:"target,omitempty"`
	Weapon   string `json:"weapon,omitempty"`
	Actor    string `json:"actor,omitempty"`
	Ability  string `json:"ability,omitempty"`
}

type ActionResponse struct {
	State  GameState       `json:"state"`
	Events []interface{}   `json:"events"`
	Logs   []string        `json:"logs"`
}

const baseURL = "http://localhost:3000"

func main() {
	printBanner()

	// Create a new game session
	fmt.Println(ColorCyan + "🎮 Starting a new game..." + ColorReset)
	sessionID, state, err := createSession()
	if err != nil {
		fmt.Printf(ColorRed + "❌ Failed to create session: %v\n" + ColorReset, err)
		return
	}

	fmt.Printf(ColorGreen + "✅ Game session created: %s\n\n" + ColorReset, sessionID)

	// Game loop
	reader := bufio.NewReader(os.Stdin)
	for !state.IsComplete {
		// Display game state
		displayGameState(state)

		// Get current character
		currentChar := getCurrentCharacter(state)
		if currentChar == nil {
			fmt.Println(ColorRed + "Error: No current character" + ColorReset)
			break
		}

		// Server automatically handles AI turns, so if we get here and it's not
		// a player turn, something went wrong
		if !currentChar.IsPlayer {
			fmt.Println(ColorRed + "Error: Unexpected AI turn (server should auto-handle)" + ColorReset)
			break
		}

		// Player's turn
		fmt.Printf(ColorGreen + ColorBold + "\n⚔️  %s's Turn!\n" + ColorReset, currentChar.Name)
		fmt.Println(ColorCyan + "What will you do?" + ColorReset)
		fmt.Println("1. ⚔️  Attack")
		fmt.Println("2. 🛡️  Defend")
		fmt.Println("3. 🏃 Flee")
		fmt.Print(ColorYellow + "Choose action (1-3): " + ColorReset)

		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		var action Action
		switch input {
		case "1":
			// Attack - select target
			enemies := getEnemies(state)
			if len(enemies) == 0 {
				fmt.Println(ColorRed + "No enemies to attack!" + ColorReset)
				continue
			}

			fmt.Println(ColorCyan + "\nSelect target:" + ColorReset)
			for i, enemy := range enemies {
				fmt.Printf("%d. %s (HP: %d/%d)\n", i+1, enemy.Name, enemy.Stats.HP, enemy.Stats.MaxHP)
			}
			fmt.Print(ColorYellow + "Target (1-" + fmt.Sprint(len(enemies)) + "): " + ColorReset)

			targetInput, _ := reader.ReadString('\n')
			targetInput = strings.TrimSpace(targetInput)
			targetIdx, err := strconv.Atoi(targetInput)
			if err != nil || targetIdx < 1 || targetIdx > len(enemies) {
				fmt.Println(ColorRed + "Invalid target!" + ColorReset)
				continue
			}

			target := enemies[targetIdx-1]
			weapon := currentChar.Weapons[0]

			action = Action{
				Kind:     "Attack",
				Attacker: currentChar.ID,
				Target:   target.ID,
				Weapon:   weapon.ID,
			}

		case "2":
			action = Action{
				Kind:  "Defend",
				Actor: currentChar.ID,
			}

		case "3":
			action = Action{
				Kind:  "Flee",
				Actor: currentChar.ID,
			}

		default:
			fmt.Println(ColorRed + "Invalid choice!" + ColorReset)
			continue
		}

		// Apply action
		fmt.Println(ColorCyan + "\n⚡ Executing action..." + ColorReset)
		state, err = applyAction(sessionID, action)
		if err != nil {
			fmt.Printf(ColorRed + "❌ Action failed: %v\n" + ColorReset, err)
			continue
		}
	}

	// Game over
	displayGameOver(state)
}

func printBanner() {
	fmt.Println(ColorPurple + ColorBold)
	fmt.Println("╔═══════════════════════════════════════════════╗")
	fmt.Println("║                                               ║")
	fmt.Println("║           ⚔️  SMOLDUNGEON CLI ⚔️             ║")
	fmt.Println("║                                               ║")
	fmt.Println("║         Turn-Based Combat Adventure           ║")
	fmt.Println("║                                               ║")
	fmt.Println("╚═══════════════════════════════════════════════╝")
	fmt.Println(ColorReset)
}

func displayGameState(state GameState) {
	clearScreen()

	fmt.Println(ColorBold + "═══════════════════════════════════════════════" + ColorReset)
	fmt.Printf(ColorCyan + "  Round %d\n" + ColorReset, state.Round)
	fmt.Println(ColorBold + "═══════════════════════════════════════════════" + ColorReset)
	fmt.Println()

	// Display players
	fmt.Println(ColorGreen + ColorBold + "🗡️  YOUR PARTY:" + ColorReset)
	players := getPlayers(state)
	for _, char := range players {
		displayCharacter(char, true)
	}
	fmt.Println()

	// Display enemies
	fmt.Println(ColorRed + ColorBold + "👹 ENEMIES:" + ColorReset)
	enemies := getEnemies(state)
	for _, char := range enemies {
		displayCharacter(char, false)
	}
	fmt.Println()
}

func displayCharacter(char Character, isPlayer bool) {
	color := ColorGreen
	if !isPlayer {
		color = ColorRed
	}

	// Current turn indicator
	marker := "  "

	// Health bar
	healthPercent := float64(char.Stats.HP) / float64(char.Stats.MaxHP)
	barLength := 20
	filledBars := int(healthPercent * float64(barLength))

	healthBar := "["
	for i := 0; i < barLength; i++ {
		if i < filledBars {
			healthBar += "█"
		} else {
			healthBar += "░"
		}
	}
	healthBar += "]"

	// Color health bar based on HP
	healthColor := ColorGreen
	if healthPercent < 0.3 {
		healthColor = ColorRed
	} else if healthPercent < 0.6 {
		healthColor = ColorYellow
	}

	fmt.Printf("%s%s%-15s%s %sHP: %3d/%3d %s%s\n",
		marker,
		color+ColorBold,
		char.Name,
		ColorReset,
		healthColor,
		char.Stats.HP,
		char.Stats.MaxHP,
		healthBar,
		ColorReset,
	)
	fmt.Printf("     ATK: %d  DEF: %d  SPD: %d  Pos: (%d,%d)\n",
		char.Stats.Attack,
		char.Stats.Defense,
		char.Stats.Speed,
		char.Position.X,
		char.Position.Y,
	)
}

func displayGameOver(state GameState) {
	clearScreen()
	fmt.Println(ColorPurple + ColorBold)
	fmt.Println("╔═══════════════════════════════════════════════╗")
	fmt.Println("║                                               ║")
	fmt.Println("║              GAME OVER                        ║")
	fmt.Println("║                                               ║")

	if state.Winner != nil {
		if *state.Winner == "player" {
			fmt.Println("║           🎉 VICTORY! 🎉                     ║")
		} else {
			fmt.Println("║           💀 DEFEAT 💀                       ║")
		}
	} else {
		fmt.Println("║              DRAW                             ║")
	}

	fmt.Println("║                                               ║")
	fmt.Println("╚═══════════════════════════════════════════════╝")
	fmt.Println(ColorReset)

	fmt.Printf(ColorCyan + "\nBattle lasted %d rounds\n" + ColorReset, state.Round)
	fmt.Println("\nThanks for playing SmolDungeon! ⚔️")
}

func getCurrentCharacter(state GameState) *Character {
	if len(state.TurnOrder) == 0 {
		return nil
	}

	currentID := state.TurnOrder[state.CurrentTurn]
	for _, char := range state.Characters {
		if char.ID == currentID {
			return &char
		}
	}
	return nil
}

func getPlayers(state GameState) []Character {
	var players []Character
	for _, char := range state.Characters {
		if char.IsPlayer && char.Stats.HP > 0 {
			players = append(players, char)
		}
	}
	return players
}

func getEnemies(state GameState) []Character {
	var enemies []Character
	for _, char := range state.Characters {
		if !char.IsPlayer && char.Stats.HP > 0 {
			enemies = append(enemies, char)
		}
	}
	return enemies
}

func clearScreen() {
	fmt.Print("\033[H\033[2J")
}

// API Functions

func createSession() (string, GameState, error) {
	resp, err := http.Post(baseURL+"/sessions", "application/json", bytes.NewBuffer([]byte("{}")))
	if err != nil {
		return "", GameState{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", GameState{}, err
	}

	var result struct {
		SessionID string    `json:"sessionId"`
		State     GameState `json:"state"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return "", GameState{}, err
	}

	return result.SessionID, result.State, nil
}

func applyAction(sessionID string, action Action) (GameState, error) {
	payload := map[string]interface{}{
		"sessionId": sessionID,
		"action":    action,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return GameState{}, err
	}

	resp, err := http.Post(baseURL+"/tools/apply_action", "application/json", bytes.NewBuffer(data))
	if err != nil {
		return GameState{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return GameState{}, err
	}

	var result ActionResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return GameState{}, fmt.Errorf("failed to parse response: %w\nBody: %s", err, string(body))
	}

	// Print logs
	if len(result.Logs) > 0 {
		fmt.Println(ColorCyan + "\n📜 Combat Log:" + ColorReset)
		for _, log := range result.Logs {
			fmt.Println(ColorWhite + "   " + log + ColorReset)
		}
		fmt.Println()
	}

	return result.State, nil
}
