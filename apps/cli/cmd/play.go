package cmd

import (
	"fmt"

	"smoldungeon-cli/api"
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
Choose your actions, select targets, and defeat your enemies!`,
	RunE: runPlay,
}

func init() {
	rootCmd.AddCommand(playCmd)
	playCmd.Flags().StringVarP(&scenarioName, "scenario", "s", "goblin-ambush", "Scenario to play")
}

func runPlay(cmd *cobra.Command, args []string) error {
	// Create API client
	client := api.NewClient(serverURL)

	// Show banner
	fmt.Println(ui.BannerStyle.Width(60).Render(`
⚔️  SMOLDUNGEON  ⚔️

Turn-Based Combat Adventure
	`))
	fmt.Println()

	// Create session
	fmt.Printf("🎮 Starting new game with scenario: %s\n", scenarioName)
	resp, err := client.CreateSession(scenarioName)
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}

	fmt.Printf("✅ Game session created: %s\n\n", resp.SessionID)
	fmt.Println("Press any key to begin...")
	fmt.Scanln()

	// Start Bubble Tea program
	model := ui.NewGameModel(client, resp.SessionID, resp.State)
	p := tea.NewProgram(model, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		return fmt.Errorf("error running game: %w", err)
	}

	return nil
}
