package main

import (
	"testing"
)

func TestAbilityDamage(t *testing.T) {
	// Create characters with abilities
	player := Character{
		ID:   NewID(),
		Name: "Warrior",
		Stats: Stat{
			HP: 50, MaxHP: 50, Attack: 8, Defense: 3, Speed: 5,
		},
		Position: Position{X: 0, Y: 0},
		Abilities: []Ability{
			{ID: NewID(), Name: "Power Attack", Cooldown: 3, Effect: "damage", Power: 15},
		},
		AbilityCooldowns: make(map[string]int),
		IsPlayer:         true,
	}

	enemy := Character{
		ID:   NewID(),
		Name: "Goblin",
		Stats: Stat{
			HP: 30, MaxHP: 30, Attack: 5, Defense: 2, Speed: 4,
		},
		Position:         Position{X: 1, Y: 1},
		Abilities:        []Ability{},
		AbilityCooldowns: make(map[string]int),
		IsPlayer:         false,
	}

	state := CreateInitialState([]Character{player}, []Character{enemy}, 12345)

	// Use damage ability
	action := Action{
		Kind:    "Ability",
		Actor:   player.ID,
		Target:  enemy.ID,
		Ability: player.Abilities[0].ID,
	}

	resolution := ApplyAction(state, action, 12345)

	// Verify ability worked
	if len(resolution.Events) == 0 {
		t.Fatal("Expected events from ability use")
	}

	// Check for ability_used event
	hasAbilityEvent := false
	for _, event := range resolution.Events {
		if event.Type == "ability_used" {
			hasAbilityEvent = true
			break
		}
	}
	if !hasAbilityEvent {
		t.Error("Expected ability_used event")
	}

	// Verify cooldown was set (cooldown decrements by 1 when turn advances)
	updatedPlayer := GetCharacterByID(resolution.State, player.ID)
	if updatedPlayer == nil {
		t.Fatal("Player not found in updated state")
	}

	cooldownKey := string(player.Abilities[0].ID)
	expectedCooldown := player.Abilities[0].Cooldown - 1 // Decrements when turn advances
	if updatedPlayer.AbilityCooldowns[cooldownKey] != expectedCooldown {
		t.Errorf("Expected cooldown to be %d, got %d", expectedCooldown, updatedPlayer.AbilityCooldowns[cooldownKey])
	}

	// Verify damage was dealt
	updatedEnemy := GetCharacterByID(resolution.State, enemy.ID)
	if updatedEnemy == nil {
		t.Fatal("Enemy not found in updated state")
	}
	if updatedEnemy.Stats.HP >= enemy.Stats.HP {
		t.Error("Expected enemy to take damage")
	}

	t.Logf("✅ Damage ability test passed! Enemy HP: %d/%d, Cooldown: %d",
		updatedEnemy.Stats.HP, enemy.Stats.HP, updatedPlayer.AbilityCooldowns[cooldownKey])
}

func TestAbilityHeal(t *testing.T) {
	// Create character with healing ability
	player := Character{
		ID:   NewID(),
		Name: "Cleric",
		Stats: Stat{
			HP: 20, MaxHP: 50, Attack: 4, Defense: 3, Speed: 5,
		},
		Position: Position{X: 0, Y: 0},
		Abilities: []Ability{
			{ID: NewID(), Name: "Second Wind", Cooldown: 2, Effect: "heal", Power: 15},
		},
		AbilityCooldowns: make(map[string]int),
		IsPlayer:         true,
	}

	enemy := Character{
		ID:   NewID(),
		Name: "Goblin",
		Stats: Stat{
			HP: 30, MaxHP: 30, Attack: 5, Defense: 2, Speed: 4,
		},
		Position:         Position{X: 1, Y: 1},
		Abilities:        []Ability{},
		AbilityCooldowns: make(map[string]int),
		IsPlayer:         false,
	}

	state := CreateInitialState([]Character{player}, []Character{enemy}, 12345)

	// Use healing ability
	action := Action{
		Kind:    "Ability",
		Actor:   player.ID,
		Ability: player.Abilities[0].ID,
	}

	resolution := ApplyAction(state, action, 12345)

	// Verify healing occurred
	updatedPlayer := GetCharacterByID(resolution.State, player.ID)
	if updatedPlayer == nil {
		t.Fatal("Player not found in updated state")
	}

	if updatedPlayer.Stats.HP <= 20 {
		t.Errorf("Expected HP to increase from 20, got %d", updatedPlayer.Stats.HP)
	}

	// Verify cooldown was set (cooldown decrements by 1 when turn advances)
	cooldownKey := string(player.Abilities[0].ID)
	expectedCooldown := player.Abilities[0].Cooldown - 1 // Decrements when turn advances
	if updatedPlayer.AbilityCooldowns[cooldownKey] != expectedCooldown {
		t.Errorf("Expected cooldown to be %d, got %d", expectedCooldown, updatedPlayer.AbilityCooldowns[cooldownKey])
	}

	t.Logf("✅ Healing ability test passed! HP: 20→%d, Cooldown: %d",
		updatedPlayer.Stats.HP, updatedPlayer.AbilityCooldowns[cooldownKey])
}

func TestAbilityCooldownPreventsUse(t *testing.T) {
	// Create character with ability on cooldown
	player := Character{
		ID:   NewID(),
		Name: "Warrior",
		Stats: Stat{
			HP: 50, MaxHP: 50, Attack: 8, Defense: 3, Speed: 5,
		},
		Position: Position{X: 0, Y: 0},
		Abilities: []Ability{
			{ID: NewID(), Name: "Power Attack", Cooldown: 3, Effect: "damage", Power: 15},
		},
		AbilityCooldowns: make(map[string]int),
		IsPlayer:         true,
	}

	// Set ability on cooldown
	player.AbilityCooldowns[string(player.Abilities[0].ID)] = 2

	enemy := Character{
		ID:   NewID(),
		Name: "Goblin",
		Stats: Stat{
			HP: 30, MaxHP: 30, Attack: 5, Defense: 2, Speed: 4,
		},
		Position:         Position{X: 1, Y: 1},
		Abilities:        []Ability{},
		AbilityCooldowns: make(map[string]int),
		IsPlayer:         false,
	}

	state := CreateInitialState([]Character{player}, []Character{enemy}, 12345)

	// Try to use ability while on cooldown
	action := Action{
		Kind:    "Ability",
		Actor:   player.ID,
		Target:  enemy.ID,
		Ability: player.Abilities[0].ID,
	}

	resolution := ApplyAction(state, action, 12345)

	// Verify ability was blocked
	if len(resolution.Logs) == 0 {
		t.Fatal("Expected log message about cooldown")
	}

	foundCooldownMsg := false
	for _, log := range resolution.Logs {
		if len(log) >= 8 && log[len(log)-8:] == "cooldown" {
			foundCooldownMsg = true
			break
		}
		// Also check for "on cooldown!"
		if len(log) >= 12 {
			for i := 0; i <= len(log)-8; i++ {
				if log[i:i+8] == "cooldown" {
					foundCooldownMsg = true
					break
				}
			}
		}
	}

	if !foundCooldownMsg {
		t.Errorf("Expected cooldown message in logs, got: %v", resolution.Logs)
	}

	// Verify enemy was not damaged
	updatedEnemy := GetCharacterByID(resolution.State, enemy.ID)
	if updatedEnemy.Stats.HP != enemy.Stats.HP {
		t.Error("Enemy should not have taken damage while ability is on cooldown")
	}

	t.Logf("✅ Cooldown prevention test passed! Ability correctly blocked")
}

func TestAbilityCooldownDecreases(t *testing.T) {
	// Create character with ability on cooldown
	player := Character{
		ID:   NewID(),
		Name: "Warrior",
		Stats: Stat{
			HP: 50, MaxHP: 50, Attack: 8, Defense: 3, Speed: 5,
		},
		Position: Position{X: 0, Y: 0},
		Abilities: []Ability{
			{ID: NewID(), Name: "Power Attack", Cooldown: 3, Effect: "damage", Power: 15},
		},
		AbilityCooldowns: make(map[string]int),
		IsPlayer:         true,
	}

	// Set ability on cooldown
	abilityKey := string(player.Abilities[0].ID)
	player.AbilityCooldowns[abilityKey] = 3

	enemy := Character{
		ID:   NewID(),
		Name: "Goblin",
		Stats: Stat{
			HP: 30, MaxHP: 30, Attack: 5, Defense: 2, Speed: 4,
		},
		Position:         Position{X: 1, Y: 1},
		Abilities:        []Ability{},
		AbilityCooldowns: make(map[string]int),
		IsPlayer:         false,
	}

	state := CreateInitialState([]Character{player}, []Character{enemy}, 12345)

	// Perform a different action (defend) to advance turn
	action := Action{
		Kind:  "Defend",
		Actor: player.ID,
	}

	resolution := ApplyAction(state, action, 12345)

	// Verify cooldown decreased
	updatedPlayer := GetCharacterByID(resolution.State, player.ID)
	if updatedPlayer == nil {
		t.Fatal("Player not found in updated state")
	}

	if updatedPlayer.AbilityCooldowns[abilityKey] != 2 {
		t.Errorf("Expected cooldown to decrease to 2, got %d", updatedPlayer.AbilityCooldowns[abilityKey])
	}

	t.Logf("✅ Cooldown decrease test passed! Cooldown: 3→%d", updatedPlayer.AbilityCooldowns[abilityKey])
}
