package cli

import (
	"github.com/firasmosbahi/container-composer/tui"
	"github.com/spf13/cobra"
)

var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Launch interactive terminal UI",
	Long: `Launch the interactive Terminal User Interface (TUI) for Container Composer.

The TUI provides a visual menu-driven interface for:
  - Initializing new projects
  - Adding services, networks, and volumes
  - Monitoring running services
  - Managing multiple projects

Press 'q' or 'ctrl+c' to exit at any time.`,
	RunE: runTUI,
}

func init() {
	rootCmd.AddCommand(tuiCmd)
}

func runTUI(cmd *cobra.Command, args []string) error {
	app := tui.NewApp()
	return app.Run()
}
