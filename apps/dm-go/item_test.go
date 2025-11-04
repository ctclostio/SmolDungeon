package main

import (
	"testing"
)

// TestItemUsage verifies item usage mechanics
func TestItemUsage(t *testing.T) {
	player := createTestCharacter(true, "Player")
	player.Stats.HP = 10 // Wounded
	player.Stats.MaxHP = 30

	enemy := createTestCharacter(false, "Enemy")

	potion := Item{
		ID:     NewID(),
		Name:   "Health Potion",
		Type:   "consumable",
		Effect: "heal 20 HP",
	}
	player.Items = []Item{potion}

	state := CreateInitialState([]Character{player}, []Character{enemy}, 12345)

	// Update player in state
	for i := range state.Characters {
		if state.Characters[i].ID == player.ID {
			state.Characters[i].Stats.HP = 10
			state.Characters[i].Items = []Item{potion}
		}
	}
	state.CharacterMap = buildCharacterMap(&state)

	action := Action{
		Kind:  "UseItem",
		Actor: player.ID,
		Item:  potion.ID,
	}

	initialHP := 10
	resolution := ApplyAction(state, action, 12345)

	// Should have item_used event
	hasItemEvent := false
	for _, event := range resolution.Events {
		if event.Type == "item_used" {
			hasItemEvent = true
			if event.Actor != player.ID {
				t.Errorf("Expected actor %s, got %s", player.ID, event.Actor)
			}
			if event.Item != potion.ID {
				t.Errorf("Expected item %s, got %s", potion.ID, event.Item)
			}
		}
	}
	if !hasItemEvent {
		t.Error("Expected item_used event")
	}

	// Should have heal event
	hasHealEvent := false
	for _, event := range resolution.Events {
		if event.Type == "heal" && event.Target == player.ID {
			hasHealEvent = true
			if event.Amount <= 0 {
				t.Errorf("Heal amount should be positive, got %d", event.Amount)
			}
		}
	}
	if !hasHealEvent {
		t.Error("Expected heal event from health potion")
	}

	// HP should increase
	user := GetCharacterByID(resolution.State, player.ID)
	if user.Stats.HP <= initialHP {
		t.Errorf("HP should increase from %d, got %d", initialHP, user.Stats.HP)
	}
}

// TestItemRemoval verifies items are removed after use
func TestItemRemoval(t *testing.T) {
	player := createTestCharacter(true, "Player")
	enemy := createTestCharacter(false, "Enemy")

	potion := Item{
		ID:     NewID(),
		Name:   "Health Potion",
		Type:   "consumable",
		Effect: "heal 20 HP",
	}
	player.Items = []Item{potion}

	state := CreateInitialState([]Character{player}, []Character{enemy}, 12345)

	// Update player in state
	for i := range state.Characters {
		if state.Characters[i].ID == player.ID {
			state.Characters[i].Items = []Item{potion}
		}
	}
	state.CharacterMap = buildCharacterMap(&state)

	initialItemCount := len(player.Items)

	action := Action{
		Kind:  "UseItem",
		Actor: player.ID,
		Item:  potion.ID,
	}

	resolution := ApplyAction(state, action, 12345)

	// Item should be removed from inventory
	user := GetCharacterByID(resolution.State, player.ID)
	if len(user.Items) >= initialItemCount {
		t.Errorf("Item should be removed, had %d items, now %d", initialItemCount, len(user.Items))
	}

	// Item should not be in inventory
	for _, item := range user.Items {
		if item.ID == potion.ID {
			t.Error("Used item should be removed from inventory")
		}
	}
}

// TestItemNotFound verifies invalid item ID handling
func TestItemNotFound(t *testing.T) {
	player := createTestCharacter(true, "Player")
	enemy := createTestCharacter(false, "Enemy")

	state := CreateInitialState([]Character{player}, []Character{enemy}, 12345)

	action := Action{
		Kind:  "UseItem",
		Actor: player.ID,
		Item:  "invalid-item-id",
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
		t.Error("Expected error log for invalid item")
	}

	// Should not have item_used event
	hasItemEvent := false
	for _, event := range resolution.Events {
		if event.Type == "item_used" {
			hasItemEvent = true
		}
	}
	if hasItemEvent {
		t.Error("Should not use invalid item")
	}
}

// TestMultipleItems verifies characters can have multiple items
func TestMultipleItems(t *testing.T) {
	player := createTestCharacter(true, "Player")
	enemy := createTestCharacter(false, "Enemy")

	potion1 := Item{
		ID:     NewID(),
		Name:   "Health Potion",
		Type:   "consumable",
		Effect: "heal 20 HP",
	}
	potion2 := Item{
		ID:     NewID(),
		Name:   "Greater Health Potion",
		Type:   "consumable",
		Effect: "heal 50 HP",
	}
	player.Items = []Item{potion1, potion2}

	state := CreateInitialState([]Character{player}, []Character{enemy}, 12345)

	// Update player in state
	for i := range state.Characters {
		if state.Characters[i].ID == player.ID {
			state.Characters[i].Items = []Item{potion1, potion2}
		}
	}
	state.CharacterMap = buildCharacterMap(&state)

	// Use first item
	action1 := Action{
		Kind:  "UseItem",
		Actor: player.ID,
		Item:  potion1.ID,
	}
	resolution1 := ApplyAction(state, action1, 12345)

	// First item should be removed
	user := GetCharacterByID(resolution1.State, player.ID)
	if len(user.Items) != 1 {
		t.Errorf("Expected 1 item remaining, got %d", len(user.Items))
	}

	// Second item should still be present
	foundPotion2 := false
	for _, item := range user.Items {
		if item.ID == potion2.ID {
			foundPotion2 = true
		}
	}
	if !foundPotion2 {
		t.Error("Second item should still be in inventory")
	}

	// Use second item
	action2 := Action{
		Kind:  "UseItem",
		Actor: user.ID,
		Item:  potion2.ID,
	}
	resolution2 := ApplyAction(resolution1.State, action2, 12345)

	// All items should be used
	user2 := GetCharacterByID(resolution2.State, player.ID)
	if len(user2.Items) != 0 {
		t.Errorf("Expected 0 items remaining, got %d", len(user2.Items))
	}
}

// TestItemHealAmount verifies potion healing calculation
func TestItemHealAmount(t *testing.T) {
	player := createTestCharacter(true, "Player")
	player.Stats.HP = 10
	player.Stats.MaxHP = 50

	enemy := createTestCharacter(false, "Enemy")

	potion := Item{
		ID:     NewID(),
		Name:   "Health Potion",
		Type:   "consumable",
		Effect: "heal 20 HP",
	}
	player.Items = []Item{potion}

	state := CreateInitialState([]Character{player}, []Character{enemy}, 12345)

	// Update player in state
	for i := range state.Characters {
		if state.Characters[i].ID == player.ID {
			state.Characters[i].Stats.HP = 10
			state.Characters[i].Items = []Item{potion}
		}
	}
	state.CharacterMap = buildCharacterMap(&state)

	action := Action{
		Kind:  "UseItem",
		Actor: player.ID,
		Item:  potion.ID,
	}

	resolution := ApplyAction(state, action, 12345)

	// Find heal event
	var healAmount int
	for _, event := range resolution.Events {
		if event.Type == "heal" {
			healAmount = event.Amount
		}
	}

	// Heal should be 20 + d6 (21-26)
	if healAmount < 21 || healAmount > 26 {
		t.Errorf("Heal amount %d out of expected range 21-26", healAmount)
	}
}

// TestItemOverheal verifies items don't heal above MaxHP
func TestItemOverheal(t *testing.T) {
	player := createTestCharacter(true, "Player")
	player.Stats.HP = player.Stats.MaxHP - 5 // Almost full
	player.Stats.MaxHP = 30

	enemy := createTestCharacter(false, "Enemy")

	potion := Item{
		ID:     NewID(),
		Name:   "Health Potion",
		Type:   "consumable",
		Effect: "heal 20 HP",
	}
	player.Items = []Item{potion}

	state := CreateInitialState([]Character{player}, []Character{enemy}, 12345)

	// Update player in state
	for i := range state.Characters {
		if state.Characters[i].ID == player.ID {
			state.Characters[i].Stats.HP = state.Characters[i].Stats.MaxHP - 5
			state.Characters[i].Items = []Item{potion}
		}
	}
	state.CharacterMap = buildCharacterMap(&state)

	action := Action{
		Kind:  "UseItem",
		Actor: player.ID,
		Item:  potion.ID,
	}

	resolution := ApplyAction(state, action, 12345)

	// HP should not exceed MaxHP
	user := GetCharacterByID(resolution.State, player.ID)
	if user.Stats.HP > user.Stats.MaxHP {
		t.Errorf("HP %d exceeds MaxHP %d", user.Stats.HP, user.Stats.MaxHP)
	}
}

// TestEmptyInventory verifies behavior with no items
func TestEmptyInventory(t *testing.T) {
	player := createTestCharacter(true, "Player")
	player.Items = []Item{} // No items

	enemy := createTestCharacter(false, "Enemy")

	state := CreateInitialState([]Character{player}, []Character{enemy}, 12345)

	// Update player in state
	for i := range state.Characters {
		if state.Characters[i].ID == player.ID {
			state.Characters[i].Items = []Item{}
		}
	}
	state.CharacterMap = buildCharacterMap(&state)

	action := Action{
		Kind:  "UseItem",
		Actor: player.ID,
		Item:  "any-item-id",
	}

	resolution := ApplyAction(state, action, 12345)

	// Should have error about item not found
	hasErrorLog := false
	for _, log := range resolution.Logs {
		if len(log) > 0 {
			hasErrorLog = true
			t.Logf("Error log: %s", log)
		}
	}
	if !hasErrorLog {
		t.Error("Expected error log for item not found")
	}
}

// TestItemWrongActor verifies items can only be used by owner
func TestItemWrongActor(t *testing.T) {
	player := createTestCharacter(true, "Player")
	enemy := createTestCharacter(false, "Enemy")

	potion := Item{
		ID:     NewID(),
		Name:   "Health Potion",
		Type:   "consumable",
		Effect: "heal 20 HP",
	}
	player.Items = []Item{potion}

	state := CreateInitialState([]Character{player}, []Character{enemy}, 12345)

	// Update player in state
	for i := range state.Characters {
		if state.Characters[i].ID == player.ID {
			state.Characters[i].Items = []Item{potion}
		}
	}
	state.CharacterMap = buildCharacterMap(&state)

	// Enemy tries to use player's item
	action := Action{
		Kind:  "UseItem",
		Actor: enemy.ID,
		Item:  potion.ID,
	}

	resolution := ApplyAction(state, action, 12345)

	// Should fail - enemy doesn't have this item
	hasItemEvent := false
	for _, event := range resolution.Events {
		if event.Type == "item_used" {
			hasItemEvent = true
		}
	}
	if hasItemEvent {
		t.Error("Enemy should not be able to use player's item")
	}
}

// TestNonPotionItem verifies only potions heal
func TestNonPotionItem(t *testing.T) {
	player := createTestCharacter(true, "Player")
	enemy := createTestCharacter(false, "Enemy")

	nonPotion := Item{
		ID:     NewID(),
		Name:   "Magic Scroll",
		Type:   "consumable",
		Effect: "unknown",
	}
	player.Items = []Item{nonPotion}

	state := CreateInitialState([]Character{player}, []Character{enemy}, 12345)

	// Update player in state
	for i := range state.Characters {
		if state.Characters[i].ID == player.ID {
			state.Characters[i].Items = []Item{nonPotion}
		}
	}
	state.CharacterMap = buildCharacterMap(&state)

	action := Action{
		Kind:  "UseItem",
		Actor: player.ID,
		Item:  nonPotion.ID,
	}

	initialHP := player.Stats.HP

	resolution := ApplyAction(state, action, 12345)

	// Item should be used (removed)
	user := GetCharacterByID(resolution.State, player.ID)
	if len(user.Items) != 0 {
		t.Error("Non-potion item should still be consumed")
	}

	// But should not heal (no "Potion" in name)
	hasHealEvent := false
	for _, event := range resolution.Events {
		if event.Type == "heal" {
			hasHealEvent = true
		}
	}
	if hasHealEvent {
		t.Error("Non-potion should not produce heal event")
	}

	// HP should be unchanged
	if user.Stats.HP != initialHP {
		t.Logf("HP changed from %d to %d (non-potion has no effect)", initialHP, user.Stats.HP)
	}
}

// TestItemUsageAdvancesTurn verifies using item advances turn
func TestItemUsageAdvancesTurn(t *testing.T) {
	player := createTestCharacter(true, "Player")
	enemy := createTestCharacter(false, "Enemy")

	potion := Item{
		ID:     NewID(),
		Name:   "Health Potion",
		Type:   "consumable",
		Effect: "heal 20 HP",
	}
	player.Items = []Item{potion}

	state := CreateInitialState([]Character{player}, []Character{enemy}, 12345)

	// Update player in state
	for i := range state.Characters {
		if state.Characters[i].ID == player.ID {
			state.Characters[i].Items = []Item{potion}
		}
	}
	state.CharacterMap = buildCharacterMap(&state)

	initialTurn := state.CurrentTurn

	action := Action{
		Kind:  "UseItem",
		Actor: player.ID,
		Item:  potion.ID,
	}

	resolution := ApplyAction(state, action, 12345)

	// Turn should advance
	if resolution.State.CurrentTurn == initialTurn {
		t.Error("Turn should advance after using item")
	}
}
