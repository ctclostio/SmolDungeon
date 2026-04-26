package main

import (
	"fmt"
	"math"
	"sort"
	"strings"
)

// Game balance constants
const (
	// Combat constants
	AttackDC            = 10 // Difficulty class for attack rolls (target defense + this)
	FleeDC              = 15 // Difficulty class for flee attempts
	DefenseBonus        = 2  // Temporary defense bonus from Defend action
	BaseDefenseMax      = 5  // Maximum base defense before bonus is removed
	MinDamage           = 1  // Minimum damage dealt on successful hit
	AttackDamageDiv     = 2  // Divisor for attack stat when calculating damage
	DefaultFistDamage   = 1  // Damage for unarmed/default attacks
	DefaultFistAccuracy = 0  // Accuracy for unarmed/default attacks
	MaxCombatRounds     = 20 // Maximum rounds before combat ends
)

// CreateInitialState creates the initial game state from players and enemies.
// Initiative rolls determine turn order (Speed + d20), sorted descending.
// The character map is built for O(1) lookups.
// Returns a State ready for the first turn of combat.
func CreateInitialState(players, enemies []Character, seed int64) State {
	rng := NewSeededRNG(seed)
	allCharacters := append(players, enemies...)

	// Create turn order based on initiative (speed + d20)
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

	state := State{
		Round:       1,
		Characters:  allCharacters,
		TurnOrder:   turnOrder,
		CurrentTurn: 0,
		IsComplete:  false,
	}

	// Build character map for O(1) lookups
	state.CharacterMap = buildCharacterMap(&state)
	return state
}

// buildCharacterMap creates a map for O(1) character lookups
func buildCharacterMap(state *State) map[ID]*Character {
	charMap := make(map[ID]*Character, len(state.Characters))
	for i := range state.Characters {
		charMap[state.Characters[i].ID] = &state.Characters[i]
	}
	return charMap
}

// GetCurrentCharacter returns the character whose turn it is in the current round.
// Uses O(1) character map lookup if available, falls back to linear search.
// Returns nil if the turn index is out of bounds or character not found.
func GetCurrentCharacter(state State) *Character {
	if state.CurrentTurn < 0 || state.CurrentTurn >= len(state.TurnOrder) {
		return nil
	}
	currentID := state.TurnOrder[state.CurrentTurn]
	if state.CharacterMap != nil {
		return state.CharacterMap[currentID]
	}
	// Fallback to linear search if map not initialized
	for i := range state.Characters {
		if state.Characters[i].ID == currentID {
			return &state.Characters[i]
		}
	}
	return nil
}

// GetCharacterByID finds a character by ID using O(1) map lookup
func GetCharacterByID(state State, id ID) *Character {
	if state.CharacterMap != nil {
		return state.CharacterMap[id]
	}
	// Fallback to linear search if map not initialized
	for i := range state.Characters {
		if state.Characters[i].ID == id {
			return &state.Characters[i]
		}
	}
	return nil
}

// GetStateSummary returns a human-readable string summary of the game state.
// Includes current round, character HP status, combat completion, winner, and current turn.
// Useful for logging, debugging, and displaying game status to players.
func GetStateSummary(state State) string {
	var summary strings.Builder

	summary.WriteString(fmt.Sprintf("Round %d\n\n", state.Round))

	// Players
	summary.WriteString("Players:\n")
	for _, char := range state.Characters {
		if char.IsPlayer {
			status := fmt.Sprintf("%d/%d HP", char.Stats.HP, char.Stats.MaxHP)
			if char.Stats.HP <= 0 {
				status = "DEFEATED"
			}
			summary.WriteString(fmt.Sprintf("  %s: %s\n", char.Name, status))
		}
	}

	// Enemies
	summary.WriteString("\nEnemies:\n")
	for _, char := range state.Characters {
		if !char.IsPlayer {
			status := fmt.Sprintf("%d/%d HP", char.Stats.HP, char.Stats.MaxHP)
			if char.Stats.HP <= 0 {
				status = "DEFEATED"
			}
			summary.WriteString(fmt.Sprintf("  %s: %s\n", char.Name, status))
		}
	}

	if state.IsComplete {
		winner := "Draw"
		if state.Winner != nil {
			winner = strings.Title(*state.Winner)
		}
		summary.WriteString(fmt.Sprintf("\nCombat Complete! Winner: %s", winner))
	} else {
		currentChar := GetCurrentCharacter(state)
		currentName := "Unknown"
		if currentChar != nil {
			currentName = currentChar.Name
		}
		summary.WriteString(fmt.Sprintf("\nCurrent Turn: %s", currentName))
	}

	return summary.String()
}

// ApplyAction applies an action to the state and returns the resolution.
// Uses direct state mutation for deterministic game logic with seeded RNG.
func ApplyAction(state State, action Action, seed int64) Resolution {
	rng := NewSeededRNG(seed)
	events := []Event{}
	logs := []string{}

	newState := deepCopyState(state)
	character := GetCharacterByID(newState, getActorID(action))

	if character == nil {
		return Resolution{
			Events: events,
			State:  state,
			Logs:   append(logs, "Invalid action: character not found"),
		}
	}

	// Validate action kind
	validKinds := []string{"Attack", "Defend", "Ability", "UseItem", "Flee"}
	valid := false
	for _, k := range validKinds {
		if action.Kind == k {
			valid = true
			break
		}
	}
	if !valid {
		return Resolution{
			Events: events,
			State:  state,
			Logs:   append(logs, "Invalid action kind"),
		}
	}

	switch action.Kind {
	case "Attack":
		return handleAttack(&newState, action, rng, events, logs)
	case "Defend":
		return handleDefend(&newState, action, rng, events, logs)
	case "Ability":
		return handleAbility(&newState, action, rng, events, logs)
	case "UseItem":
		return handleUseItem(&newState, action, rng, events, logs)
	case "Flee":
		return handleFlee(&newState, action, rng, events, logs)
	default:
		return Resolution{
			Events: events,
			State:  state,
			Logs:   append(logs, "Unknown action kind"),
		}
	}
}

func getActorID(action Action) ID {
	switch action.Kind {
	case "Attack":
		return action.Attacker
	case "Defend", "Ability", "UseItem", "Flee":
		return action.Actor
	default:
		return ""
	}
}

func handleAttack(state *State, action Action, rng *SeededRNG, events []Event, logs []string) Resolution {
	attacker := GetCharacterByID(*state, action.Attacker)
	target := GetCharacterByID(*state, action.Target)

	if attacker == nil || target == nil {
		return Resolution{Events: events, State: *state, Logs: append(logs, "Invalid attack action")}
	}

	// Find weapon
	var weapon *Weapon
	for i := range attacker.Weapons {
		if attacker.Weapons[i].ID == action.Weapon {
			weapon = &attacker.Weapons[i]
			break
		}
	}

	if weapon == nil {
		logs = append(logs, "Weapon not found - using default")
		weapon = &Weapon{Name: "Fist", Damage: DefaultFistDamage, Accuracy: DefaultFistAccuracy}
	}

	attackRoll := rng.RollD20()
	hit := attackRoll+attacker.Stats.Attack >= target.Stats.Defense+AttackDC

	if hit {
		baseDamage := weapon.Damage + (attacker.Stats.Attack / AttackDamageDiv)
		damageRoll := rng.RollD6()
		totalDamage := int(math.Max(MinDamage, float64(baseDamage+damageRoll-target.Stats.Defense)))

		target.Stats.HP = int(math.Max(0, float64(target.Stats.HP-totalDamage)))

		events = append(events, Event{
			Type:   "damage",
			Target: target.ID,
			Amount: totalDamage,
			Source: attacker.ID,
		})

		logs = append(logs, fmt.Sprintf("%s attacks %s with %s for %d damage!", attacker.Name, target.Name, weapon.Name, totalDamage))

		if target.Stats.HP == 0 {
			events = append(events, Event{
				Type:   "death",
				Target: target.ID,
			})
			logs = append(logs, fmt.Sprintf("%s has been defeated!", target.Name))
		}
	} else {
		logs = append(logs, fmt.Sprintf("%s misses %s!", attacker.Name, target.Name))
	}

	updatedState := advanceTurn(*state)
	return Resolution{Events: events, State: updatedState, Logs: logs}
}

func handleDefend(state *State, action Action, rng *SeededRNG, events []Event, logs []string) Resolution {
	character := GetCharacterByID(*state, action.Actor)

	if character == nil {
		return Resolution{Events: events, State: *state, Logs: append(logs, "Invalid defend action")}
	}

	character.Stats.Defense += DefenseBonus
	logs = append(logs, fmt.Sprintf("%s takes a defensive stance!", character.Name))

	updatedState := advanceTurn(*state)
	return Resolution{Events: events, State: updatedState, Logs: logs}
}

func handleAbility(state *State, action Action, rng *SeededRNG, events []Event, logs []string) Resolution {
	character := GetCharacterByID(*state, action.Actor)

	if character == nil {
		return Resolution{Events: events, State: *state, Logs: append(logs, "Invalid ability action")}
	}

	// Find ability
	var ability *Ability
	for i := range character.Abilities {
		if character.Abilities[i].ID == action.Ability {
			ability = &character.Abilities[i]
			break
		}
	}

	if ability == nil {
		return Resolution{Events: events, State: *state, Logs: append(logs, "Ability not found")}
	}

	// Defensive: Initialize AbilityCooldowns map if nil to prevent panics
	if character.AbilityCooldowns == nil {
		character.AbilityCooldowns = make(map[string]int)
	}

	cooldownKey := string(ability.ID)
	currentCooldown := character.AbilityCooldowns[cooldownKey]

	if currentCooldown > 0 {
		return Resolution{Events: events, State: *state, Logs: append(logs, fmt.Sprintf("%s is on cooldown!", ability.Name))}
	}

	character.AbilityCooldowns[cooldownKey] = ability.Cooldown

	events = append(events, Event{
		Type:    "ability_used",
		Actor:   character.ID,
		Ability: ability.ID,
		Target:  action.Target,
	})

	switch ability.Effect {
	case "damage":
		if action.Target != "" {
			target := GetCharacterByID(*state, action.Target)
			if target != nil {
				damage := ability.Power + rng.RollD6()
				target.Stats.HP = int(math.Max(0, float64(target.Stats.HP-damage)))

				events = append(events, Event{
					Type:   "damage",
					Target: target.ID,
					Amount: damage,
					Source: character.ID,
				})

				logs = append(logs, fmt.Sprintf("%s uses %s on %s for %d damage!", character.Name, ability.Name, target.Name, damage))

				if target.Stats.HP == 0 {
					events = append(events, Event{
						Type:   "death",
						Target: target.ID,
					})
					logs = append(logs, fmt.Sprintf("%s has been defeated!", target.Name))
				}
			}
		}
	case "heal":
		healAmount := ability.Power + rng.RollD6()
		character.Stats.HP = int(math.Min(float64(character.Stats.MaxHP), float64(character.Stats.HP+healAmount)))

		events = append(events, Event{
			Type:   "heal",
			Target: character.ID,
			Amount: healAmount,
		})

		logs = append(logs, fmt.Sprintf("%s uses %s and heals for %d HP!", character.Name, ability.Name, healAmount))
	}

	updatedState := advanceTurn(*state)
	return Resolution{Events: events, State: updatedState, Logs: logs}
}

func handleUseItem(state *State, action Action, rng *SeededRNG, events []Event, logs []string) Resolution {
	character := GetCharacterByID(*state, action.Actor)

	if character == nil {
		return Resolution{Events: events, State: *state, Logs: append(logs, "Invalid item action")}
	}

	// Find and remove item
	itemIndex := -1
	var item *Item
	for i := range character.Items {
		if character.Items[i].ID == action.Item {
			itemIndex = i
			item = &character.Items[i]
			break
		}
	}

	if itemIndex == -1 || item == nil {
		return Resolution{Events: events, State: *state, Logs: append(logs, "Item not found")}
	}

	// Remove item from inventory
	character.Items = append(character.Items[:itemIndex], character.Items[itemIndex+1:]...)

	events = append(events, Event{
		Type:  "item_used",
		Actor: character.ID,
		Item:  item.ID,
	})

	if strings.Contains(item.Name, "Potion") {
		healAmount := 20 + rng.RollD6()
		character.Stats.HP = int(math.Min(float64(character.Stats.MaxHP), float64(character.Stats.HP+healAmount)))

		events = append(events, Event{
			Type:   "heal",
			Target: character.ID,
			Amount: healAmount,
		})

		logs = append(logs, fmt.Sprintf("%s uses %s and heals for %d HP!", character.Name, item.Name, healAmount))
	}

	updatedState := advanceTurn(*state)
	return Resolution{Events: events, State: updatedState, Logs: logs}
}

func handleFlee(state *State, action Action, rng *SeededRNG, events []Event, logs []string) Resolution {
	character := GetCharacterByID(*state, action.Actor)

	if character == nil {
		return Resolution{Events: events, State: *state, Logs: append(logs, "Invalid flee action")}
	}

	fleeRoll := rng.RollD20()
	success := fleeRoll+character.Stats.Speed >= FleeDC

	if success {
		events = append(events, Event{
			Type:  "flee",
			Actor: character.ID,
		})

		logs = append(logs, fmt.Sprintf("%s successfully flees from combat!", character.Name))

		winner := "player"
		updatedState := State{
			Round:       state.Round,
			Characters:  state.Characters,
			TurnOrder:   state.TurnOrder,
			CurrentTurn: state.CurrentTurn,
			IsComplete:  true,
			Winner:      &winner,
		}
		return Resolution{Events: events, State: updatedState, Logs: logs}
	} else {
		logs = append(logs, fmt.Sprintf("%s fails to flee!", character.Name))
		updatedState := advanceTurn(*state)
		return Resolution{Events: events, State: updatedState, Logs: logs}
	}
}

func advanceTurn(state State) State {
	updatedState := deepCopyState(state)

	// Decrease ability cooldowns
	for i := range updatedState.Characters {
		char := &updatedState.Characters[i]
		for abilityID, cooldown := range char.AbilityCooldowns {
			if cooldown > 0 {
				char.AbilityCooldowns[abilityID] = cooldown - 1
			}
		}

		// Reset defense bonus if it was increased
		if char.Stats.Defense > BaseDefenseMax {
			char.Stats.Defense = int(math.Max(0, float64(char.Stats.Defense-DefenseBonus)))
		}
	}

	// Check for combat end
	alivePlayers := 0
	aliveEnemies := 0
	for _, char := range updatedState.Characters {
		if char.IsPlayer && char.Stats.HP > 0 {
			alivePlayers++
		} else if !char.IsPlayer && char.Stats.HP > 0 {
			aliveEnemies++
		}
	}

	if alivePlayers == 0 {
		updatedState.IsComplete = true
		winner := "enemy"
		updatedState.Winner = &winner
	} else if aliveEnemies == 0 {
		updatedState.IsComplete = true
		winner := "player"
		updatedState.Winner = &winner
	}

	if !updatedState.IsComplete {
		updatedState.CurrentTurn = (updatedState.CurrentTurn + 1) % len(updatedState.TurnOrder)

		if updatedState.CurrentTurn == 0 {
			updatedState.Round++
		}
	}

	return updatedState
}

// deepCopyState creates a deep copy of the state using proper value copying
func deepCopyState(state State) State {
	// Copy characters slice
	characters := make([]Character, len(state.Characters))
	for i, char := range state.Characters {
		// Copy character with deep copies of slices and maps
		characters[i] = Character{
			ID:       char.ID,
			Name:     char.Name,
			Stats:    char.Stats, // Stat is a simple struct, value copy is fine
			Position: char.Position,
			IsPlayer: char.IsPlayer,
		}

		// Deep copy weapons
		if len(char.Weapons) > 0 {
			characters[i].Weapons = make([]Weapon, len(char.Weapons))
			copy(characters[i].Weapons, char.Weapons)
		}

		// Deep copy abilities
		if len(char.Abilities) > 0 {
			characters[i].Abilities = make([]Ability, len(char.Abilities))
			copy(characters[i].Abilities, char.Abilities)
		}

		// Deep copy items
		if len(char.Items) > 0 {
			characters[i].Items = make([]Item, len(char.Items))
			copy(characters[i].Items, char.Items)
		}

		// Deep copy ability cooldowns map - ALWAYS initialize to prevent nil map panics
		characters[i].AbilityCooldowns = make(map[string]int)
		for k, v := range char.AbilityCooldowns {
			characters[i].AbilityCooldowns[k] = v
		}
	}

	// Copy turn order
	turnOrder := make([]ID, len(state.TurnOrder))
	copy(turnOrder, state.TurnOrder)

	// Copy winner if present
	var winner *string
	if state.Winner != nil {
		w := *state.Winner
		winner = &w
	}

	newState := State{
		Round:       state.Round,
		Characters:  characters,
		TurnOrder:   turnOrder,
		CurrentTurn: state.CurrentTurn,
		IsComplete:  state.IsComplete,
		Winner:      winner,
	}

	// Rebuild character map for the new state
	newState.CharacterMap = buildCharacterMap(&newState)

	return newState
}

// CheckCombatEnd checks if combat should end
func CheckCombatEnd(state State) bool {
	return state.IsComplete || state.Round >= MaxCombatRounds
}
