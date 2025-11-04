package cmd

import (
	"fmt"

	"smoldungeon-cli/saves"
	"smoldungeon-cli/ui"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

// savesCmd represents the saves command
var savesCmd = &cobra.Command{
	Use:   "saves",
	Short: "List saved games",
	Long:  `List all saved game files with their details.`,
	RunE:  runSaves,
}

func init() {
	rootCmd.AddCommand(savesCmd)
}

func runSaves(cmd *cobra.Command, args []string) error {
	fmt.Println(ui.TitleStyle.Render("💾 Saved Games"))
	fmt.Println()

	saveList, err := saves.ListSaves(saveDir)
	if err != nil {
		return fmt.Errorf("failed to list saves: %w", err)
	}

	if len(saveList) == 0 {
		fmt.Println(ui.ErrorMessageStyle.Render("No save files found"))
		fmt.Println()
		fmt.Println(ui.HelpStyle.Render("Play a game to create a save file"))
		return nil
	}

	// Create styled list
	for i, save := range saveList {
		// Status indicator
		status := "⏸️"
		statusText := "In Progress"
		if save.IsComplete {
			status = "✅"
			statusText = "Complete"
		}

		// Save info panel
		var content string
		content += lipgloss.NewStyle().Bold(true).Foreground(ui.ColorPrimary).Render(save.Name) + "\n"
		content += lipgloss.NewStyle().Foreground(ui.ColorTextMuted).Render(fmt.Sprintf("Scenario: %s", save.ScenarioName)) + "\n"
		content += lipgloss.NewStyle().Foreground(ui.ColorTextMuted).Render(fmt.Sprintf("Round: %d  |  %s %s", save.Round, status, statusText)) + "\n"
		content += lipgloss.NewStyle().Foreground(ui.ColorTextMuted).Italic(true).Render(fmt.Sprintf("Saved: %s", save.SavedAt.Format("2006-01-02 15:04:05")))

		panel := ui.PanelStyle.Render(content)
		fmt.Println(panel)

		if i < len(saveList)-1 {
			fmt.Println()
		}
	}

	fmt.Println()
	fmt.Println(ui.HelpStyle.Render("Use 'smoldungeon continue --save <name>' to load a save"))

	return nil
}
