package main

import (
	"testing"
)

// TestVictoryAllEnemiesDead verifies players win when all enemies are defeated
func TestVictoryAllEnemiesDead(t *testing.T) {
	player := createTestCharacter(true, "Player")
	player.Stats.Attack = 50 // Very high to ensure kill

	enemy := createTestCharacter(false, "Enemy")
	enemy.Stats.HP = 1
	enemy.Stats.MaxHP = 1
	enemy.Stats.Defense = 0

	state := CreateInitialState([]Character{player}, []Character{enemy}, 12345)

	action := Action{
		Kind:     "Attack",
		Attacker: player.ID,
		Target:   enemy.ID,
		Weapon:   player.Weapons[0].ID,
	}

	resolution := ApplyAction(state, action, 12345)

	// Combat should be complete
	if !resolution.State.IsComplete {
		t.Error("Combat should be complete when all enemies are dead")
	}

	// Player should win
	if resolution.State.Winner == nil {
		t.Fatal("Winner should be set")
	}
	if *resolution.State.Winner != "player" {
		t.Errorf("Expected player to win, got %s", *resolution.State.Winner)
	}
}

// TestDefeatAllPlayersDead verifies enemies win when all players are defeated
func TestDefeatAllPlayersDead(t *testing.T) {
	player := createTestCharacter(true, "Player")
	player.Stats.HP = 1
	player.Stats.MaxHP = 1
	player.Stats.Defense = 0

	enemy := createTestCharacter(false, "Enemy")
	enemy.Stats.Attack = 50 // Very high to ensure kill

	state := CreateInitialState([]Character{player}, []Character{enemy}, 12345)

	// Ensure enemy goes first by manipulating turn order
	state.TurnOrder = []ID{enemy.ID, player.ID}
	state.CurrentTurn = 0
	state.CharacterMap = buildCharacterMap(&state)

	action := Action{
		Kind:     "Attack",
		Attacker: enemy.ID,
		Target:   player.ID,
		Weapon:   enemy.Weapons[0].ID,
	}

	resolution := ApplyAction(state, action, 12345)

	// Combat should be complete
	if !resolution.State.IsComplete {
		t.Error("Combat should be complete when all players are dead")
	}

	// Enemy should win
	if resolution.State.Winner == nil {
		t.Fatal("Winner should be set")
	}
	if *resolution.State.Winner != "enemy" {
		t.Errorf("Expected enemy to win, got %s", *resolution.State.Winner)
	}
}

// TestMultiplePlayerVictory verifies victory with multiple players
func TestMultiplePlayerVictory(t *testing.T) {
	player1 := createTestCharacter(true, "Player1")
	player2 := createTestCharacter(true, "Player2")
	player1.Stats.Attack = 50
	player2.Stats.Attack = 50

	enemy := createTestCharacter(false, "Enemy")
	enemy.Stats.HP = 1
	enemy.Stats.MaxHP = 1

	state := CreateInitialState([]Character{player1, player2}, []Character{enemy}, 12345)

	action := Action{
		Kind:     "Attack",
		Attacker: player1.ID,
		Target:   enemy.ID,
		Weapon:   player1.Weapons[0].ID,
	}

	resolution := ApplyAction(state, action, 12345)

	// Combat should end with player victory
	if !resolution.State.IsComplete {
		t.Error("Combat should be complete")
	}

	if resolution.State.Winner == nil || *resolution.State.Winner != "player" {
		t.Error("Players should win")
	}
}

// TestMultipleEnemyVictory verifies defeat with multiple enemies
func TestMultipleEnemyVictory(t *testing.T) {
	player := createTestCharacter(true, "Player")
	player.Stats.HP = 1
	player.Stats.MaxHP = 1

	enemy1 := createTestCharacter(false, "Enemy1")
	enemy2 := createTestCharacter(false, "Enemy2")
	enemy1.Stats.Attack = 50
	enemy2.Stats.Attack = 50

	state := CreateInitialState([]Character{player}, []Character{enemy1, enemy2}, 12345)

	// Ensure enemy goes first
	state.TurnOrder = []ID{enemy1.ID, player.ID, enemy2.ID}
	state.CurrentTurn = 0
	state.CharacterMap = buildCharacterMap(&state)

	action := Action{
		Kind:     "Attack",
		Attacker: enemy1.ID,
		Target:   player.ID,
		Weapon:   enemy1.Weapons[0].ID,
	}

	resolution := ApplyAction(state, action, 12345)

	// Combat should end with enemy victory
	if !resolution.State.IsComplete {
		t.Error("Combat should be complete")
	}

	if resolution.State.Winner == nil || *resolution.State.Winner != "enemy" {
		t.Error("Enemies should win")
	}
}

// TestPartialVictory verifies game continues with survivors on both sides
func TestPartialVictory(t *testing.T) {
	player1 := createTestCharacter(true, "Player1")
	player2 := createTestCharacter(true, "Player2")
	player2.Stats.HP = 1

	enemy1 := createTestCharacter(false, "Enemy1")
	enemy2 := createTestCharacter(false, "Enemy2")

	state := CreateInitialState([]Character{player1, player2}, []Character{enemy1, enemy2}, 12345)

	// Kill one player
	for i := range state.Characters {
		if state.Characters[i].ID == player2.ID {
			state.Characters[i].Stats.HP = 0
		}
	}
	state.CharacterMap = buildCharacterMap(&state)

	// Game should continue (1 player and 2 enemies alive)
	newState := advanceTurn(state)

	if newState.IsComplete {
		t.Error("Game should not be complete with survivors on both sides")
	}

	if newState.Winner != nil {
		t.Error("Should not have a winner yet")
	}
}

// TestLastPlayerStanding verifies victory when only one player survives
func TestLastPlayerStanding(t *testing.T) {
	player1 := createTestCharacter(true, "Player1")
	player2 := createTestCharacter(true, "Player2")
	player2.Stats.HP = 0 // Dead

	enemy := createTestCharacter(false, "Enemy")
	enemy.Stats.HP = 1

	state := CreateInitialState([]Character{player1, player2}, []Character{enemy}, 12345)

	// Set player2 as dead
	for i := range state.Characters {
		if state.Characters[i].ID == player2.ID {
			state.Characters[i].Stats.HP = 0
		}
	}
	state.CharacterMap = buildCharacterMap(&state)

	// Player1 kills enemy
	action := Action{
		Kind:     "Attack",
		Attacker: player1.ID,
		Target:   enemy.ID,
		Weapon:   player1.Weapons[0].ID,
	}

	// Set high attack to guarantee kill
	for i := range state.Characters {
		if state.Characters[i].ID == player1.ID {
			state.Characters[i].Stats.Attack = 50
		}
	}
	state.CharacterMap = buildCharacterMap(&state)

	resolution := ApplyAction(state, action, 12345)

	// Should win
	if !resolution.State.IsComplete {
		t.Error("Combat should be complete")
	}

	if resolution.State.Winner == nil || *resolution.State.Winner != "player" {
		t.Error("Player should win even with one player dead")
	}
}

// TestSimultaneousDeath verifies handling when last characters on both sides die
func TestSimultaneousDeath(t *testing.T) {
	// This is an edge case - in practice, turns are sequential
	// but let's test the state check
	player := createTestCharacter(true, "Player")
	player.Stats.HP = 0

	enemy := createTestCharacter(false, "Enemy")
	enemy.Stats.HP = 0

	state := State{
		Round:      1,
		Characters: []Character{player, enemy},
		TurnOrder:  []ID{player.ID, enemy.ID},
	}

	// Set all HP to 0
	for i := range state.Characters {
		state.Characters[i].Stats.HP = 0
	}
	state.CharacterMap = buildCharacterMap(&state)

	// Check victory condition
	newState := advanceTurn(state)

	// With all players dead, enemy wins
	if !newState.IsComplete {
		t.Error("Game should end when all characters are dead")
	}

	if newState.Winner == nil {
		t.Error("Should have a winner")
	}

	// Players checked first, so enemy wins when all are dead
	if *newState.Winner != "enemy" {
		t.Errorf("Expected enemy to win, got %s", *newState.Winner)
	}
}

// TestVictoryStateImmutable verifies completed games don't change
func TestVictoryStateImmutable(t *testing.T) {
	player := createTestCharacter(true, "Player")
	enemy := createTestCharacter(false, "Enemy")

	state := State{
		Round:      1,
		Characters: []Character{player, enemy},
		TurnOrder:  []ID{player.ID, enemy.ID},
		IsComplete: true,
		Winner:     stringPtr("player"),
	}
	state.CharacterMap = buildCharacterMap(&state)

	action := Action{
		Kind:     "Attack",
		Attacker: player.ID,
		Target:   enemy.ID,
		Weapon:   player.Weapons[0].ID,
	}

	resolution := ApplyAction(state, action, 12345)

	// Game should remain complete
	if !resolution.State.IsComplete {
		t.Error("Completed game should stay complete")
	}

	// Winner should not change
	if resolution.State.Winner == nil || *resolution.State.Winner != "player" {
		t.Error("Winner should not change in completed game")
	}
}

// TestCheckCombatEnd verifies combat end checking
func TestCheckCombatEnd(t *testing.T) {
	tests := []struct {
		name     string
		state    State
		expected bool
	}{
		{
			name: "Complete game",
			state: State{
				IsComplete: true,
			},
			expected: true,
		},
		{
			name: "Max rounds reached",
			state: State{
				Round:      MaxCombatRounds,
				IsComplete: false,
			},
			expected: true,
		},
		{
			name: "Over max rounds",
			state: State{
				Round:      MaxCombatRounds + 5,
				IsComplete: false,
			},
			expected: true,
		},
		{
			name: "Normal combat",
			state: State{
				Round:      5,
				IsComplete: false,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CheckCombatEnd(tt.state)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// TestGetStateSummary verifies state summary generation
func TestGetStateSummary(t *testing.T) {
	player := createTestCharacter(true, "Hero")
	enemy := createTestCharacter(false, "Goblin")

	state := CreateInitialState([]Character{player}, []Character{enemy}, 12345)

	summary := GetStateSummary(state)

	// Summary should contain key information
	if len(summary) == 0 {
		t.Error("Summary should not be empty")
	}

	// Should mention round
	if !contains(summary, "Round") {
		t.Error("Summary should mention round")
	}

	// Should mention character names
	if !contains(summary, "Hero") {
		t.Error("Summary should mention Hero")
	}
	if !contains(summary, "Goblin") {
		t.Error("Summary should mention Goblin")
	}
}

// TestGetStateSummaryWithVictory verifies summary shows winner
func TestGetStateSummaryWithVictory(t *testing.T) {
	player := createTestCharacter(true, "Hero")
	enemy := createTestCharacter(false, "Goblin")
	enemy.Stats.HP = 0

	state := State{
		Round:      5,
		Characters: []Character{player, enemy},
		TurnOrder:  []ID{player.ID, enemy.ID},
		IsComplete: true,
		Winner:     stringPtr("player"),
	}

	for i := range state.Characters {
		if !state.Characters[i].IsPlayer {
			state.Characters[i].Stats.HP = 0
		}
	}

	summary := GetStateSummary(state)

	// Should mention completion
	if !contains(summary, "Complete") {
		t.Error("Summary should mention combat complete")
	}

	// Should mention winner
	if !contains(summary, "Player") {
		t.Error("Summary should mention winner")
	}
}

// TestFleeVictory verifies flee counts as victory
func TestFleeVictory(t *testing.T) {
	player := createTestCharacter(true, "Player")
	player.Stats.Speed = 50 // High speed to guarantee flee success

	enemy := createTestCharacter(false, "Enemy")

	state := CreateInitialState([]Character{player}, []Character{enemy}, 12345)

	action := Action{
		Kind:  "Flee",
		Actor: player.ID,
	}

	resolution := ApplyAction(state, action, 12345)

	// Check if flee succeeded
	hasFlee := false
	for _, event := range resolution.Events {
		if event.Type == "flee" {
			hasFlee = true
		}
	}

	if hasFlee {
		// Combat should be complete
		if !resolution.State.IsComplete {
			t.Error("Combat should be complete after successful flee")
		}

		// Winner should be player (flee = escape = win)
		if resolution.State.Winner == nil {
			t.Error("Winner should be set after flee")
		}
		if *resolution.State.Winner != "player" {
			t.Errorf("Expected player to win on flee, got %s", *resolution.State.Winner)
		}
	}
}

// Helper to create string pointer
func stringPtr(s string) *string {
	return &s
}
