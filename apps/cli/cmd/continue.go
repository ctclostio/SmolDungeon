package cmd

import (
	"fmt"

	"smoldungeon-cli/saves"
	"smoldungeon-cli/ui"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

var (
	saveName string
)

// continueCmd represents the continue command
var continueCmd = &cobra.Command{
	Use:   "continue",
	Short: "Continue a saved game",
	Long: `Continue a previously saved game.

Load a saved game state and resume playing from where you left off.`,
	RunE: runContinue,
}

func init() {
	rootCmd.AddCommand(continueCmd)
	continueCmd.Flags().StringVarP(&saveName, "save", "s", "", "Save file name (without .json extension)")
}

func runContinue(cmd *cobra.Command, args []string) error {
	// If no save name provided, try to load the most recent auto-save
	if saveName == "" {
		// List saves and find the most recent
		saveList, err := saves.ListSaves(saveDir)
		if err != nil {
			return fmt.Errorf("failed to list saves: %w", err)
		}

		if len(saveList) == 0 {
			return fmt.Errorf("no save files found")
		}

		// Use the most recent save
		saveName = saveList[0].Name
		fmt.Printf("Loading most recent save: %s\n", saveName)
	}

	// Load the save
	save, err := saves.LoadGame(saveDir, saveName)
	if err != nil {
		return fmt.Errorf("failed to load save: %w", err)
	}

	// Show banner
	fmt.Println(ui.BannerStyle.Width(60).Render(`
⚔️  SMOLDUNGEON  ⚔️

Continue Your Adventure
	`))
	fmt.Println()

	fmt.Printf("📖 Scenario: %s\n", save.ScenarioName)
	fmt.Printf("🔄 Round: %d\n", save.State.Round)
	fmt.Printf("💾 Saved: %s\n\n", save.SavedAt.Format("2006-01-02 15:04:05"))

	if save.State.IsComplete {
		fmt.Println(ui.WarningStyle.Render("⚠️  This game is already complete!"))
		fmt.Println()
	}

	fmt.Println("Press any key to continue...")
	fmt.Scanln()

	// Start Bubble Tea program
	model := ui.NewGameModel(save.State, save.ScenarioName, saveDir, autoSave)
	p := tea.NewProgram(model, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		return fmt.Errorf("error running game: %w", err)
	}

	return nil
}
