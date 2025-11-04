package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	serverURL string
)

// rootCmd represents the base command
var rootCmd = &cobra.Command{
	Use:   "smoldungeon",
	Short: "SmolDungeon - A turn-based D&D combat game",
	Long: `SmolDungeon is a terminal-based turn-based combat game.

Experience D&D-style combat with beautiful terminal graphics,
interactive menus, and real-time combat against AI opponents.`,
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
	// Global flags
	rootCmd.PersistentFlags().StringVar(&serverURL, "server", "http://localhost:3000", "Game server URL")
}
