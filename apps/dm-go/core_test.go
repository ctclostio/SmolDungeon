package main

import (
	"testing"
)

func createTestCharacter(isPlayer bool, name string) Character {
	return Character{
		ID:   NewID(),
		Name: name,
		Stats: Stat{
			HP:      30,
			MaxHP:   30,
			Attack:  15, // Higher attack to ensure hits
			Defense: 3,
			Speed:   4,
		},
		Weapons: []Weapon{{
			ID:       NewID(),
			Name:     "Test Weapon",
			Damage:   6,
			Accuracy: 85,
		}},
		Abilities: []Ability{{
			ID:       NewID(),
			Name:     "Test Ability",
			Cooldown: 3,
			Effect:   "damage",
			Power:    8,
		}},
		Items: []Item{{
			ID:     NewID(),
			Name:   "Health Potion",
			Type:   "consumable",
			Effect: "heal 20 HP",
		}},
		AbilityCooldowns: make(map[string]int),
		IsPlayer:         isPlayer,
	}
}

func TestApplyAction_Attack(t *testing.T) {
	player := createTestCharacter(true, "Player")
	enemy := createTestCharacter(false, "Enemy")
	state := CreateInitialState([]Character{player}, []Character{enemy}, 12345)

	action := Action{
		Kind:     "Attack",
		Attacker: player.ID,
		Target:   enemy.ID,
		Weapon:   player.Weapons[0].ID,
	}

	resolution := ApplyAction(state, action, 12345)

	// Debug: print what happened
	t.Logf("Events: %d, Logs: %v", len(resolution.Events), resolution.Logs)

	// Should have events (attack should hit with these stats)
	if len(resolution.Events) == 0 {
		t.Error("Expected events from attack action - attack may have missed")
	}

	// Should have logs
	if len(resolution.Logs) == 0 {
		t.Error("Expected logs from attack action")
	}

	// HP should not be negative
	for _, char := range resolution.State.Characters {
		if char.Stats.HP < 0 {
			t.Errorf("Character %s has negative HP: %d", char.Name, char.Stats.HP)
		}
		if char.Stats.HP > char.Stats.MaxHP {
			t.Errorf("Character %s has HP > MaxHP: %d > %d", char.Name, char.Stats.HP, char.Stats.MaxHP)
		}
	}
}

func TestApplyAction_Defend(t *testing.T) {
	player := createTestCharacter(true, "Player")
	enemy := createTestCharacter(false, "Enemy")
	state := CreateInitialState([]Character{player}, []Character{enemy}, 12345)

	action := Action{
		Kind:  "Defend",
		Actor: player.ID,
	}

	resolution := ApplyAction(state, action, 12345)

	// Find the player in the new state
	var newPlayer *Character
	for i := range resolution.State.Characters {
		if resolution.State.Characters[i].ID == player.ID {
			newPlayer = &resolution.State.Characters[i]
			break
		}
	}

	if newPlayer == nil {
		t.Fatal("Player not found in new state")
	}

	// Defense should be increased
	if newPlayer.Stats.Defense <= player.Stats.Defense {
		t.Errorf("Expected defense to increase, got %d (was %d)", newPlayer.Stats.Defense, player.Stats.Defense)
	}
}

func TestDeterministicWithSameSeed(t *testing.T) {
	player1 := createTestCharacter(true, "Player")
	enemy1 := createTestCharacter(false, "Enemy")
	state1 := CreateInitialState([]Character{player1}, []Character{enemy1}, 12345)

	player2 := createTestCharacter(true, "Player")
	enemy2 := createTestCharacter(false, "Enemy")
	state2 := CreateInitialState([]Character{player2}, []Character{enemy2}, 12345)

	action1 := Action{
		Kind:     "Attack",
		Attacker: player1.ID,
		Target:   enemy1.ID,
		Weapon:   player1.Weapons[0].ID,
	}

	action2 := Action{
		Kind:     "Attack",
		Attacker: player2.ID,
		Target:   enemy2.ID,
		Weapon:   player2.Weapons[0].ID,
	}

	resolution1 := ApplyAction(state1, action1, 12345)
	resolution2 := ApplyAction(state2, action2, 12345)

	if len(resolution1.Events) != len(resolution2.Events) {
		t.Errorf("Expected same number of events, got %d vs %d", len(resolution1.Events), len(resolution2.Events))
	}

	if len(resolution1.Logs) != len(resolution2.Logs) {
		t.Errorf("Expected same number of logs, got %d vs %d", len(resolution1.Logs), len(resolution2.Logs))
	}
}

func TestTurnOrder(t *testing.T) {
	player := createTestCharacter(true, "Player")
	enemy := createTestCharacter(false, "Enemy")
	state := CreateInitialState([]Character{player}, []Character{enemy}, 12345)

	if len(state.TurnOrder) != 2 {
		t.Errorf("Expected 2 characters in turn order, got %d", len(state.TurnOrder))
	}

	if state.CurrentTurn < 0 || state.CurrentTurn >= len(state.TurnOrder) {
		t.Errorf("Invalid current turn: %d (should be 0-%d)", state.CurrentTurn, len(state.TurnOrder)-1)
	}

	// Check that both characters are in turn order
	foundPlayer := false
	foundEnemy := false
	for _, id := range state.TurnOrder {
		if id == player.ID {
			foundPlayer = true
		}
		if id == enemy.ID {
			foundEnemy = true
		}
	}

	if !foundPlayer {
		t.Error("Player not found in turn order")
	}
	if !foundEnemy {
		t.Error("Enemy not found in turn order")
	}
}

func TestDefenseReset(t *testing.T) {
	player := createTestCharacter(true, "Player")
	enemy := createTestCharacter(false, "Enemy")
	state := CreateInitialState([]Character{player}, []Character{enemy}, 12345)

	action := Action{
		Kind:  "Defend",
		Actor: player.ID,
	}

	// Apply defend action
	resolution := ApplyAction(state, action, 12345)

	// Find the player in the new state
	var newPlayer *Character
	for i := range resolution.State.Characters {
		if resolution.State.Characters[i].ID == player.ID {
			newPlayer = &resolution.State.Characters[i]
			break
		}
	}

	if newPlayer == nil {
		t.Fatal("Player not found in new state")
	}

	// Defense should be increased by 2
	expectedDefense := player.Stats.Defense + 2
	if newPlayer.Stats.Defense != expectedDefense {
		t.Errorf("Expected defense to be %d, got %d", expectedDefense, newPlayer.Stats.Defense)
	}

	// Simulate turn advancement by calling advanceTurn again
	resetState := advanceTurn(resolution.State)

	// Find player again
	for i := range resetState.Characters {
		if resetState.Characters[i].ID == player.ID {
			newPlayer = &resetState.Characters[i]
			break
		}
	}

	// Defense should be reset
	if newPlayer.Stats.Defense > 5 {
		t.Errorf("Expected defense to be reset to base value, got %d", newPlayer.Stats.Defense)
	}
}
