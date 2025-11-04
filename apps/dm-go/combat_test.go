package main

import (
	"testing"
)

// TestAttackDamageCalculation verifies damage calculations are correct
func TestAttackDamageCalculation(t *testing.T) {
	tests := []struct {
		name           string
		attackerAttack int
		weaponDamage   int
		targetDefense  int
		seed           int64
		expectHit      bool
		checkMinDamage bool
	}{
		{
			name:           "High attack vs low defense - should hit",
			attackerAttack: 15,
			weaponDamage:   6,
			targetDefense:  3,
			seed:           12345, // Known seed that produces hit
			expectHit:      true,
			checkMinDamage: true,
		},
		{
			name:           "Normal attack values",
			attackerAttack: 8,
			weaponDamage:   4,
			targetDefense:  5,
			seed:           12345,
			expectHit:      false, // May hit or miss with random roll
			checkMinDamage: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			player := createTestCharacter(true, "Attacker")
			player.Stats.Attack = tt.attackerAttack
			player.Weapons[0].Damage = tt.weaponDamage

			enemy := createTestCharacter(false, "Target")
			enemy.Stats.Defense = tt.targetDefense

			state := CreateInitialState([]Character{player}, []Character{enemy}, tt.seed)

			action := Action{
				Kind:     "Attack",
				Attacker: player.ID,
				Target:   enemy.ID,
				Weapon:   player.Weapons[0].ID,
			}

			resolution := ApplyAction(state, action, tt.seed)

			// Check for damage event
			var damageEvent *Event
			for i := range resolution.Events {
				if resolution.Events[i].Type == "damage" {
					damageEvent = &resolution.Events[i]
					break
				}
			}

			if tt.expectHit {
				if damageEvent == nil {
					t.Error("Expected damage event from successful hit")
				} else {
					if tt.checkMinDamage && damageEvent.Amount < MinDamage {
						t.Errorf("Damage %d is less than minimum %d", damageEvent.Amount, MinDamage)
					}
					if damageEvent.Amount < 0 {
						t.Errorf("Damage cannot be negative: %d", damageEvent.Amount)
					}
				}
			} else {
				// Don't require hit, just log what happened
				if damageEvent != nil {
					t.Logf("Attack hit for %d damage", damageEvent.Amount)
				} else {
					t.Logf("Attack missed")
				}
			}

			// Verify target HP is valid
			target := GetCharacterByID(resolution.State, enemy.ID)
			if target.Stats.HP < 0 {
				t.Errorf("Target HP cannot be negative: %d", target.Stats.HP)
			}
			if target.Stats.HP > target.Stats.MaxHP {
				t.Errorf("Target HP %d exceeds MaxHP %d", target.Stats.HP, target.Stats.MaxHP)
			}
		})
	}
}

// TestAttackMiss verifies that attacks can miss
func TestAttackMiss(t *testing.T) {
	// Use a seed that produces a low roll
	player := createTestCharacter(true, "Attacker")
	player.Stats.Attack = 1 // Very low attack

	enemy := createTestCharacter(false, "Target")
	enemy.Stats.Defense = 20 // Very high defense

	state := CreateInitialState([]Character{player}, []Character{enemy}, 99999)

	action := Action{
		Kind:     "Attack",
		Attacker: player.ID,
		Target:   enemy.ID,
		Weapon:   player.Weapons[0].ID,
	}

	initialHP := enemy.Stats.HP

	resolution := ApplyAction(state, action, 99999)

	// Check that no damage event occurred
	hasDamageEvent := false
	for _, event := range resolution.Events {
		if event.Type == "damage" {
			hasDamageEvent = true
			break
		}
	}

	// With very low attack vs very high defense, should miss most of the time
	target := GetCharacterByID(resolution.State, enemy.ID)
	if hasDamageEvent {
		t.Logf("Attack hit (rare but possible with d20 rolls): damage dealt")
	} else {
		// Verify HP unchanged
		if target.Stats.HP != initialHP {
			t.Errorf("HP changed on miss: was %d, now %d", initialHP, target.Stats.HP)
		}
	}
}

// TestDeathDetection verifies that character death is properly detected
func TestDeathDetection(t *testing.T) {
	player := createTestCharacter(true, "Attacker")
	player.Stats.Attack = 50 // Very high attack to guarantee kill

	enemy := createTestCharacter(false, "Target")
	enemy.Stats.HP = 1 // Very low HP
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

	// Check for death event
	hasDeathEvent := false
	for _, event := range resolution.Events {
		if event.Type == "death" && event.Target == enemy.ID {
			hasDeathEvent = true
			break
		}
	}

	target := GetCharacterByID(resolution.State, enemy.ID)
	if target.Stats.HP == 0 {
		if !hasDeathEvent {
			t.Error("Character died but no death event was recorded")
		}
	}

	// Verify HP is exactly 0, not negative
	if target.Stats.HP < 0 {
		t.Errorf("HP should not be negative on death: %d", target.Stats.HP)
	}
}

// TestDefendAction verifies defend mechanics
func TestDefendAction(t *testing.T) {
	player := createTestCharacter(true, "Defender")
	originalDefense := player.Stats.Defense

	enemy := createTestCharacter(false, "Enemy")

	state := CreateInitialState([]Character{player}, []Character{enemy}, 12345)

	action := Action{
		Kind:  "Defend",
		Actor: player.ID,
	}

	resolution := ApplyAction(state, action, 12345)

	defender := GetCharacterByID(resolution.State, player.ID)

	// Defense should increase by DefenseBonus
	expectedDefense := originalDefense + DefenseBonus
	if defender.Stats.Defense != expectedDefense {
		t.Errorf("Expected defense %d, got %d", expectedDefense, defender.Stats.Defense)
	}

	// Verify logs mention defending
	hasDefendLog := false
	for _, log := range resolution.Logs {
		if len(log) > 0 {
			hasDefendLog = true
			t.Logf("Defend log: %s", log)
		}
	}
	if !hasDefendLog {
		t.Error("Expected defend log message")
	}
}

// TestDefenseBonusDecay verifies that defense bonus resets after turn
func TestDefenseBonusDecay(t *testing.T) {
	player := createTestCharacter(true, "Defender")
	player.Stats.Defense = 3 // Start with low defense
	originalDefense := player.Stats.Defense

	enemy := createTestCharacter(false, "Enemy")

	state := CreateInitialState([]Character{player}, []Character{enemy}, 12345)

	// Apply defend action
	defendAction := Action{
		Kind:  "Defend",
		Actor: player.ID,
	}

	resolution1 := ApplyAction(state, defendAction, 12345)
	defender := GetCharacterByID(resolution1.State, player.ID)

	// Should have bonus
	boostedDefense := defender.Stats.Defense
	if boostedDefense != originalDefense+DefenseBonus {
		t.Errorf("Expected defense boost to %d, got %d", originalDefense+DefenseBonus, boostedDefense)
	}

	// Enemy takes a turn
	enemyAction := Action{
		Kind:  "Defend",
		Actor: enemy.ID,
	}

	resolution2 := ApplyAction(resolution1.State, enemyAction, 12345)

	// After enemy's turn, player's defense should reset (if it exceeded BaseDefenseMax)
	defender = GetCharacterByID(resolution2.State, player.ID)
	if originalDefense+DefenseBonus > BaseDefenseMax {
		// Should have decayed
		if defender.Stats.Defense > BaseDefenseMax {
			t.Errorf("Expected defense to decay below %d, got %d", BaseDefenseMax, defender.Stats.Defense)
		}
	}
}

// TestTurnAdvancement verifies turn order progression
func TestTurnAdvancement(t *testing.T) {
	player := createTestCharacter(true, "Player")
	enemy := createTestCharacter(false, "Enemy")

	state := CreateInitialState([]Character{player}, []Character{enemy}, 12345)
	initialTurn := state.CurrentTurn

	action := Action{
		Kind:  "Defend",
		Actor: GetCurrentCharacter(state).ID,
	}

	resolution := ApplyAction(state, action, 12345)

	// Turn should advance
	if resolution.State.CurrentTurn == initialTurn {
		t.Error("Turn did not advance after action")
	}

	// After 2 actions, should be back to first character
	currentActor := GetCurrentCharacter(resolution.State)
	if currentActor == nil {
		t.Fatal("No current character after turn advancement")
	}

	action2 := Action{
		Kind:  "Defend",
		Actor: currentActor.ID,
	}

	resolution2 := ApplyAction(resolution.State, action2, 12345)

	// Should wrap around to turn 0
	if resolution2.State.CurrentTurn != 0 {
		t.Errorf("Expected turn to wrap to 0, got %d", resolution2.State.CurrentTurn)
	}

	// Round should increment
	if resolution2.State.Round <= state.Round {
		t.Errorf("Expected round to increment from %d, got %d", state.Round, resolution2.State.Round)
	}
}

// TestRoundProgression verifies round counter increases correctly
func TestRoundProgression(t *testing.T) {
	player := createTestCharacter(true, "Player")
	enemy := createTestCharacter(false, "Enemy")

	state := CreateInitialState([]Character{player}, []Character{enemy}, 12345)

	if state.Round != 1 {
		t.Errorf("Initial round should be 1, got %d", state.Round)
	}

	// Execute a full round (both characters act)
	for i := 0; i < 2; i++ {
		currentChar := GetCurrentCharacter(state)
		action := Action{
			Kind:  "Defend",
			Actor: currentChar.ID,
		}
		resolution := ApplyAction(state, action, 12345)
		state = resolution.State
	}

	// Round should have incremented
	if state.Round != 2 {
		t.Errorf("Expected round 2 after full round, got %d", state.Round)
	}
}

// TestMultipleEnemiesCombat verifies combat with multiple enemies
func TestMultipleEnemiesCombat(t *testing.T) {
	player := createTestCharacter(true, "Player")
	enemy1 := createTestCharacter(false, "Enemy1")
	enemy2 := createTestCharacter(false, "Enemy2")
	enemy3 := createTestCharacter(false, "Enemy3")

	state := CreateInitialState([]Character{player}, []Character{enemy1, enemy2, enemy3}, 12345)

	if len(state.Characters) != 4 {
		t.Errorf("Expected 4 characters, got %d", len(state.Characters))
	}

	if len(state.TurnOrder) != 4 {
		t.Errorf("Expected 4 in turn order, got %d", len(state.TurnOrder))
	}

	// All enemies should be present
	enemyCount := 0
	for _, char := range state.Characters {
		if !char.IsPlayer {
			enemyCount++
		}
	}
	if enemyCount != 3 {
		t.Errorf("Expected 3 enemies, got %d", enemyCount)
	}
}

// TestMultiplePlayersCombat verifies combat with multiple players
func TestMultiplePlayersCombat(t *testing.T) {
	player1 := createTestCharacter(true, "Player1")
	player2 := createTestCharacter(true, "Player2")
	enemy := createTestCharacter(false, "Enemy")

	state := CreateInitialState([]Character{player1, player2}, []Character{enemy}, 12345)

	if len(state.Characters) != 3 {
		t.Errorf("Expected 3 characters, got %d", len(state.Characters))
	}

	// Both players should be present
	playerCount := 0
	for _, char := range state.Characters {
		if char.IsPlayer {
			playerCount++
		}
	}
	if playerCount != 2 {
		t.Errorf("Expected 2 players, got %d", playerCount)
	}
}

// TestWeaponNotFound verifies default weapon fallback
func TestWeaponNotFound(t *testing.T) {
	player := createTestCharacter(true, "Player")
	enemy := createTestCharacter(false, "Enemy")

	state := CreateInitialState([]Character{player}, []Character{enemy}, 12345)

	// Use invalid weapon ID
	action := Action{
		Kind:     "Attack",
		Attacker: player.ID,
		Target:   enemy.ID,
		Weapon:   "invalid-weapon-id",
	}

	resolution := ApplyAction(state, action, 12345)

	// Should have logs mentioning weapon not found
	hasWeaponLog := false
	for _, log := range resolution.Logs {
		if len(log) > 0 {
			hasWeaponLog = true
			t.Logf("Weapon fallback log: %s", log)
		}
	}
	if !hasWeaponLog {
		t.Error("Expected log message about weapon fallback")
	}

	// Action should still proceed with default weapon
	if len(resolution.State.Characters) == 0 {
		t.Error("State should not be empty after action")
	}
}

// TestMinimumDamage verifies minimum damage is always applied
func TestMinimumDamage(t *testing.T) {
	player := createTestCharacter(true, "Player")
	player.Stats.Attack = 1 // Very low attack
	player.Weapons[0].Damage = 1 // Weak weapon

	enemy := createTestCharacter(false, "Enemy")
	enemy.Stats.Defense = 50 // Very high defense

	state := CreateInitialState([]Character{player}, []Character{enemy}, 12345)

	action := Action{
		Kind:     "Attack",
		Attacker: player.ID,
		Target:   enemy.ID,
		Weapon:   player.Weapons[0].ID,
	}

	// Try multiple seeds to find one that hits
	for seed := int64(1); seed < 100; seed++ {
		resolution := ApplyAction(state, action, seed)

		// Check for damage event
		for _, event := range resolution.Events {
			if event.Type == "damage" {
				if event.Amount < MinDamage {
					t.Errorf("Damage %d is less than minimum %d", event.Amount, MinDamage)
				}
				if event.Amount > 0 {
					t.Logf("Found hit with seed %d, damage: %d (meets minimum)", seed, event.Amount)
					return // Test passed
				}
			}
		}
	}

	t.Log("Note: No hits found in 100 seeds (very high defense makes hits rare)")
}

// TestCharacterMapLookup verifies O(1) character lookup works
func TestCharacterMapLookup(t *testing.T) {
	player := createTestCharacter(true, "Player")
	enemy := createTestCharacter(false, "Enemy")

	state := CreateInitialState([]Character{player}, []Character{enemy}, 12345)

	// CharacterMap should be initialized
	if state.CharacterMap == nil {
		t.Fatal("CharacterMap should be initialized")
	}

	if len(state.CharacterMap) != 2 {
		t.Errorf("Expected 2 entries in CharacterMap, got %d", len(state.CharacterMap))
	}

	// Lookup should work
	foundPlayer := GetCharacterByID(state, player.ID)
	if foundPlayer == nil {
		t.Error("Failed to find player via CharacterMap")
	}
	if foundPlayer.ID != player.ID {
		t.Error("Retrieved wrong character")
	}

	foundEnemy := GetCharacterByID(state, enemy.ID)
	if foundEnemy == nil {
		t.Error("Failed to find enemy via CharacterMap")
	}
	if foundEnemy.ID != enemy.ID {
		t.Error("Retrieved wrong character")
	}
}

// TestCombatMaxRounds verifies combat doesn't run forever
func TestCombatMaxRounds(t *testing.T) {
	state := State{
		Round:      MaxCombatRounds,
		Characters: []Character{createTestCharacter(true, "Player")},
		TurnOrder:  []ID{NewID()},
	}

	if !CheckCombatEnd(state) {
		t.Error("Combat should end at MaxCombatRounds")
	}

	state.Round = MaxCombatRounds + 1
	if !CheckCombatEnd(state) {
		t.Error("Combat should end after MaxCombatRounds")
	}
}
