package main

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"
)

// TestGoblinAmbushBalance runs multiple simulations to analyze balance
func TestGoblinAmbushBalance(t *testing.T) {
	// Load the goblin-ambush scenario
	scenarioPath := filepath.Join("..", "..", "scenarios", "goblin-ambush.yaml")
	scenario, err := loadScenario(scenarioPath)
	if err != nil {
		t.Fatalf("Failed to load scenario: %v", err)
	}

	t.Logf("=== GOBLIN AMBUSH BALANCE ANALYSIS ===\n")
	t.Logf("Scenario: %s\n", scenario.Name)
	t.Logf("Description: %s\n\n", scenario.Description)

	// Print initial stats
	t.Logf("=== INITIAL STATS ===")
	for _, player := range scenario.Players {
		t.Logf("PLAYER: %s", player.Name)
		t.Logf("  HP: %d/%d", player.Stats.HP, player.Stats.MaxHP)
		t.Logf("  Attack: %d, Defense: %d, Speed: %d", player.Stats.Attack, player.Stats.Defense, player.Stats.Speed)
		for _, weapon := range player.Weapons {
			t.Logf("  Weapon: %s (Damage: %d, Accuracy: %d%%)", weapon.Name, weapon.Damage, weapon.Accuracy)
		}
		for _, ability := range player.Abilities {
			t.Logf("  Ability: %s (Power: %d, Cooldown: %d)", ability.Name, ability.Power, ability.Cooldown)
		}
	}
	t.Logf("")

	for _, enemy := range scenario.Enemies {
		t.Logf("ENEMY: %s", enemy.Name)
		t.Logf("  HP: %d/%d", enemy.Stats.HP, enemy.Stats.MaxHP)
		t.Logf("  Attack: %d, Defense: %d, Speed: %d", enemy.Stats.Attack, enemy.Stats.Defense, enemy.Stats.Speed)
		for _, weapon := range enemy.Weapons {
			t.Logf("  Weapon: %s (Damage: %d, Accuracy: %d%%)", weapon.Name, weapon.Damage, weapon.Accuracy)
		}
		for _, ability := range enemy.Abilities {
			t.Logf("  Ability: %s (Power: %d, Cooldown: %d)", ability.Name, ability.Power, ability.Cooldown)
		}
	}
	t.Logf("")

	// Run simulations
	numSimulations := 50
	playerWins := 0
	enemyWins := 0
	totalDamageTaken := 0
	minHPRemaining := 999
	maxHPRemaining := 0

	t.Logf("=== RUNNING %d SIMULATIONS ===\n", numSimulations)

	for i := 0; i < numSimulations; i++ {
		seed := int64(12345 + i*1000)
		winner, rounds, summary, fighterHP := simulateBattle(scenario, seed)

		if i < 10 { // Only log first 10 for brevity
			t.Logf("--- Simulation %d (seed: %d) ---", i+1, seed)
			t.Logf("%s", summary)
			t.Logf("Winner: %s in %d rounds\n", winner, rounds)
		}

		if winner == "player" {
			playerWins++
			totalDamageTaken += (scenario.Players[0].Stats.MaxHP - fighterHP)
			if fighterHP < minHPRemaining {
				minHPRemaining = fighterHP
			}
			if fighterHP > maxHPRemaining {
				maxHPRemaining = fighterHP
			}
		} else if winner == "enemy" {
			enemyWins++
		}
	}

	// Print summary
	t.Logf("\n=== SIMULATION SUMMARY ===")
	t.Logf("Total Simulations: %d", numSimulations)
	t.Logf("Player Wins: %d (%.1f%%)", playerWins, float64(playerWins)/float64(numSimulations)*100)
	t.Logf("Enemy Wins: %d (%.1f%%)", enemyWins, float64(enemyWins)/float64(numSimulations)*100)
	if playerWins > 0 {
		avgDamage := float64(totalDamageTaken) / float64(playerWins)
		t.Logf("\nPlayer Victory Stats:")
		t.Logf("  Average damage taken: %.1f", avgDamage)
		t.Logf("  HP remaining range: %d-%d", minHPRemaining, maxHPRemaining)
		t.Logf("  Closest call: %d HP remaining", minHPRemaining)
	}

	// Balance assessment
	t.Logf("\n=== BALANCE ASSESSMENT ===")
	winRate := float64(playerWins) / float64(numSimulations) * 100
	if winRate < 30 {
		t.Logf("❌ SEVERELY UNBALANCED - Player win rate too low (%.1f%%)", winRate)
		t.Logf("   Recommendation: Major buffs needed (HP, damage, or reduce enemies)")
	} else if winRate < 50 {
		t.Logf("⚠️  UNBALANCED - Player win rate below 50%% (%.1f%%)", winRate)
		t.Logf("   Recommendation: Moderate buffs needed")
	} else if winRate < 70 {
		t.Logf("✅ BALANCED - Fair challenge (%.1f%% win rate)", winRate)
	} else if winRate < 85 {
		t.Logf("✅ SLIGHTLY EASY - Good for tutorial (%.1f%% win rate)", winRate)
	} else {
		t.Logf("⚠️  TOO EASY - Player wins almost always (%.1f%%)", winRate)
		t.Logf("   Recommendation: Increase difficulty")
	}
}

// simulateBattle runs a full battle simulation
func simulateBattle(scenario *Scenario, seed int64) (winner string, rounds int, summary string, fighterHP int) {
	state := ConvertScenarioToState(scenario, seed)
	rng := NewSeededRNG(seed)

	turnCount := 0
	maxTurns := 100 // Safety limit

	// Track starting HP
	startingHP := make(map[ID]int)
	for _, char := range state.Characters {
		startingHP[char.ID] = char.Stats.HP
	}

	for !state.IsComplete && turnCount < maxTurns {
		currentChar := GetCurrentCharacter(state)
		if currentChar == nil || currentChar.Stats.HP <= 0 {
			// Skip dead characters
			state = advanceTurn(state)
			continue
		}

		// Simple AI: attack the first alive enemy/player
		var action Action
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
					break
				}
			}
		}

		// Apply action
		actionSeed := rng.RollD20() + int(seed)
		resolution := ApplyAction(state, action, int64(actionSeed))
		state = resolution.State
		turnCount++
	}

	// Calculate summary
	var summaryText string
	fighterFinalHP := 0
	for _, char := range state.Characters {
		damageTaken := startingHP[char.ID] - char.Stats.HP
		summaryText += fmt.Sprintf("  %s: %d HP → %d HP (took %d damage)\n",
			char.Name, startingHP[char.ID], char.Stats.HP, damageTaken)
		if char.IsPlayer {
			fighterFinalHP = char.Stats.HP
		}
	}

	if state.Winner != nil {
		winner = *state.Winner
	} else {
		winner = "draw"
	}

	return winner, state.Round, summaryText, fighterFinalHP
}

// loadScenario loads a scenario from a YAML file
func loadScenario(path string) (*Scenario, error) {
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
