package main

import (
	"testing"
)

// TestRNGDeterminism verifies RNG produces same results with same seed
func TestRNGDeterminism(t *testing.T) {
	seed := int64(12345)

	rng1 := NewSeededRNG(seed)
	rng2 := NewSeededRNG(seed)

	// Roll multiple dice
	for i := 0; i < 10; i++ {
		roll1 := rng1.RollD20()
		roll2 := rng2.RollD20()

		if roll1 != roll2 {
			t.Errorf("RNG not deterministic: roll %d gave %d and %d", i, roll1, roll2)
		}
	}
}

// TestRollD20Range verifies d20 rolls are in valid range
func TestRollD20Range(t *testing.T) {
	rng := NewSeededRNG(12345)

	for i := 0; i < 1000; i++ {
		roll := rng.RollD20()
		if roll < 1 || roll > 20 {
			t.Errorf("D20 roll out of range: %d", roll)
		}
	}
}

// TestRollD6Range verifies d6 rolls are in valid range
func TestRollD6Range(t *testing.T) {
	rng := NewSeededRNG(12345)

	for i := 0; i < 1000; i++ {
		roll := rng.RollD6()
		if roll < 1 || roll > 6 {
			t.Errorf("D6 roll out of range: %d", roll)
		}
	}
}

// TestRollD100Range verifies d100 rolls are in valid range
func TestRollD100Range(t *testing.T) {
	rng := NewSeededRNG(12345)

	for i := 0; i < 1000; i++ {
		roll := rng.RollD100()
		if roll < 1 || roll > 100 {
			t.Errorf("D100 roll out of range: %d", roll)
		}
	}
}

// TestRandomIntRange verifies RandomInt stays in bounds
func TestRandomIntRange(t *testing.T) {
	rng := NewSeededRNG(12345)

	tests := []struct {
		min int
		max int
	}{
		{1, 10},
		{0, 5},
		{-5, 5},
		{100, 200},
	}

	for _, tt := range tests {
		for i := 0; i < 100; i++ {
			result := rng.RandomInt(tt.min, tt.max)
			if result < tt.min || result > tt.max {
				t.Errorf("RandomInt(%d, %d) out of range: %d", tt.min, tt.max, result)
			}
		}
	}
}

// TestRandomIntEdgeCases verifies edge cases
func TestRandomIntEdgeCases(t *testing.T) {
	rng := NewSeededRNG(12345)

	// Same min and max
	result := rng.RandomInt(5, 5)
	if result != 5 {
		t.Errorf("RandomInt(5, 5) should be 5, got %d", result)
	}

	// Min greater than max (should return min)
	result = rng.RandomInt(10, 5)
	if result != 10 {
		t.Errorf("RandomInt(10, 5) should return min (10), got %d", result)
	}
}

// TestRNGDistribution verifies d20 has reasonable distribution
func TestRNGDistribution(t *testing.T) {
	rng := NewSeededRNG(12345)

	rolls := make(map[int]int)
	totalRolls := 10000

	// Collect rolls
	for i := 0; i < totalRolls; i++ {
		roll := rng.RollD20()
		rolls[roll]++
	}

	// Check all values appear
	for i := 1; i <= 20; i++ {
		if rolls[i] == 0 {
			t.Errorf("D20 never rolled %d in %d rolls", i, totalRolls)
		}
	}

	// Check distribution is reasonable (each value should appear ~500 times)
	// Allow wide tolerance for randomness
	minExpected := totalRolls / 20 / 2   // 250
	maxExpected := totalRolls / 20 * 2   // 1000

	for value, count := range rolls {
		if count < minExpected || count > maxExpected {
			t.Logf("Warning: D20 value %d appeared %d times (expected ~%d)", value, count, totalRolls/20)
		}
	}
}

// TestD6Distribution verifies d6 has reasonable distribution
func TestD6Distribution(t *testing.T) {
	rng := NewSeededRNG(54321)

	rolls := make(map[int]int)
	totalRolls := 6000

	for i := 0; i < totalRolls; i++ {
		roll := rng.RollD6()
		rolls[roll]++
	}

	// Check all values appear
	for i := 1; i <= 6; i++ {
		if rolls[i] == 0 {
			t.Errorf("D6 never rolled %d in %d rolls", i, totalRolls)
		}
	}
}

// TestDifferentSeedsDifferentResults verifies different seeds produce different sequences
func TestDifferentSeedsDifferentResults(t *testing.T) {
	rng1 := NewSeededRNG(12345)
	rng2 := NewSeededRNG(54321)

	// Sequences should diverge
	same := 0
	total := 100

	for i := 0; i < total; i++ {
		roll1 := rng1.RollD20()
		roll2 := rng2.RollD20()
		if roll1 == roll2 {
			same++
		}
	}

	// Some overlap is expected, but not all
	if same == total {
		t.Error("Different seeds produced identical sequences")
	}

	// Log how different they are
	t.Logf("Different seeds: %d/%d rolls were the same", same, total)
}

// TestRNGInCombat verifies RNG works correctly in combat
func TestRNGInCombat(t *testing.T) {
	seed := int64(12345)

	player := createTestCharacter(true, "Player")
	enemy := createTestCharacter(false, "Enemy")

	state1 := CreateInitialState([]Character{player}, []Character{enemy}, seed)
	state2 := CreateInitialState([]Character{player}, []Character{enemy}, seed)

	// Initiative order should be identical
	if len(state1.TurnOrder) != len(state2.TurnOrder) {
		t.Error("Turn order length differs with same seed")
	}

	for i := range state1.TurnOrder {
		if state1.TurnOrder[i] != state2.TurnOrder[i] {
			t.Error("Turn order differs with same seed")
		}
	}

	// Combat actions should be deterministic
	action := Action{
		Kind:     "Attack",
		Attacker: player.ID,
		Target:   enemy.ID,
		Weapon:   player.Weapons[0].ID,
	}

	res1 := ApplyAction(state1, action, seed)
	res2 := ApplyAction(state2, action, seed)

	// Results should be identical
	if len(res1.Events) != len(res2.Events) {
		t.Error("Different number of events with same seed")
	}

	if len(res1.Logs) != len(res2.Logs) {
		t.Error("Different number of logs with same seed")
	}

	// Damage should be identical
	enemy1 := GetCharacterByID(res1.State, enemy.ID)
	enemy2 := GetCharacterByID(res2.State, enemy.ID)

	if enemy1.Stats.HP != enemy2.Stats.HP {
		t.Errorf("Different HP after attack with same seed: %d vs %d", enemy1.Stats.HP, enemy2.Stats.HP)
	}
}

// TestAbilityDeterminism verifies abilities are deterministic
func TestAbilityDeterminism(t *testing.T) {
	seed := int64(99999)

	ability := Ability{
		ID:       NewID(),
		Name:     "Power Attack",
		Cooldown: 3,
		Effect:   "damage",
		Power:    10,
	}

	player1 := createTestCharacter(true, "Player")
	player1.Abilities = []Ability{ability}
	player1.AbilityCooldowns = make(map[string]int)

	player2 := createTestCharacter(true, "Player")
	player2.Abilities = []Ability{ability}
	player2.AbilityCooldowns = make(map[string]int)

	enemy1 := createTestCharacter(false, "Enemy")
	enemy2 := createTestCharacter(false, "Enemy")

	state1 := CreateInitialState([]Character{player1}, []Character{enemy1}, seed)
	state2 := CreateInitialState([]Character{player2}, []Character{enemy2}, seed)

	action1 := Action{
		Kind:    "Ability",
		Actor:   player1.ID,
		Ability: ability.ID,
		Target:  enemy1.ID,
	}

	action2 := Action{
		Kind:    "Ability",
		Actor:   player2.ID,
		Ability: ability.ID,
		Target:  enemy2.ID,
	}

	res1 := ApplyAction(state1, action1, seed)
	res2 := ApplyAction(state2, action2, seed)

	// Damage should be identical
	enemy1Updated := GetCharacterByID(res1.State, enemy1.ID)
	enemy2Updated := GetCharacterByID(res2.State, enemy2.ID)

	if enemy1Updated.Stats.HP != enemy2Updated.Stats.HP {
		t.Errorf("Ability damage not deterministic: %d vs %d HP remaining",
			enemy1Updated.Stats.HP, enemy2Updated.Stats.HP)
	}
}

// TestItemDeterminism verifies item effects are deterministic
func TestItemDeterminism(t *testing.T) {
	seed := int64(77777)

	potion1 := Item{
		ID:     NewID(),
		Name:   "Health Potion",
		Type:   "consumable",
		Effect: "heal 20 HP",
	}

	potion2 := Item{
		ID:     NewID(),
		Name:   "Health Potion",
		Type:   "consumable",
		Effect: "heal 20 HP",
	}

	player1 := createTestCharacter(true, "Player")
	player1.Stats.HP = 10
	player1.Items = []Item{potion1}

	player2 := createTestCharacter(true, "Player")
	player2.Stats.HP = 10
	player2.Items = []Item{potion2}

	enemy1 := createTestCharacter(false, "Enemy")
	enemy2 := createTestCharacter(false, "Enemy")

	state1 := CreateInitialState([]Character{player1}, []Character{enemy1}, seed)
	state2 := CreateInitialState([]Character{player2}, []Character{enemy2}, seed)

	// Set HP in state
	for i := range state1.Characters {
		if state1.Characters[i].ID == player1.ID {
			state1.Characters[i].Stats.HP = 10
			state1.Characters[i].Items = []Item{potion1}
		}
	}
	state1.CharacterMap = buildCharacterMap(&state1)

	for i := range state2.Characters {
		if state2.Characters[i].ID == player2.ID {
			state2.Characters[i].Stats.HP = 10
			state2.Characters[i].Items = []Item{potion2}
		}
	}
	state2.CharacterMap = buildCharacterMap(&state2)

	action1 := Action{
		Kind:  "UseItem",
		Actor: player1.ID,
		Item:  potion1.ID,
	}

	action2 := Action{
		Kind:  "UseItem",
		Actor: player2.ID,
		Item:  potion2.ID,
	}

	res1 := ApplyAction(state1, action1, seed)
	res2 := ApplyAction(state2, action2, seed)

	// Healing should be identical
	player1Updated := GetCharacterByID(res1.State, player1.ID)
	player2Updated := GetCharacterByID(res2.State, player2.ID)

	if player1Updated.Stats.HP != player2Updated.Stats.HP {
		t.Errorf("Item healing not deterministic: %d vs %d HP",
			player1Updated.Stats.HP, player2Updated.Stats.HP)
	}
}

// TestCriticalHitDeterminism verifies critical hits are deterministic
func TestCriticalHitDeterminism(t *testing.T) {
	// Run many attacks with same seed to see if outcomes match
	seed := int64(42)

	player := createTestCharacter(true, "Player")
	enemy := createTestCharacter(false, "Enemy")

	// Track if same seed produces same outcome
	outcomes1 := []int{}
	outcomes2 := []int{}

	for trial := 0; trial < 10; trial++ {
		state1 := CreateInitialState([]Character{player}, []Character{enemy}, seed)
		state2 := CreateInitialState([]Character{player}, []Character{enemy}, seed)

		action := Action{
			Kind:     "Attack",
			Attacker: player.ID,
			Target:   enemy.ID,
			Weapon:   player.Weapons[0].ID,
		}

		res1 := ApplyAction(state1, action, int64(trial))
		res2 := ApplyAction(state2, action, int64(trial))

		enemy1 := GetCharacterByID(res1.State, enemy.ID)
		enemy2 := GetCharacterByID(res2.State, enemy.ID)

		outcomes1 = append(outcomes1, enemy1.Stats.HP)
		outcomes2 = append(outcomes2, enemy2.Stats.HP)
	}

	// Outcomes should match
	for i := range outcomes1 {
		if outcomes1[i] != outcomes2[i] {
			t.Errorf("Trial %d: outcomes differ with same seeds: %d vs %d",
				i, outcomes1[i], outcomes2[i])
		}
	}
}
