package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	logsFollow bool
	logsTail   string
	logsFilter string
)

var logsCmd = &cobra.Command{
	Use:   "logs [services...]",
	Short: "View and filter service logs",
	Long: `View logs from services with advanced filtering, search, and highlighting.
This is one of Container Composer's key features - making log analysis much easier.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Fetching logs...")

		if logsFollow {
			fmt.Println("Following log output...")
		}

		if logsTail != "" {
			fmt.Printf("Tailing last %s lines...\n", logsTail)
		}

		if logsFilter != "" {
			fmt.Printf("Filtering logs with pattern: %s\n", logsFilter)
		}

		// TODO: Implement advanced log aggregation
		fmt.Println("Advanced log viewing coming soon!")

		if len(args) > 0 {
			fmt.Printf("Services: %v\n", args)
		}
	},
}

func init() {
	logsCmd.Flags().BoolVarP(&logsFollow, "follow", "f", false, "follow log output")
	logsCmd.Flags().StringVarP(&logsTail, "tail", "t", "", "number of lines to show from the end")
	logsCmd.Flags().StringVar(&logsFilter, "filter", "", "regex pattern to filter logs")

	rootCmd.AddCommand(logsCmd)
}