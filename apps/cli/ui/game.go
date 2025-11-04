package ui

import (
	"fmt"
	"strings"

	"smoldungeon-cli/api"
	"smoldungeon-cli/types"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ActionType represents different action types
type ActionType int

const (
	ActionAttack ActionType = iota
	ActionDefend
	ActionFlee
	ActionAbility
	ActionItem
)

// GameModel represents the Bubble Tea model for the game
type GameModel struct {
	client      *api.Client
	sessionID   string
	state       types.GameState
	logs        []string
	cursor      int
	menuOptions []string
	width       int
	height      int
	err         error
	quitting    bool
	actionMode  bool // true when selecting action details (like target)
	targetList  []types.Character
	targetIndex int
}

// NewGameModel creates a new game model
func NewGameModel(client *api.Client, sessionID string, initialState types.GameState) GameModel {
	return GameModel{
		client:      client,
		sessionID:   sessionID,
		state:       initialState,
		logs:        []string{},
		cursor:      0,
		menuOptions: []string{"⚔️  Attack", "🛡️  Defend", "🏃 Flee"},
		width:       80,
		height:      24,
	}
}

// Init initializes the model
func (m GameModel) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the model
func (m GameModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit

		case "up", "k":
			if m.actionMode {
				if m.targetIndex > 0 {
					m.targetIndex--
				}
			} else {
				if m.cursor > 0 {
					m.cursor--
				}
			}

		case "down", "j":
			if m.actionMode {
				if m.targetIndex < len(m.targetList)-1 {
					m.targetIndex++
				}
			} else {
				if m.cursor < len(m.menuOptions)-1 {
					m.cursor++
				}
			}

		case "enter", " ":
			if m.actionMode {
				return m.executeAction()
			}
			return m.selectAction()
		}

	case actionResultMsg:
		m.state = msg.state
		m.logs = append(m.logs, msg.logs...)
		// Keep only last 10 log entries
		if len(m.logs) > 10 {
			m.logs = m.logs[len(m.logs)-10:]
		}
		m.actionMode = false
		m.targetList = nil
		m.targetIndex = 0

		// Check if game is over
		if m.state.IsComplete {
			return m, tea.Quit
		}
		return m, nil

	case errMsg:
		m.err = msg.err
		return m, nil
	}

	return m, nil
}

// selectAction handles action selection
func (m GameModel) selectAction() (tea.Model, tea.Cmd) {
	switch m.cursor {
	case 0: // Attack
		// Get alive enemies to select target
		enemies := getEnemies(m.state)
		if len(enemies) == 0 {
			m.err = fmt.Errorf("no enemies to attack")
			return m, nil
		}
		m.actionMode = true
		m.targetList = enemies
		m.targetIndex = 0
		return m, nil

	case 1: // Defend
		return m.executeDefend()

	case 2: // Flee
		return m.executeFlee()
	}

	return m, nil
}

// executeAction executes the selected action
func (m GameModel) executeAction() (tea.Model, tea.Cmd) {
	currentChar := getCurrentCharacter(m.state)
	if currentChar == nil {
		return m, nil
	}

	switch m.cursor {
	case 0: // Attack
		if m.targetIndex < 0 || m.targetIndex >= len(m.targetList) {
			return m, nil
		}

		target := m.targetList[m.targetIndex]
		var weapon types.Weapon
		if len(currentChar.Weapons) > 0 {
			weapon = currentChar.Weapons[0]
		}

		action := types.Action{
			Kind:     "Attack",
			Attacker: currentChar.ID,
			Target:   target.ID,
			Weapon:   weapon.ID,
		}

		return m, executeActionCmd(m.client, m.sessionID, action)
	}

	return m, nil
}

// executeDefend executes a defend action
func (m GameModel) executeDefend() (tea.Model, tea.Cmd) {
	currentChar := getCurrentCharacter(m.state)
	if currentChar == nil {
		return m, nil
	}

	action := types.Action{
		Kind:  "Defend",
		Actor: currentChar.ID,
	}

	return m, executeActionCmd(m.client, m.sessionID, action)
}

// executeFlee executes a flee action
func (m GameModel) executeFlee() (tea.Model, tea.Cmd) {
	currentChar := getCurrentCharacter(m.state)
	if currentChar == nil {
		return m, nil
	}

	action := types.Action{
		Kind:  "Flee",
		Actor: currentChar.ID,
	}

	return m, executeActionCmd(m.client, m.sessionID, action)
}

// View renders the UI
func (m GameModel) View() string {
	if m.quitting {
		if m.state.IsComplete {
			return m.renderGameOver()
		}
		return "Thanks for playing! ⚔️\n"
	}

	var b strings.Builder

	// Title
	title := TitleStyle.Width(m.width).Render(fmt.Sprintf("⚔️  SMOLDUNGEON - Round %d  ⚔️", m.state.Round))
	b.WriteString(title + "\n")

	// Game state panels
	b.WriteString(m.renderCharacters())
	b.WriteString("\n")

	// Combat log
	if len(m.logs) > 0 {
		b.WriteString(m.renderCombatLog())
		b.WriteString("\n")
	}

	// Action menu or target selection
	if m.actionMode {
		b.WriteString(m.renderTargetSelection())
	} else {
		b.WriteString(m.renderActionMenu())
	}

	// Error display
	if m.err != nil {
		b.WriteString("\n")
		b.WriteString(ErrorMessageStyle.Render(fmt.Sprintf("Error: %v", m.err)))
		m.err = nil // Clear error after displaying
	}

	// Help
	b.WriteString("\n\n")
	b.WriteString(HelpStyle.Render("↑/↓: Navigate • Enter: Select • q: Quit"))

	return BaseStyle.Render(b.String())
}

// renderCharacters renders the character panels
func (m GameModel) renderCharacters() string {
	players := getPlayers(m.state)
	enemies := getEnemies(m.state)

	var b strings.Builder

	// Players section
	b.WriteString(PlayerNameStyle.Render("🗡️  YOUR PARTY") + "\n")
	for _, char := range players {
		b.WriteString(m.renderCharacter(char, true))
	}

	b.WriteString("\n")

	// Enemies section
	b.WriteString(EnemyNameStyle.Render("👹 ENEMIES") + "\n")
	for _, char := range enemies {
		b.WriteString(m.renderCharacter(char, false))
	}

	return b.String()
}

// renderCharacter renders a single character card
func (m GameModel) renderCharacter(char types.Character, isPlayer bool) string {
	currentChar := getCurrentCharacter(m.state)
	isCurrentTurn := currentChar != nil && char.ID == currentChar.ID

	var content strings.Builder

	// Name with turn indicator
	nameStyle := PlayerNameStyle
	if !isPlayer {
		nameStyle = EnemyNameStyle
	}

	name := nameStyle.Render(char.Name)
	if isCurrentTurn {
		name += TurnIndicatorStyle.Render("← CURRENT TURN")
	}
	content.WriteString(name + "\n")

	// Stats
	stats := fmt.Sprintf("%s %s  %s %s  %s %s  %s (%d,%d)",
		StatLabelStyle.Render("HP:"),
		StatValueStyle.Render(fmt.Sprintf("%d/%d", char.Stats.HP, char.Stats.MaxHP)),
		StatLabelStyle.Render("ATK:"),
		StatValueStyle.Render(fmt.Sprint(char.Stats.Attack)),
		StatLabelStyle.Render("DEF:"),
		StatValueStyle.Render(fmt.Sprint(char.Stats.Defense)),
		StatLabelStyle.Render("Pos:"),
		char.Position.X,
		char.Position.Y,
	)
	content.WriteString(stats + "\n")

	// Health bar
	healthBar := RenderHealthBar(char.Stats.HP, char.Stats.MaxHP, 30)
	content.WriteString(healthBar)

	// Weapons
	if len(char.Weapons) > 0 {
		content.WriteString("\n" + StatLabelStyle.Render("Weapon: ") + char.Weapons[0].Name)
	}

	panelStyle := PlayerPanelStyle
	if !isPlayer {
		panelStyle = EnemyPanelStyle
	}

	return panelStyle.Width(m.width - 4).Render(content.String()) + "\n"
}

// renderActionMenu renders the action selection menu
func (m GameModel) renderActionMenu() string {
	currentChar := getCurrentCharacter(m.state)
	if currentChar == nil {
		return ""
	}

	var b strings.Builder
	b.WriteString(TurnIndicatorStyle.Render(fmt.Sprintf("%s's Turn - Choose Action:", currentChar.Name)) + "\n\n")

	for i, option := range m.menuOptions {
		cursor := "  "
		if i == m.cursor {
			cursor = "▶ "
		}

		style := MenuItemStyle
		if i == m.cursor {
			style = SelectedMenuItemStyle
		}

		b.WriteString(cursor + style.Render(option) + "\n")
	}

	return b.String()
}

// renderTargetSelection renders the target selection menu
func (m GameModel) renderTargetSelection() string {
	var b strings.Builder
	b.WriteString(TurnIndicatorStyle.Render("Select Target:") + "\n\n")

	for i, target := range m.targetList {
		cursor := "  "
		if i == m.targetIndex {
			cursor = "▶ "
		}

		style := MenuItemStyle
		if i == m.targetIndex {
			style = SelectedMenuItemStyle
		}

		targetInfo := fmt.Sprintf("%s (HP: %d/%d)", target.Name, target.Stats.HP, target.Stats.MaxHP)
		b.WriteString(cursor + style.Render(targetInfo) + "\n")
	}

	return b.String()
}

// renderCombatLog renders the combat log
func (m GameModel) renderCombatLog() string {
	var b strings.Builder
	b.WriteString(StatLabelStyle.Render("📜 Combat Log:") + "\n")

	// Show last few logs
	startIdx := 0
	if len(m.logs) > 5 {
		startIdx = len(m.logs) - 5
	}

	for i := startIdx; i < len(m.logs); i++ {
		b.WriteString(LogStyle.Render("  " + m.logs[i]) + "\n")
	}

	return b.String()
}

// renderGameOver renders the game over screen
func (m GameModel) renderGameOver() string {
	var b strings.Builder

	if m.state.Winner != nil {
		if *m.state.Winner == "player" {
			title := VictoryStyle.Width(m.width).Render("🎉 VICTORY! 🎉")
			b.WriteString("\n" + title + "\n\n")
		} else {
			title := DefeatStyle.Width(m.width).Render("💀 DEFEAT 💀")
			b.WriteString("\n" + title + "\n\n")
		}
	} else {
		title := lipgloss.NewStyle().Bold(true).Align(lipgloss.Center).Render("DRAW")
		b.WriteString("\n" + title + "\n\n")
	}

	b.WriteString(RoundStyle.Render(fmt.Sprintf("Battle lasted %d rounds", m.state.Round)) + "\n\n")
	b.WriteString("Thanks for playing SmolDungeon! ⚔️\n")

	return b.String()
}

// Helper functions

func getCurrentCharacter(state types.GameState) *types.Character {
	if len(state.TurnOrder) == 0 {
		return nil
	}

	currentID := state.TurnOrder[state.CurrentTurn]
	for _, char := range state.Characters {
		if char.ID == currentID {
			return &char
		}
	}
	return nil
}

func getPlayers(state types.GameState) []types.Character {
	var players []types.Character
	for _, char := range state.Characters {
		if char.IsPlayer && char.Stats.HP > 0 {
			players = append(players, char)
		}
	}
	return players
}

func getEnemies(state types.GameState) []types.Character {
	var enemies []types.Character
	for _, char := range state.Characters {
		if !char.IsPlayer && char.Stats.HP > 0 {
			enemies = append(enemies, char)
		}
	}
	return enemies
}

// Messages

type actionResultMsg struct {
	state types.GameState
	logs  []string
}

type errMsg struct {
	err error
}

// Commands

func executeActionCmd(client *api.Client, sessionID string, action types.Action) tea.Cmd {
	return func() tea.Msg {
		resp, err := client.ApplyAction(sessionID, action)
		if err != nil {
			return errMsg{err}
		}
		return actionResultMsg{
			state: resp.State,
			logs:  resp.Logs,
		}
	}
}
