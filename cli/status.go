package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show service health and status",
	Long: `Display a comprehensive dashboard showing the health and status of all services.
Includes resource usage, container state, and health check results.`,
	Aliases: []string{"ps", "list"},
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Fetching service status...")

		// TODO: Implement service health monitoring
		fmt.Println("Service health monitoring coming soon!")
		fmt.Println("\nPlanned features:")
		fmt.Println("  - Container state (running, stopped, restarting)")
		fmt.Println("  - Health check status")
		fmt.Println("  - Resource usage (CPU, memory, network)")
		fmt.Println("  - Port mappings")
		fmt.Println("  - Dependency graph")
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}