package main

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"
)

// TestDetailedCombatAnalysis provides turn-by-turn analysis
func TestDetailedCombatAnalysis(t *testing.T) {
	scenarioPath := filepath.Join("..", "..", "scenarios", "goblin-ambush.yaml")
	scenario, err := loadScenarioFromFile(scenarioPath)
	if err != nil {
		t.Fatalf("Failed to load scenario: %v", err)
	}

	// Test a few different seeds to see variation
	seeds := []int64{18345, 19345, 20345} // Mix of outcomes

	for _, seed := range seeds {
		t.Logf("\n======================================================================")
		t.Logf("DETAILED COMBAT ANALYSIS - Seed %d", seed)
		t.Logf("======================================================================\n")

		state := ConvertScenarioToState(scenario, seed)
		rng := NewSeededRNG(seed)

		// Show initiative order
		t.Logf("=== INITIATIVE ORDER ===")
		for i, charID := range state.TurnOrder {
			char := GetCharacterByID(state, charID)
			if char != nil {
				t.Logf("%d. %s (Speed: %d, %s)", i+1, char.Name, char.Stats.Speed,
					map[bool]string{true: "PLAYER", false: "ENEMY"}[char.IsPlayer])
			}
		}
		t.Logf("")

		turnNum := 0
		maxTurns := 50

		for !state.IsComplete && turnNum < maxTurns {
			currentChar := GetCurrentCharacter(state)
			if currentChar == nil {
				state = advanceTurn(state)
				continue
			}

			if currentChar.Stats.HP <= 0 {
				state = advanceTurn(state)
				continue
			}

			turnNum++
			t.Logf("--- Round %d, Turn %d ---", state.Round, turnNum)
			t.Logf("Current: %s (%d/%d HP)", currentChar.Name, currentChar.Stats.HP, currentChar.Stats.MaxHP)

			// Show all character status
			for _, char := range state.Characters {
				status := "ALIVE"
				if char.Stats.HP <= 0 {
					status = "DEAD"
				}
				t.Logf("  [%s] %s: %d/%d HP", status, char.Name, char.Stats.HP, char.Stats.MaxHP)
			}

			// Determine action
			var action Action
			var targetName string
			if currentChar.IsPlayer {
				// Find first alive enemy
				for _, char := range state.Characters {
					if !char.IsPlayer && char.Stats.HP > 0 {
						action = Action{
							Kind:     "Attack",
							Attacker: currentChar.ID,
							Target:   char.ID,
							Weapon:   currentChar.Weapons[0].ID,
						}
						targetName = char.Name
						break
					}
				}
			} else {
				// Find first alive player
				for _, char := range state.Characters {
					if char.IsPlayer && char.Stats.HP > 0 {
						action = Action{
							Kind:     "Attack",
							Attacker: currentChar.ID,
							Target:   char.ID,
							Weapon:   currentChar.Weapons[0].ID,
						}
						targetName = char.Name
						break
					}
				}
			}

			if action.Kind == "" {
				break
			}

			t.Logf("ACTION: %s attacks %s with %s", currentChar.Name, targetName, currentChar.Weapons[0].Name)

			// Apply action
			actionSeed := rng.RollD20() + int(seed)
			resolution := ApplyAction(state, action, int64(actionSeed))

			// Show logs
			for _, log := range resolution.Logs {
				t.Logf("  → %s", log)
			}

			state = resolution.State
			t.Logf("")

			if state.IsComplete {
				break
			}
		}

		// Final summary
		t.Logf("\n=== BATTLE COMPLETE ===")
		t.Logf("Winner: %s", *state.Winner)
		t.Logf("Rounds: %d, Total Turns: %d", state.Round, turnNum)
		t.Logf("\nFinal Status:")
		for _, char := range state.Characters {
			t.Logf("  %s: %d/%d HP", char.Name, char.Stats.HP, char.Stats.MaxHP)
		}
	}
}

// TestDamageCalculations shows expected damage values
func TestDamageCalculations(t *testing.T) {
	scenarioPath := filepath.Join("..", "..", "scenarios", "goblin-ambush.yaml")
	scenario, err := loadScenarioFromFile(scenarioPath)
	if err != nil {
		t.Fatalf("Failed to load scenario: %v", err)
	}

	t.Logf("\n=== DAMAGE CALCULATIONS ===\n")

	// Calculate expected damage for each attacker-target pair
	for _, attacker := range append(scenario.Players, scenario.Enemies...) {
		t.Logf("ATTACKER: %s (Attack: %d)", attacker.Name, attacker.Stats.Attack)

		for _, weapon := range attacker.Weapons {
			t.Logf("  Weapon: %s (Base Damage: %d)", weapon.Name, weapon.Damage)

			// Calculate against each possible target
			targets := append(scenario.Players, scenario.Enemies...)
			for _, target := range targets {
				if target.Name == attacker.Name {
					continue
				}

				// Damage formula from core.go:
				// baseDamage := weapon.Damage + (attacker.Stats.Attack / 2)
				// damageRoll := 1d6 (average 3.5)
				// totalDamage := baseDamage + damageRoll - target.Stats.Defense
				// (with minimum 1)

				baseDamage := weapon.Damage + (attacker.Stats.Attack / 2)
				minDamage := max(1, baseDamage + 1 - target.Stats.Defense) // Minimum roll (1)
				avgDamage := max(1, baseDamage + 3 - target.Stats.Defense) // Average roll (3.5, rounded down)
				maxDamageValue := max(1, baseDamage + 6 - target.Stats.Defense) // Maximum roll (6)

				t.Logf("    vs %s (Defense: %d): %d-%d damage (avg: %d)",
					target.Name, target.Stats.Defense, minDamage, maxDamageValue, avgDamage)

				// Calculate time to kill
				ttk := (target.Stats.HP + avgDamage - 1) / avgDamage
				t.Logf("      Time to kill: ~%d hits", ttk)
			}
		}

		// Calculate ability damage
		for _, ability := range attacker.Abilities {
			if ability.Effect == "damage" {
				t.Logf("  Ability: %s", ability.Name)
				for _, target := range append(scenario.Players, scenario.Enemies...) {
					if target.Name == attacker.Name {
						continue
					}

					// Ability damage: Power + 1d6 (no defense subtraction)
					minDmg := ability.Power + 1
					avgDmg := ability.Power + 3
					maxDmg := ability.Power + 6

					t.Logf("    vs %s: %d-%d damage (avg: %d)", target.Name, minDmg, maxDmg, avgDmg)
				}
			}
		}
		t.Logf("")
	}

	// Action economy analysis
	t.Logf("=== ACTION ECONOMY ANALYSIS ===")
	playerCount := len(scenario.Players)
	enemyCount := len(scenario.Enemies)
	t.Logf("Players: %d", playerCount)
	t.Logf("Enemies: %d", enemyCount)
	t.Logf("Action Ratio: %d player actions : %d enemy actions per round", playerCount, enemyCount)
	if enemyCount > playerCount {
		t.Logf("⚠️  WARNING: Enemies have action economy advantage (%d:%d)", enemyCount, playerCount)
	}

	// Speed analysis
	t.Logf("\n=== SPEED ANALYSIS ===")
	for _, player := range scenario.Players {
		t.Logf("Player %s: Speed %d", player.Name, player.Stats.Speed)
	}
	for _, enemy := range scenario.Enemies {
		t.Logf("Enemy %s: Speed %d", enemy.Name, enemy.Stats.Speed)
	}

	// Check if player is slower than all enemies
	playerSpeed := scenario.Players[0].Stats.Speed
	allEnemiesFaster := true
	for _, enemy := range scenario.Enemies {
		if enemy.Stats.Speed <= playerSpeed {
			allEnemiesFaster = false
		}
	}
	if allEnemiesFaster {
		t.Logf("⚠️  WARNING: All enemies are faster than player!")
		t.Logf("   Player will likely take damage before acting each round")
	}
}

func loadScenarioFromFile(path string) (*Scenario, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read scenario file: %w", err)
	}

	var scenario Scenario
	if err := yaml.Unmarshal(data, &scenario); err != nil {
		return nil, fmt.Errorf("failed to parse scenario YAML: %w", err)
	}

	return &scenario, nil
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
