package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	shellUser string
	shellCmd  string
)

var shellCommand = &cobra.Command{
	Use:   "shell <service>",
	Short: "Open a shell in a service container",
	Long: `Quick access to a shell in any running service container.
Defaults to /bin/bash but can be customized.`,
	Aliases: []string{"exec", "sh"},
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		serviceName := args[0]

		fmt.Printf("Opening shell in service: %s\n", serviceName)

		if shellUser != "" {
			fmt.Printf("User: %s\n", shellUser)
		}

		if shellCmd != "" {
			fmt.Printf("Shell: %s\n", shellCmd)
		}

		// TODO: Implement docker exec wrapper
		fmt.Println("Shell access coming soon!")
	},
}

func init() {
	shellCommand.Flags().StringVarP(&shellUser, "user", "u", "", "user to run shell as")
	shellCommand.Flags().StringVarP(&shellCmd, "cmd", "c", "/bin/bash", "shell command to run")

	rootCmd.AddCommand(shellCommand)
}