package main

import (
	"log"
	"math"
)

// AIDecisionMaker handles intelligent enemy AI behavior
type AIDecisionMaker struct{}

// NewAIDecisionMaker creates a new AI decision maker
func NewAIDecisionMaker() *AIDecisionMaker {
	return &AIDecisionMaker{}
}

// DecideAction determines the best action for an enemy character
// Strategy:
// 1. Self-preservation: Heal if HP < 30%
// 2. Ability usage: Use damage abilities when off cooldown on good targets
// 3. Attack: Target low HP enemies first, then high threat targets
// 4. Never flee or defend (boring for players)
func (ai *AIDecisionMaker) DecideAction(state State, enemyID ID) Action {
	enemy := GetCharacterByID(state, enemyID)
	if enemy == nil || enemy.IsPlayer {
		log.Printf("AI: Invalid enemy ID %s", enemyID)
		return ai.fallbackAction(enemyID, state)
	}

	log.Printf("AI: Making decision for %s (HP: %d/%d)", enemy.Name, enemy.Stats.HP, enemy.Stats.MaxHP)

	// 1. Check if we need healing (HP < 30%)
	hpPercent := float64(enemy.Stats.HP) / float64(enemy.Stats.MaxHP)
	if hpPercent < 0.3 {
		// Try to use healing item first
		if healAction := ai.tryUseHealingItem(enemy, state); healAction != nil {
			log.Printf("AI: %s using healing item (HP: %d/%d)", enemy.Name, enemy.Stats.HP, enemy.Stats.MaxHP)
			return *healAction
		}

		// Try to use healing ability
		if healAction := ai.tryUseHealingAbility(enemy, state); healAction != nil {
			log.Printf("AI: %s using healing ability (HP: %d/%d)", enemy.Name, enemy.Stats.HP, enemy.Stats.MaxHP)
			return *healAction
		}
	}

	// 2. Try to use damage ability on a good target
	if abilityAction := ai.tryUseDamageAbility(enemy, state); abilityAction != nil {
		log.Printf("AI: %s using damage ability", enemy.Name)
		return *abilityAction
	}

	// 3. Attack with best weapon
	attackAction := ai.chooseAttackAction(enemy, state)
	if attackAction != nil {
		target := GetCharacterByID(state, attackAction.Target)
		targetName := "Unknown"
		if target != nil {
			targetName = target.Name
		}
		log.Printf("AI: %s attacking %s", enemy.Name, targetName)
		return *attackAction
	}

	// Fallback: defend if nothing else works
	log.Printf("AI: %s defending (no valid actions)", enemy.Name)
	return Action{
		Kind:  "Defend",
		Actor: enemyID,
	}
}

// tryUseHealingItem attempts to use a healing item
func (ai *AIDecisionMaker) tryUseHealingItem(enemy *Character, state State) *Action {
	// Look for healing potions
	for _, item := range enemy.Items {
		if item.Type == "consumable" && (item.Name == "Health Potion" || item.Name == "Healing Potion") {
			return &Action{
				Kind:  "UseItem",
				Actor: enemy.ID,
				Item:  item.ID,
			}
		}
	}
	return nil
}

// tryUseHealingAbility attempts to use a healing ability
func (ai *AIDecisionMaker) tryUseHealingAbility(enemy *Character, state State) *Action {
	for _, ability := range enemy.Abilities {
		if ability.Effect == "heal" {
			cooldownKey := string(ability.ID)
			if enemy.AbilityCooldowns[cooldownKey] == 0 {
				return &Action{
					Kind:    "Ability",
					Actor:   enemy.ID,
					Ability: ability.ID,
				}
			}
		}
	}
	return nil
}

// tryUseDamageAbility attempts to use a damage ability on the best target
func (ai *AIDecisionMaker) tryUseDamageAbility(enemy *Character, state State) *Action {
	// Find available damage abilities (not on cooldown)
	var bestAbility *Ability
	for i := range enemy.Abilities {
		ability := &enemy.Abilities[i]
		if ability.Effect == "damage" {
			cooldownKey := string(ability.ID)
			if enemy.AbilityCooldowns[cooldownKey] == 0 {
				// Choose the most powerful available ability
				if bestAbility == nil || ability.Power > bestAbility.Power {
					bestAbility = ability
				}
			}
		}
	}

	if bestAbility == nil {
		return nil
	}

	// Select best target for the ability
	target := ai.selectBestTarget(enemy, state)
	if target == nil {
		return nil
	}

	log.Printf("AI: %s will use ability %s (Power: %d) on %s",
		enemy.Name, bestAbility.Name, bestAbility.Power, target.Name)

	return &Action{
		Kind:    "Ability",
		Actor:   enemy.ID,
		Ability: bestAbility.ID,
		Target:  target.ID,
	}
}

// chooseAttackAction creates an attack action with the best weapon and target
func (ai *AIDecisionMaker) chooseAttackAction(enemy *Character, state State) *Action {
	// Select best weapon (highest damage)
	var bestWeapon *Weapon
	for i := range enemy.Weapons {
		weapon := &enemy.Weapons[i]
		if bestWeapon == nil || weapon.Damage > bestWeapon.Damage {
			bestWeapon = weapon
		}
	}

	// No weapon available
	if bestWeapon == nil {
		log.Printf("AI: %s has no weapons", enemy.Name)
		return nil
	}

	// Select best target
	target := ai.selectBestTarget(enemy, state)
	if target == nil {
		log.Printf("AI: %s found no valid targets", enemy.Name)
		return nil
	}

	return &Action{
		Kind:     "Attack",
		Attacker: enemy.ID,
		Target:   target.ID,
		Weapon:   bestWeapon.ID,
	}
}

// selectBestTarget selects the optimal target for an attack or ability
// Priority:
// 1. Low HP targets (< 30% HP) - finish them off
// 2. High threat targets (high attack stat)
// 3. First alive player
func (ai *AIDecisionMaker) selectBestTarget(enemy *Character, state State) *Character {
	var targets []*Character
	var lowHPTargets []*Character
	var highThreatTarget *Character
	maxThreat := 0

	// Collect all valid targets (alive players)
	for i := range state.Characters {
		char := &state.Characters[i]
		if char.IsPlayer && char.Stats.HP > 0 {
			targets = append(targets, char)

			// Check if low HP (< 30%)
			hpPercent := float64(char.Stats.HP) / float64(char.Stats.MaxHP)
			if hpPercent < 0.3 {
				lowHPTargets = append(lowHPTargets, char)
			}

			// Track highest threat
			if char.Stats.Attack > maxThreat {
				maxThreat = char.Stats.Attack
				highThreatTarget = char
			}
		}
	}

	// No valid targets
	if len(targets) == 0 {
		return nil
	}

	// Priority 1: Finish off low HP targets (lowest HP first)
	if len(lowHPTargets) > 0 {
		var lowestHP *Character
		minHP := math.MaxInt32
		for _, target := range lowHPTargets {
			if target.Stats.HP < minHP {
				minHP = target.Stats.HP
				lowestHP = target
			}
		}
		log.Printf("AI: Targeting low HP enemy %s (%d/%d HP)",
			lowestHP.Name, lowestHP.Stats.HP, lowestHP.Stats.MaxHP)
		return lowestHP
	}

	// Priority 2: Target high threat player
	if highThreatTarget != nil {
		log.Printf("AI: Targeting high threat enemy %s (ATK: %d)",
			highThreatTarget.Name, highThreatTarget.Stats.Attack)
		return highThreatTarget
	}

	// Priority 3: Fallback to first target
	log.Printf("AI: Targeting first available enemy %s", targets[0].Name)
	return targets[0]
}

// fallbackAction creates a safe fallback action
func (ai *AIDecisionMaker) fallbackAction(enemyID ID, state State) Action {
	// Try to find any alive player to attack
	for i := range state.Characters {
		char := &state.Characters[i]
		if char.IsPlayer && char.Stats.HP > 0 {
			// Get enemy to find weapon
			enemy := GetCharacterByID(state, enemyID)
			if enemy != nil && len(enemy.Weapons) > 0 {
				return Action{
					Kind:     "Attack",
					Attacker: enemyID,
					Target:   char.ID,
					Weapon:   enemy.Weapons[0].ID,
				}
			}
		}
	}

	// Ultimate fallback: defend
	return Action{
		Kind:  "Defend",
		Actor: enemyID,
	}
}
