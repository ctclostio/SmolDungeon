package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"sort"
	"strings"
)

// CreateInitialState creates the initial game state
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

	return State{
		Round:       1,
		Characters:  allCharacters,
		TurnOrder:   turnOrder,
		CurrentTurn: 0,
		IsComplete:  false,
	}
}

// GetCurrentCharacter returns the character whose turn it is
func GetCurrentCharacter(state State) *Character {
	if state.CurrentTurn >= len(state.TurnOrder) {
		return nil
	}
	currentID := state.TurnOrder[state.CurrentTurn]
	for i := range state.Characters {
		if state.Characters[i].ID == currentID {
			return &state.Characters[i]
		}
	}
	return nil
}

// GetCharacterByID finds a character by ID
func GetCharacterByID(state State, id ID) *Character {
	for i := range state.Characters {
		if state.Characters[i].ID == id {
			return &state.Characters[i]
		}
	}
	return nil
}

// GetStateSummary returns a string summary of the state
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

// ApplyAction applies an action to the state and returns the resolution
func ApplyAction(state State, action Action, seed int64) Resolution {
	// Stub for actor system: In full impl, send action as message to character actor goroutine
	// For now, log bypass and proceed with direct (to highlight violation)
	log.Printf("WARNING: Bypassing actor system for action %s", action.Kind)
	rng := NewSeededRNG(seed)
	events := []Event{}
	logs := []string{}
	logs = append(logs, "Actor bypass: Direct mutation used")

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
		weapon = &Weapon{Name: "Fist", Damage: 1, Accuracy: 0} // Default
	}

	attackRoll := rng.RollD20()
	hit := attackRoll+attacker.Stats.Attack >= target.Stats.Defense+10

	if hit {
		baseDamage := weapon.Damage + (attacker.Stats.Attack / 2)
		damageRoll := rng.RollD6()
		totalDamage := int(math.Max(1, float64(baseDamage+damageRoll-target.Stats.Defense)))

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

	character.Stats.Defense += 2
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
	success := fleeRoll+character.Stats.Speed >= 15

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

		// Reset defense bonus if it was increased (assuming base defense is around 3-5)
		if char.Stats.Defense > 5 {
			char.Stats.Defense = int(math.Max(0, float64(char.Stats.Defense-2)))
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

// deepCopyState creates a deep copy of the state
func deepCopyState(state State) State {
	stateBytes, err := json.Marshal(state)
	if err != nil {
		log.Printf("Deep copy failed: %v", err)
		return state // Fallback to shallow
	}
	var newState State
	if err := json.Unmarshal(stateBytes, &newState); err != nil {
		log.Printf("Deep copy failed: %v", err)
		return state // Fallback to shallow
	}
	return newState
}

// CheckCombatEnd checks if combat should end
func CheckCombatEnd(state State) bool {
	return state.IsComplete || state.Round >= 20
}
