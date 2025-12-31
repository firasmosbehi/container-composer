package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	upDetached bool
	upBuild    bool
)

var upCmd = &cobra.Command{
	Use:   "up [services...]",
	Short: "Start services",
	Long: `Start all services defined in docker-compose.yml or specific services if provided.
This command wraps docker-compose up with better error messages and additional features.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Starting services...")

		if upBuild {
			fmt.Println("Building services before starting...")
		}

		if upDetached {
			fmt.Println("Running in detached mode...")
		}

		// TODO: Implement docker-compose up wrapper
		fmt.Println("Service startup coming soon!")

		if len(args) > 0 {
			fmt.Printf("Services to start: %v\n", args)
		}
	},
}

func init() {
	upCmd.Flags().BoolVarP(&upDetached, "detach", "D", false, "run services in background")
	upCmd.Flags().BoolVarP(&upBuild, "build", "b", false, "build images before starting")

	rootCmd.AddCommand(upCmd)
}