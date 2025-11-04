package engine

import (
	"fmt"
	"os"
	"sort"

	"gopkg.in/yaml.v3"
)

// ScenarioCharacter represents a character in a scenario file
type ScenarioCharacter struct {
	Name      string                `yaml:"name"`
	Stats     ScenarioStat          `yaml:"stats"`
	Position  Position              `yaml:"position"`
	Weapons   []Weapon              `yaml:"weapons,omitempty"`
	Abilities []Ability             `yaml:"abilities,omitempty"`
	Items     []Item                `yaml:"items,omitempty"`
}

// ScenarioStat represents stats in a scenario file
type ScenarioStat struct {
	HP      int `yaml:"hp"`
	MaxHP   int `yaml:"maxHp"`
	Attack  int `yaml:"attack"`
	Defense int `yaml:"defense"`
	Speed   int `yaml:"speed"`
}

// Scenario represents a game scenario
type Scenario struct {
	Name        string              `yaml:"name"`
	Description string              `yaml:"description"`
	Players     []ScenarioCharacter `yaml:"players"`
	Enemies     []ScenarioCharacter `yaml:"enemies"`
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

	// Use CreateInitialState from core.go
	return CreateInitialState(players, enemies, seed)
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
	char.Position = sc.Position

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

// ListScenarios returns available scenario names from a directory
func ListScenarios(scenarioDir string) ([]string, error) {
	entries, err := os.ReadDir(scenarioDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read scenarios directory: %w", err)
	}

	var scenarios []string
	for _, entry := range entries {
		if !entry.IsDir() && len(entry.Name()) > 5 && entry.Name()[len(entry.Name())-5:] == ".yaml" {
			scenarios = append(scenarios, entry.Name()[:len(entry.Name())-5])
		}
	}

	sort.Strings(scenarios)
	return scenarios, nil
}
