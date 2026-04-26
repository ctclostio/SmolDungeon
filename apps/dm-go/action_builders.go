package main

func createActionFromClientAction(actionStr string, currentChar *Character, state State) Action {
	if currentChar == nil {
		return Action{}
	}

	switch actionStr {
	case "attack":
		targetID := firstAliveTargetID(state, !currentChar.IsPlayer)
		if targetID == "" {
			return Action{}
		}

		return Action{
			Kind:     "Attack",
			Attacker: currentChar.ID,
			Target:   targetID,
			Weapon:   firstWeaponID(currentChar),
		}

	case "defend":
		return Action{
			Kind:  "Defend",
			Actor: currentChar.ID,
		}

	case "ability":
		for _, ability := range currentChar.Abilities {
			if currentChar.AbilityCooldowns != nil && currentChar.AbilityCooldowns[string(ability.ID)] > 0 {
				continue
			}

			action := Action{
				Kind:    "Ability",
				Actor:   currentChar.ID,
				Ability: ability.ID,
			}
			if ability.Effect == "damage" || ability.Effect == "debuff" {
				action.Target = firstAliveTargetID(state, !currentChar.IsPlayer)
				if action.Target == "" {
					return Action{}
				}
			}
			return action
		}
		return Action{}

	case "item":
		if len(currentChar.Items) == 0 {
			return Action{}
		}
		return Action{
			Kind:  "UseItem",
			Actor: currentChar.ID,
			Item:  currentChar.Items[0].ID,
		}

	case "flee":
		return Action{
			Kind:  "Flee",
			Actor: currentChar.ID,
		}
	}

	return Action{}
}

func firstAliveTargetID(state State, isPlayer bool) ID {
	for _, char := range state.Characters {
		if char.IsPlayer == isPlayer && char.Stats.HP > 0 {
			return char.ID
		}
	}
	return ""
}

func firstWeaponID(character *Character) ID {
	if len(character.Weapons) == 0 {
		return ""
	}
	return character.Weapons[0].ID
}
