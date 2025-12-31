package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init [project-name]",
	Short: "Initialize a new Container Composer project",
	Long: `Initialize a new Container Composer project with an interactive wizard.
You can choose from various templates like LAMP, MEAN, microservices, etc.`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		projectName := "."
		if len(args) > 0 {
			projectName = args[0]
		}

		fmt.Printf("Initializing Container Composer project in: %s\n", projectName)
		// TODO: Implement interactive wizard
		fmt.Println("Project initialization wizard coming soon!")
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}