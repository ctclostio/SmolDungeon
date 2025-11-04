package cmd

import (
	"fmt"

	"smoldungeon-cli/api"
	"smoldungeon-cli/ui"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

// scenariosCmd represents the scenarios command
var scenariosCmd = &cobra.Command{
	Use:   "scenarios",
	Short: "List available scenarios",
	Long:  `List all available game scenarios from the server.`,
	RunE:  runScenarios,
}

func init() {
	rootCmd.AddCommand(scenariosCmd)
}

func runScenarios(cmd *cobra.Command, args []string) error {
	client := api.NewClient(serverURL)

	fmt.Println(ui.TitleStyle.Render("📚 Available Scenarios"))
	fmt.Println()

	scenarios, err := client.GetScenarios()
	if err != nil {
		return fmt.Errorf("failed to fetch scenarios: %w", err)
	}

	if len(scenarios) == 0 {
		fmt.Println(ui.ErrorMessageStyle.Render("No scenarios available"))
		return nil
	}

	// Create styled list
	listStyle := lipgloss.NewStyle().
		Padding(0, 2).
		Foreground(ui.ColorText)

	for i, scenario := range scenarios {
		bullet := ui.SuccessMessageStyle.Render("▸")
		scenarioText := listStyle.Render(scenario)
		fmt.Printf("%s %s\n", bullet, scenarioText)

		if i < len(scenarios)-1 {
			fmt.Println()
		}
	}

	fmt.Println()
	fmt.Println(ui.HelpStyle.Render("Use 'smoldungeon play --scenario <name>' to start a game"))

	return nil
}
