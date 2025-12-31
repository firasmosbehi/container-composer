package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	downVolumes bool
	downOrphans bool
)

var downCmd = &cobra.Command{
	Use:   "down",
	Short: "Stop and remove services",
	Long: `Stop and remove all containers, networks, and optionally volumes.
This command wraps docker-compose down with better feedback.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Stopping services...")

		if downVolumes {
			fmt.Println("Removing volumes...")
		}

		if downOrphans {
			fmt.Println("Removing orphan containers...")
		}

		// TODO: Implement docker-compose down wrapper
		fmt.Println("Service shutdown coming soon!")
	},
}

func init() {
	downCmd.Flags().BoolVarP(&downVolumes, "volumes", "V", false, "remove named volumes")
	downCmd.Flags().BoolVar(&downOrphans, "remove-orphans", false, "remove orphan containers")

	rootCmd.AddCommand(downCmd)
}