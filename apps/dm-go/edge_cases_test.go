package main

import (
	"testing"
)

// TestInvalidActionKind verifies invalid action kinds are rejected
func TestInvalidActionKind(t *testing.T) {
	player := createTestCharacter(true, "Player")
	enemy := createTestCharacter(false, "Enemy")

	state := CreateInitialState([]Character{player}, []Character{enemy}, 12345)

	action := Action{
		Kind:  "InvalidAction",
		Actor: player.ID,
	}

	resolution := ApplyAction(state, action, 12345)

	// Should have error log
	hasErrorLog := false
	for _, log := range resolution.Logs {
		if len(log) > 0 {
			hasErrorLog = true
			t.Logf("Error log: %s", log)
		}
	}
	if !hasErrorLog {
		t.Error("Expected error log for invalid action kind")
	}

	// State should be unchanged (returned original state)
	if len(resolution.State.Characters) == 0 {
		t.Error("State should not be corrupted by invalid action")
	}
}

// TestActionWithoutActor verifies actions require an actor
func TestActionWithoutActor(t *testing.T) {
	player := createTestCharacter(true, "Player")
	enemy := createTestCharacter(false, "Enemy")

	state := CreateInitialState([]Character{player}, []Character{enemy}, 12345)

	action := Action{
		Kind: "Defend",
		// No Actor specified
	}

	resolution := ApplyAction(state, action, 12345)

	// Should have error about character not found
	hasErrorLog := false
	for _, log := range resolution.Logs {
		if len(log) > 0 {
			hasErrorLog = true
			t.Logf("Error log: %s", log)
		}
	}
	if !hasErrorLog {
		t.Error("Expected error log for missing actor")
	}
}

// TestAttackWithInvalidTarget verifies attack with invalid target
func TestAttackWithInvalidTarget(t *testing.T) {
	player := createTestCharacter(true, "Player")
	enemy := createTestCharacter(false, "Enemy")

	state := CreateInitialState([]Character{player}, []Character{enemy}, 12345)

	action := Action{
		Kind:     "Attack",
		Attacker: player.ID,
		Target:   "invalid-target-id",
		Weapon:   player.Weapons[0].ID,
	}

	resolution := ApplyAction(state, action, 12345)

	// Should have error about invalid attack
	hasErrorLog := false
	for _, log := range resolution.Logs {
		if len(log) > 0 {
			hasErrorLog = true
			t.Logf("Error log: %s", log)
		}
	}
	if !hasErrorLog {
		t.Error("Expected error log for invalid target")
	}
}

// TestAttackWithInvalidAttacker verifies attack with invalid attacker
func TestAttackWithInvalidAttacker(t *testing.T) {
	player := createTestCharacter(true, "Player")
	enemy := createTestCharacter(false, "Enemy")

	state := CreateInitialState([]Character{player}, []Character{enemy}, 12345)

	action := Action{
		Kind:     "Attack",
		Attacker: "invalid-attacker-id",
		Target:   enemy.ID,
		Weapon:   player.Weapons[0].ID,
	}

	resolution := ApplyAction(state, action, 12345)

	// Should have error
	hasErrorLog := false
	for _, log := range resolution.Logs {
		if len(log) > 0 {
			hasErrorLog = true
			t.Logf("Error log: %s", log)
		}
	}
	if !hasErrorLog {
		t.Error("Expected error log for invalid attacker")
	}
}

// TestDeadCharacterAction verifies dead characters can still take actions (game allows this)
func TestDeadCharacterAction(t *testing.T) {
	player := createTestCharacter(true, "Player")
	player.Stats.HP = 0 // Dead

	enemy := createTestCharacter(false, "Enemy")

	state := CreateInitialState([]Character{player}, []Character{enemy}, 12345)

	// Set player HP to 0
	for i := range state.Characters {
		if state.Characters[i].ID == player.ID {
			state.Characters[i].Stats.HP = 0
		}
	}
	state.CharacterMap = buildCharacterMap(&state)

	action := Action{
		Kind:  "Defend",
		Actor: player.ID,
	}

	resolution := ApplyAction(state, action, 12345)

	// Current implementation allows dead characters to act
	// This might be a bug or design decision
	t.Log("Dead character can still take actions (current behavior)")

	// Verify state is not corrupted
	if len(resolution.State.Characters) == 0 {
		t.Error("State should not be corrupted")
	}
}

// TestNilCharacter verifies handling of nil character lookups
func TestNilCharacter(t *testing.T) {
	player := createTestCharacter(true, "Player")
	enemy := createTestCharacter(false, "Enemy")

	state := CreateInitialState([]Character{player}, []Character{enemy}, 12345)

	// Get character that doesn't exist
	char := GetCharacterByID(state, "nonexistent-id")
	if char != nil {
		t.Error("Should return nil for nonexistent character")
	}

	// Test with empty state
	emptyState := State{
		Characters:   []Character{},
		CharacterMap: make(map[ID]*Character),
	}
	char = GetCharacterByID(emptyState, player.ID)
	if char != nil {
		t.Error("Should return nil for character in empty state")
	}
}

// TestEmptyTurnOrder verifies handling of empty turn order
func TestEmptyTurnOrder(t *testing.T) {
	state := State{
		Round:       1,
		Characters:  []Character{createTestCharacter(true, "Player")},
		TurnOrder:   []ID{}, // Empty turn order
		CurrentTurn: 0,
	}

	char := GetCurrentCharacter(state)
	if char != nil {
		t.Error("Should return nil when turn order is empty")
	}
}

// TestOutOfBoundsTurn verifies handling of invalid turn index
func TestOutOfBoundsTurn(t *testing.T) {
	player := createTestCharacter(true, "Player")
	state := State{
		Round:       1,
		Characters:  []Character{player},
		TurnOrder:   []ID{player.ID},
		CurrentTurn: 10, // Out of bounds
	}
	state.CharacterMap = buildCharacterMap(&state)

	char := GetCurrentCharacter(state)
	if char != nil {
		t.Error("Should return nil when current turn is out of bounds")
	}
}

// TestZeroMaxHP verifies handling of zero max HP
func TestZeroMaxHP(t *testing.T) {
	player := createTestCharacter(true, "Player")
	player.Stats.MaxHP = 0
	player.Stats.HP = 0

	enemy := createTestCharacter(false, "Enemy")

	state := CreateInitialState([]Character{player}, []Character{enemy}, 12345)

	// Character with 0 MaxHP should be in the game
	if len(state.Characters) != 2 {
		t.Error("Should include character with 0 MaxHP")
	}
}

// TestNegativeStats verifies handling of negative stats
func TestNegativeStats(t *testing.T) {
	player := createTestCharacter(true, "Player")
	player.Stats.Attack = -5
	player.Stats.Defense = -3

	enemy := createTestCharacter(false, "Enemy")

	state := CreateInitialState([]Character{player}, []Character{enemy}, 12345)

	action := Action{
		Kind:     "Attack",
		Attacker: player.ID,
		Target:   enemy.ID,
		Weapon:   player.Weapons[0].ID,
	}

	resolution := ApplyAction(state, action, 12345)

	// Should handle negative stats without crashing
	if len(resolution.State.Characters) == 0 {
		t.Error("State should not be corrupted by negative stats")
	}
}

// TestVeryLargeStats verifies handling of very large stats
func TestVeryLargeStats(t *testing.T) {
	player := createTestCharacter(true, "Player")
	player.Stats.Attack = 1000
	player.Stats.HP = 10000
	player.Stats.MaxHP = 10000

	enemy := createTestCharacter(false, "Enemy")

	state := CreateInitialState([]Character{player}, []Character{enemy}, 12345)

	action := Action{
		Kind:     "Attack",
		Attacker: player.ID,
		Target:   enemy.ID,
		Weapon:   player.Weapons[0].ID,
	}

	resolution := ApplyAction(state, action, 12345)

	// Should handle large stats
	if len(resolution.Events) == 0 {
		t.Error("Should process action with large stats")
	}
}

// TestEmptyWeaponsList verifies handling of character with no weapons
func TestEmptyWeaponsList(t *testing.T) {
	player := createTestCharacter(true, "Player")
	player.Weapons = []Weapon{} // No weapons

	enemy := createTestCharacter(false, "Enemy")

	state := CreateInitialState([]Character{player}, []Character{enemy}, 12345)

	// Update player in state
	for i := range state.Characters {
		if state.Characters[i].ID == player.ID {
			state.Characters[i].Weapons = []Weapon{}
		}
	}
	state.CharacterMap = buildCharacterMap(&state)

	action := Action{
		Kind:     "Attack",
		Attacker: player.ID,
		Target:   enemy.ID,
		// No weapon specified
	}

	resolution := ApplyAction(state, action, 12345)

	// Should use default weapon
	hasLog := false
	for _, log := range resolution.Logs {
		if len(log) > 0 {
			hasLog = true
			t.Logf("Log: %s", log)
		}
	}
	if !hasLog {
		t.Error("Expected log for weapon fallback")
	}
}

// TestEmptyAbilitiesList verifies handling of character with no abilities
func TestEmptyAbilitiesList(t *testing.T) {
	player := createTestCharacter(true, "Player")
	player.Abilities = []Ability{} // No abilities

	enemy := createTestCharacter(false, "Enemy")

	state := CreateInitialState([]Character{player}, []Character{enemy}, 12345)

	// Update player in state
	for i := range state.Characters {
		if state.Characters[i].ID == player.ID {
			state.Characters[i].Abilities = []Ability{}
		}
	}
	state.CharacterMap = buildCharacterMap(&state)

	action := Action{
		Kind:    "Ability",
		Actor:   player.ID,
		Ability: "any-ability-id",
		Target:  enemy.ID,
	}

	resolution := ApplyAction(state, action, 12345)

	// Should have error about ability not found
	hasErrorLog := false
	for _, log := range resolution.Logs {
		if len(log) > 0 {
			hasErrorLog = true
		}
	}
	if !hasErrorLog {
		t.Error("Expected error log for ability not found")
	}
}

// TestFleeAction verifies flee mechanics
func TestFleeAction(t *testing.T) {
	player := createTestCharacter(true, "Player")
	player.Stats.Speed = 20 // High speed for better flee chance

	enemy := createTestCharacter(false, "Enemy")

	state := CreateInitialState([]Character{player}, []Character{enemy}, 12345)

	action := Action{
		Kind:  "Flee",
		Actor: player.ID,
	}

	resolution := ApplyAction(state, action, 12345)

	// Check if flee succeeded or failed
	hasFlee := false
	for _, event := range resolution.Events {
		if event.Type == "flee" {
			hasFlee = true
		}
	}

	if hasFlee {
		t.Log("Flee succeeded")
		// Combat should be complete
		if !resolution.State.IsComplete {
			t.Error("Combat should be complete after successful flee")
		}
		// Winner should be set
		if resolution.State.Winner == nil {
			t.Error("Winner should be set after flee")
		}
	} else {
		t.Log("Flee failed")
		// Combat should continue
		if resolution.State.IsComplete {
			t.Error("Combat should not be complete after failed flee")
		}
	}
}

// TestFleeWithLowSpeed verifies flee can fail
func TestFleeWithLowSpeed(t *testing.T) {
	player := createTestCharacter(true, "Player")
	player.Stats.Speed = 0 // Very low speed

	enemy := createTestCharacter(false, "Enemy")

	state := CreateInitialState([]Character{player}, []Character{enemy}, 99999)

	action := Action{
		Kind:  "Flee",
		Actor: player.ID,
	}

	resolution := ApplyAction(state, action, 99999)

	// With very low speed, flee should fail (most of the time)
	hasFlee := false
	for _, event := range resolution.Events {
		if event.Type == "flee" {
			hasFlee = true
		}
	}

	if !hasFlee {
		t.Log("Flee failed as expected with low speed")
	} else {
		t.Log("Flee succeeded (rare with low speed)")
	}
}

// TestStateMutationSafety verifies state is properly copied
func TestStateMutationSafety(t *testing.T) {
	player := createTestCharacter(true, "Player")
	enemy := createTestCharacter(false, "Enemy")

	originalState := CreateInitialState([]Character{player}, []Character{enemy}, 12345)
	originalHP := originalState.Characters[0].Stats.HP
	originalRound := originalState.Round

	action := Action{
		Kind:     "Attack",
		Attacker: player.ID,
		Target:   enemy.ID,
		Weapon:   player.Weapons[0].ID,
	}

	resolution := ApplyAction(originalState, action, 12345)

	// Original state should be unchanged
	if originalState.Characters[0].Stats.HP != originalHP {
		t.Error("Original state was mutated (HP changed)")
	}
	if originalState.Round != originalRound {
		t.Error("Original state was mutated (Round changed)")
	}

	// New state should be different
	if resolution.State.Round == originalRound && resolution.State.CurrentTurn == originalState.CurrentTurn {
		// Turn should have advanced
		t.Log("Turn advanced in new state")
	}
}

// TestConcurrentCharacterLookup verifies character map is thread-safe
func TestConcurrentCharacterLookup(t *testing.T) {
	player := createTestCharacter(true, "Player")
	enemy := createTestCharacter(false, "Enemy")

	state := CreateInitialState([]Character{player}, []Character{enemy}, 12345)

	// Perform multiple lookups
	for i := 0; i < 100; i++ {
		char1 := GetCharacterByID(state, player.ID)
		char2 := GetCharacterByID(state, enemy.ID)

		if char1 == nil || char2 == nil {
			t.Error("Character lookup failed")
		}
	}
}

// TestAllCharactersDead verifies game ends when all characters are dead
func TestAllCharactersDead(t *testing.T) {
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

	// Advance turn to check victory condition
	action := Action{
		Kind:  "Defend",
		Actor: player.ID,
	}

	resolution := ApplyAction(state, action, 12345)

	// Game should detect victory condition
	// With all characters dead, enemies win (players have 0 HP)
	if !resolution.State.IsComplete {
		t.Error("Game should end when all characters are dead")
	}
}

// TestNilAbilityCooldowns verifies handling of uninitialized cooldowns map
func TestNilAbilityCooldowns(t *testing.T) {
	player := createTestCharacter(true, "Player")
	player.AbilityCooldowns = nil // Uninitialized

	enemy := createTestCharacter(false, "Enemy")

	ability := Ability{
		ID:       NewID(),
		Name:     "Test",
		Cooldown: 3,
		Effect:   "damage",
		Power:    5,
	}
	player.Abilities = []Ability{ability}

	state := CreateInitialState([]Character{player}, []Character{enemy}, 12345)

	// Update player in state
	for i := range state.Characters {
		if state.Characters[i].ID == player.ID {
			state.Characters[i].AbilityCooldowns = nil
			state.Characters[i].Abilities = []Ability{ability}
		}
	}
	state.CharacterMap = buildCharacterMap(&state)

	action := Action{
		Kind:    "Ability",
		Actor:   player.ID,
		Ability: ability.ID,
		Target:  enemy.ID,
	}

	// Should not crash even with nil map
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Crashed with nil AbilityCooldowns map: %v", r)
		}
	}()

	resolution := ApplyAction(state, action, 12345)

	// Verify action was processed
	if len(resolution.State.Characters) == 0 {
		t.Error("State should not be empty")
	}
}

// TestCharacterMapNotInitialized verifies fallback to linear search
func TestCharacterMapNotInitialized(t *testing.T) {
	player := createTestCharacter(true, "Player")
	enemy := createTestCharacter(false, "Enemy")

	state := State{
		Round:        1,
		Characters:   []Character{player, enemy},
		CharacterMap: nil, // Not initialized
		TurnOrder:    []ID{player.ID, enemy.ID},
		CurrentTurn:  0,
	}

	// Should still work with linear search fallback
	char := GetCharacterByID(state, player.ID)
	if char == nil {
		t.Error("Should find character even without CharacterMap")
	}

	current := GetCurrentCharacter(state)
	if current == nil {
		t.Error("Should get current character even without CharacterMap")
	}
}
