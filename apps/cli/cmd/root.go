package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var (
	saveDir      string
	scenarioDir  string
	autoSave     bool
)

// rootCmd represents the base command
var rootCmd = &cobra.Command{
	Use:   "smoldungeon",
	Short: "SmolDungeon - A standalone turn-based D&D combat game",
	Long: `SmolDungeon is a terminal-based turn-based combat game with no server required.

Experience D&D-style combat with beautiful terminal graphics,
interactive menus, and real-time combat against AI opponents.

All game data is stored locally - no internet connection needed!`,
	Run: func(cmd *cobra.Command, args []string) {
		// Show help if no subcommand is provided
		cmd.Help()
	},
}

// Execute runs the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	// Get home directory for default save location
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}

	defaultSaveDir := filepath.Join(homeDir, ".smoldungeon", "saves")
	defaultScenarioDir := "./scenarios"

	// Global flags
	rootCmd.PersistentFlags().StringVar(&saveDir, "save-dir", defaultSaveDir, "Directory for save files")
	rootCmd.PersistentFlags().StringVar(&scenarioDir, "scenario-dir", defaultScenarioDir, "Directory for scenario files")
	rootCmd.PersistentFlags().BoolVar(&autoSave, "auto-save", true, "Automatically save game progress")
}
