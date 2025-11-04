package main

import (
	"fmt"
	"strings"
)

// HTML rendering helper functions for templates

// renderCharacters renders characters as HTML
func renderCharacters(state State) string {
	var html strings.Builder
	for _, char := range state.Characters {
		class := "character"
		if char.IsPlayer {
			class += " player"
		} else {
			class += " enemy"
		}
		html.WriteString(fmt.Sprintf(`<div class="%s">
	           <strong>%s</strong><br>
	           HP: %d/%d<br>
	           Position: (%d, %d)
	       </div>`, class, char.Name, char.Stats.HP, char.Stats.MaxHP, char.Position.X, char.Position.Y))
	}
	return html.String()
}

// renderCombatMap renders the combat map as HTML
func renderCombatMap(state State) string {
	var html strings.Builder

	// Create 5x5 grid centered on (0,0)
	for y := -2; y <= 2; y++ {
		for x := -2; x <= 2; x++ {
			html.WriteString(`<div class="character">`)

			// Find character at this position
			var charAtPos *Character
			for i := range state.Characters {
				if state.Characters[i].Position.X == x && state.Characters[i].Position.Y == y {
					charAtPos = &state.Characters[i]
					break
				}
			}

			if charAtPos != nil {
				class := "character"
				if charAtPos.IsPlayer {
					class += " player"
				} else {
					class += " enemy"
				}

				if charAtPos.Stats.HP == 0 {
					class += " dead"
				} else if charAtPos.Stats.HP < charAtPos.Stats.MaxHP/3 {
					class += " low-health"
				}

				html.WriteString(fmt.Sprintf(`<div>%s</div>`, charAtPos.Name))
				html.WriteString(fmt.Sprintf(`<div class="health-bar"><div class="health-fill" style="width: %d%%"></div></div>`,
					(charAtPos.Stats.HP*100)/charAtPos.Stats.MaxHP))
			} else {
				html.WriteString(`<div>·</div>`)
			}

			html.WriteString(`</div>`)
		}
	}

	return html.String()
}

// renderTurnIndicator renders the turn indicator
func renderTurnIndicator(state State) string {
	currentChar := GetCurrentCharacter(state)
	if currentChar == nil {
		return "Unknown Turn"
	}
	return fmt.Sprintf("%s's Turn", currentChar.Name)
}

// renderActionButtons renders action buttons for the current player
func renderActionButtons(state State, isPlayerTurn bool) string {
	if !isPlayerTurn {
		return `<div id="action-buttons" style="display: none;"></div>`
	}

	return `
	<div id="action-buttons">
		<div class="action-buttons">
			<button class="btn btn-attack" onclick="sendAction('attack')">⚔️ Attack</button>
			<button class="btn btn-defend" onclick="sendAction('defend')">🛡️ Defend</button>
			<button class="btn btn-ability" onclick="sendAction('ability')">✨ Ability</button>
			<button class="btn btn-item" onclick="sendAction('item')">🎒 Use Item</button>
			<button class="btn btn-flee" onclick="sendAction('flee')">🏃 Flee</button>
		</div>
	</div>`
}

// renderStateJSON renders state as JSON for JavaScript
func renderStateJSON(state State) string {
	return fmt.Sprintf(`{"round":%d,"currentTurn":%d,"isComplete":%t}`,
		state.Round, state.CurrentTurn, state.IsComplete)
}
