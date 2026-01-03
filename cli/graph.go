package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/firasmosbahi/container-composer/core"
	"github.com/spf13/cobra"
)

var (
	graphFormat           string
	graphService          string
	graphOutput           string
	graphShowNetworks     bool
	graphShowVolumes      bool
	graphShowHealthChecks bool
	graphHighlightCycles  bool
)

var graphCmd = &cobra.Command{
	Use:   "graph",
	Short: "Visualize service dependencies and relationships",
	Long: `Generate dependency graph visualizations showing:
  - Service dependencies (depends_on)
  - Network connections
  - Volume sharing
  - Health check status

Supports multiple output formats for different use cases:
  - ascii: Terminal-friendly tree diagram
  - dot: Graphviz DOT format for rendering with graphviz

Examples:
  container-composer graph                           # Show ASCII graph
  container-composer graph --format=dot              # Output DOT format
  container-composer graph --service=api             # Filter by service
  container-composer graph -o graph.dot              # Save to file
  container-composer graph --format=dot | dot -Tpng > graph.png`,
	RunE: runGraph,
}

func init() {
	graphCmd.Flags().StringVarP(&graphFormat, "format", "f", "ascii",
		"output format: ascii or dot")
	graphCmd.Flags().StringVarP(&graphService, "service", "s", "",
		"filter graph to show only this service and its dependencies")
	graphCmd.Flags().StringVarP(&graphOutput, "output", "o", "",
		"output file (default: stdout)")
	graphCmd.Flags().BoolVar(&graphShowNetworks, "networks", true,
		"show network relationships")
	graphCmd.Flags().BoolVar(&graphShowVolumes, "volumes", true,
		"show volume relationships")
	graphCmd.Flags().BoolVar(&graphShowHealthChecks, "health", true,
		"show health check indicators")
	graphCmd.Flags().BoolVar(&graphHighlightCycles, "highlight-cycles", true,
		"highlight circular dependencies (dot format only)")

	rootCmd.AddCommand(graphCmd)
}

func runGraph(cmd *cobra.Command, args []string) error {
	// Check if docker-compose.yml exists
	composePath := "docker-compose.yml"
	if _, err := os.Stat(composePath); os.IsNotExist(err) {
		return fmt.Errorf("docker-compose.yml not found in current directory")
	}

	// Parse compose file
	composeFile, err := core.ParseComposeFile(composePath)
	if err != nil {
		return fmt.Errorf("failed to parse docker-compose.yml: %w", err)
	}

	// Build dependency graph
	graph, err := composeFile.BuildDependencyGraph()
	if err != nil {
		return fmt.Errorf("failed to build dependency graph: %w", err)
	}

	// Filter by service if specified
	if graphService != "" {
		if !composeFile.ServiceExists(graphService) {
			return fmt.Errorf("service '%s' not found", graphService)
		}
		graph, err = graph.FilterByService(graphService, -1) // -1 = unlimited depth
		if err != nil {
			return fmt.Errorf("failed to filter graph: %w", err)
		}
	}

	// Warn about circular dependencies
	if graph.HasCircularDependencies() {
		fmt.Fprintf(os.Stderr, "\n⚠️  WARNING: Circular dependencies detected!\n")
		fmt.Fprintf(os.Stderr, "Docker Compose will handle these, but they may cause issues.\n\n")
		for _, cycle := range graph.CircularDeps {
			fmt.Fprintf(os.Stderr, "  Cycle: %s\n", formatCycle(cycle))
		}
		fmt.Fprintf(os.Stderr, "\n")
	}

	// Generate output based on format
	var output string
	switch graphFormat {
	case "ascii":
		options := core.ASCIIOptions{
			ShowNetworks:     graphShowNetworks,
			ShowVolumes:      graphShowVolumes,
			ShowHealthChecks: graphShowHealthChecks,
		}
		output = graph.FormatASCII(options)

	case "dot":
		options := core.DOTOptions{
			ShowNetworks:     graphShowNetworks,
			ShowVolumes:      graphShowVolumes,
			ShowHealthChecks: graphShowHealthChecks,
			HighlightCycles:  graphHighlightCycles,
		}
		output = graph.FormatDOT(options)

	default:
		return fmt.Errorf("unknown format: %s (supported: ascii, dot)", graphFormat)
	}

	// Write output
	if graphOutput != "" {
		if err := os.WriteFile(graphOutput, []byte(output), 0644); err != nil {
			return fmt.Errorf("failed to write output file: %w", err)
		}
		fmt.Printf("Graph saved to %s\n", graphOutput)
	} else {
		fmt.Print(output)
	}

	return nil
}

func formatCycle(cycle []string) string {
	if len(cycle) == 0 {
		return ""
	}
	return strings.Join(cycle, " → ")
}
