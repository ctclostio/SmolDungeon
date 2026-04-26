package main

import "fmt"

// validatePlayerAction verifies that a public client action is legal for the
// current turn. AI actions are generated internally and bypass this validator.
func validatePlayerAction(state State, action Action) error {
	if state.IsComplete {
		return fmt.Errorf("combat is already complete")
	}

	current := GetCurrentCharacter(state)
	if current == nil {
		return fmt.Errorf("no current character")
	}
	if current.Stats.HP <= 0 {
		return fmt.Errorf("current character is defeated")
	}
	if !current.IsPlayer {
		return fmt.Errorf("not a player turn")
	}

	actorID := getActorID(action)
	if actorID == "" {
		return fmt.Errorf("actor is required")
	}
	if actorID != current.ID {
		return fmt.Errorf("action actor must match current turn")
	}

	switch action.Kind {
	case "Attack":
		return validateAttackAction(state, current, action)
	case "Defend", "Flee":
		return nil
	case "Ability":
		return validateAbilityAction(state, current, action)
	case "UseItem":
		return validateItemAction(current, action)
	default:
		return fmt.Errorf("invalid action kind")
	}
}

func validateAttackAction(state State, actor *Character, action Action) error {
	target, err := validateTarget(state, actor, action.Target, true)
	if err != nil {
		return err
	}
	if target.ID == actor.ID {
		return fmt.Errorf("cannot target self with attack")
	}
	if action.Weapon != "" && !characterHasWeapon(actor, action.Weapon) {
		return fmt.Errorf("weapon does not belong to current character")
	}
	return nil
}

func validateAbilityAction(state State, actor *Character, action Action) error {
	ability := findCharacterAbility(actor, action.Ability)
	if ability == nil {
		return fmt.Errorf("ability does not belong to current character")
	}

	switch ability.Effect {
	case "damage", "debuff":
		if _, err := validateTarget(state, actor, action.Target, true); err != nil {
			return err
		}
	case "heal", "buff":
		if action.Target == "" {
			return nil
		}
		if _, err := validateTarget(state, actor, action.Target, false); err != nil {
			return err
		}
	}

	return nil
}

func validateItemAction(actor *Character, action Action) error {
	if !characterHasItem(actor, action.Item) {
		return fmt.Errorf("item does not belong to current character")
	}
	return nil
}

func validateTarget(state State, actor *Character, targetID ID, requireOpponent bool) (*Character, error) {
	if targetID == "" {
		return nil, fmt.Errorf("target is required")
	}

	target := GetCharacterByID(state, targetID)
	if target == nil {
		return nil, fmt.Errorf("target not found")
	}
	if target.Stats.HP <= 0 {
		return nil, fmt.Errorf("target is defeated")
	}
	if requireOpponent && target.IsPlayer == actor.IsPlayer {
		return nil, fmt.Errorf("target must be an opponent")
	}
	if !requireOpponent && target.IsPlayer != actor.IsPlayer {
		return nil, fmt.Errorf("target must be an ally")
	}

	return target, nil
}

func characterHasWeapon(character *Character, weaponID ID) bool {
	for _, weapon := range character.Weapons {
		if weapon.ID == weaponID {
			return true
		}
	}
	return false
}

func findCharacterAbility(character *Character, abilityID ID) *Ability {
	for i := range character.Abilities {
		if character.Abilities[i].ID == abilityID {
			return &character.Abilities[i]
		}
	}
	return nil
}

func characterHasItem(character *Character, itemID ID) bool {
	for _, item := range character.Items {
		if item.ID == itemID {
			return true
		}
	}
	return false
}
