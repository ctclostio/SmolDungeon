package cmd

import (
	"fmt"
	"path/filepath"
	"time"

	"smoldungeon-cli/engine"
	"smoldungeon-cli/ui"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

var (
	scenarioName string
)

// playCmd represents the play command
var playCmd = &cobra.Command{
	Use:   "play",
	Short: "Start a new game",
	Long: `Start a new SmolDungeon game session.

Launch an interactive terminal UI for turn-based combat.
Choose your actions, select targets, and defeat your enemies!

The game runs entirely locally - no server required.`,
	RunE: runPlay,
}

func init() {
	rootCmd.AddCommand(playCmd)
	playCmd.Flags().StringVarP(&scenarioName, "scenario", "s", "goblin-ambush", "Scenario to play")
}

func runPlay(cmd *cobra.Command, args []string) error {
	// Show banner
	fmt.Println(ui.BannerStyle.Width(60).Render(`
⚔️  SMOLDUNGEON  ⚔️

Turn-Based Combat Adventure
	`))
	fmt.Println()

	// Load scenario
	scenarioPath := filepath.Join(scenarioDir, scenarioName+".yaml")
	fmt.Printf("🎮 Loading scenario: %s\n", scenarioName)

	scenario, err := engine.LoadScenario(scenarioPath)
	if err != nil {
		return fmt.Errorf("failed to load scenario: %w", err)
	}

	fmt.Printf("📖 %s\n", scenario.Description)
	fmt.Printf("   %d players vs %d enemies\n\n", len(scenario.Players), len(scenario.Enemies))

	// Create initial state
	seed := time.Now().UnixNano()
	initialState := engine.ConvertScenarioToState(scenario, seed)

	fmt.Println("✅ Game ready!")
	fmt.Println("Press any key to begin...")
	fmt.Scanln()

	// Start Bubble Tea program
	model := ui.NewGameModel(initialState, scenarioName, saveDir, autoSave)
	p := tea.NewProgram(model, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		return fmt.Errorf("error running game: %w", err)
	}

	return nil
}
