package main

import (
	"fmt"
	"os"
	"sort"

	"gopkg.in/yaml.v3"
)

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
