package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	// Version information (to be set at build time)
	Version   = "dev"
	BuildDate = "unknown"

	// Global flags
	verbose bool
	debug   bool
)

var rootCmd = &cobra.Command{
	Use:   "container-composer",
	Short: "A powerful CLI tool for managing Docker Compose projects",
	Long: `Container Composer enhances your Docker Compose workflow with intelligent features,
better debugging capabilities, and an improved developer experience.

It works seamlessly with your existing docker-compose.yml files without modification.`,
	Version: Version,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "debug mode with detailed logging")

	rootCmd.SetVersionTemplate(fmt.Sprintf("container-composer version %s (built on %s)\n", Version, BuildDate))
}

// GetVerbose returns the verbose flag value
func GetVerbose() bool {
	return verbose
}

// GetDebug returns the debug flag value
func GetDebug() bool {
	return debug
}

// Exit prints an error message and exits with status code 1
func Exit(err error) {
	fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	os.Exit(1)
}