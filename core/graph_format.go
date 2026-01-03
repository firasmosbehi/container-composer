package core

import (
	"fmt"
	"sort"
	"strings"
)

// ASCIIOptions configures ASCII output formatting
type ASCIIOptions struct {
	ShowNetworks     bool
	ShowVolumes      bool
	ShowHealthChecks bool
}

// DOTOptions configures DOT output formatting
type DOTOptions struct {
	ShowNetworks     bool
	ShowVolumes      bool
	ShowHealthChecks bool
	HighlightCycles  bool
}

// FormatASCII generates ASCII tree representation of the dependency graph
func (g *DependencyGraph) FormatASCII(options ASCIIOptions) string {
	var builder strings.Builder

	builder.WriteString("Dependency Graph\n")
	builder.WriteString(strings.Repeat("=", 80) + "\n\n")

	// Get root services (services with no dependencies)
	rootServices := g.GetRootServices()
	visited := make(map[string]bool)

	// Render each root service and its dependency tree
	for _, serviceName := range rootServices {
		g.renderServiceTree(&builder, serviceName, "", true, visited, options, 0)
	}

	// Render orphan services (not in dependency chains)
	orphans := []string{}
	for serviceName := range g.Services {
		if !visited[serviceName] {
			orphans = append(orphans, serviceName)
		}
	}
	sort.Strings(orphans)
	for _, serviceName := range orphans {
		g.renderServiceTree(&builder, serviceName, "", true, visited, options, 0)
	}

	// Show circular dependencies
	if len(g.CircularDeps) > 0 {
		builder.WriteString("\n")
		builder.WriteString("⚠️  Circular Dependencies Detected:\n")
		for _, cycle := range g.CircularDeps {
			builder.WriteString("    " + strings.Join(cycle, " → ") + "\n")
		}
	}

	return builder.String()
}

// renderServiceTree renders a service and its dependencies as a tree
func (g *DependencyGraph) renderServiceTree(
	builder *strings.Builder,
	serviceName string,
	prefix string,
	isLast bool,
	visited map[string]bool,
	options ASCIIOptions,
	depth int,
) {
	if depth > 20 {
		return // Prevent infinite recursion
	}

	node, exists := g.Services[serviceName]
	if !exists {
		return
	}

	// If already visited and depth > 0, show reference and return
	if visited[serviceName] && depth > 0 {
		marker := "├── "
		if isLast {
			marker = "└── "
		}
		builder.WriteString(prefix + marker + "◆ " + serviceName + " (see above)\n")
		return
	}

	visited[serviceName] = true

	// Draw service node
	marker := "├── "
	if isLast {
		marker = "└── "
	}

	healthIcon := ""
	if options.ShowHealthChecks && node.HasHealthCheck {
		healthIcon = " ⚡"
	}

	// Root services have no prefix marker
	if depth == 0 {
		builder.WriteString("◆ " + serviceName + healthIcon + "\n")
	} else {
		builder.WriteString(prefix + marker + "◆ " + serviceName + healthIcon + "\n")
	}

	// Calculate new prefix
	newPrefix := prefix
	if depth > 0 {
		if isLast {
			newPrefix += "    "
		} else {
			newPrefix += "│   "
		}
	}

	// Count total items to render
	totalItems := len(node.DependsOn)
	if options.ShowNetworks {
		totalItems += len(node.Networks)
	}
	if options.ShowVolumes {
		totalItems += len(node.Volumes)
	}

	currentItem := 0

	// Render dependencies
	for i, dep := range node.DependsOn {
		currentItem++
		isLastItem := currentItem == totalItems

		// Add "depends_on:" label before first dependency
		if i == 0 && depth > 0 {
			builder.WriteString(newPrefix + "├── depends_on:\n")
		}

		depPrefix := newPrefix
		if i == 0 && depth > 0 {
			depPrefix += "│   "
		}

		g.renderServiceTree(builder, dep.Name, depPrefix, isLastItem && i == len(node.DependsOn)-1, visited, options, depth+1)
	}

	// Render network connections
	if options.ShowNetworks && len(node.Networks) > 0 {
		for _, network := range node.Networks {
			currentItem++
			isLastItem := currentItem == totalItems

			netMarker := "├── "
			if isLastItem {
				netMarker = "└── "
			}

			peers := ""
			if networkPeers, ok := node.NetworkPeers[network]; ok && len(networkPeers) > 0 {
				peerNames := []string{}
				for _, peer := range networkPeers {
					if peer.Name != serviceName {
						peerNames = append(peerNames, peer.Name)
					}
				}
				if len(peerNames) > 0 {
					peers = " (shared with: " + strings.Join(peerNames, ", ") + ")"
				}
			}

			builder.WriteString(newPrefix + netMarker + "🌐 network: " + network + peers + "\n")
		}
	}

	// Render volumes
	if options.ShowVolumes && len(node.Volumes) > 0 {
		for _, volume := range node.Volumes {
			currentItem++
			isLastItem := currentItem == totalItems

			volMarker := "├── "
			if isLastItem {
				volMarker = "└── "
			}

			builder.WriteString(newPrefix + volMarker + "💾 volume: " + volume + "\n")
		}
	}
}

// FormatDOT generates Graphviz DOT format representation
func (g *DependencyGraph) FormatDOT(options DOTOptions) string {
	var builder strings.Builder

	builder.WriteString("digraph dependencies {\n")
	builder.WriteString("  rankdir=LR;\n")
	builder.WriteString("  node [shape=box, style=rounded];\n\n")

	// Get sorted service names for consistent output
	serviceNames := []string{}
	for name := range g.Services {
		serviceNames = append(serviceNames, name)
	}
	sort.Strings(serviceNames)

	// Define nodes
	builder.WriteString("  // Nodes\n")
	for _, name := range serviceNames {
		node := g.Services[name]
		attrs := []string{}

		if node.HasHealthCheck && options.ShowHealthChecks {
			attrs = append(attrs, "color=green", "penwidth=2")
		}

		// Check if part of circular dependency
		if options.HighlightCycles && g.isInCycle(name) {
			attrs = append(attrs, "color=red", "penwidth=3")
		}

		label := name
		if options.ShowHealthChecks && node.HasHealthCheck {
			label += "\\n⚡HealthCheck"
		}

		attrStr := ""
		if len(attrs) > 0 {
			attrStr = " [" + strings.Join(attrs, ", ") + ", label=\"" + label + "\"]"
		} else {
			attrStr = " [label=\"" + label + "\"]"
		}

		builder.WriteString(fmt.Sprintf("  \"%s\"%s;\n", name, attrStr))
	}

	builder.WriteString("\n")

	// Define edges for depends_on
	builder.WriteString("  // Dependencies\n")
	for _, name := range serviceNames {
		node := g.Services[name]
		for _, dep := range node.DependsOn {
			style := ""
			if options.HighlightCycles && g.isCycleEdge(name, dep.Name) {
				style = " [color=red, penwidth=2, label=\"CYCLE\"]"
			} else {
				style = " [color=blue]"
			}
			builder.WriteString(fmt.Sprintf("  \"%s\" -> \"%s\"%s;\n", name, dep.Name, style))
		}
	}

	// Network relationships (as subgraphs)
	if options.ShowNetworks {
		networkGroups := g.getNetworkGroups()
		if len(networkGroups) > 0 {
			builder.WriteString("\n  // Network relationships\n")
			for network, services := range networkGroups {
				if len(services) > 1 {
					// Create safe cluster name (replace invalid chars)
					clusterName := strings.ReplaceAll(network, "-", "_")
					clusterName = strings.ReplaceAll(clusterName, ".", "_")

					builder.WriteString(fmt.Sprintf("  subgraph cluster_%s {\n", clusterName))
					builder.WriteString(fmt.Sprintf("    label=\"Network: %s\";\n", network))
					builder.WriteString("    style=dashed;\n")
					builder.WriteString("    color=gray;\n")
					for _, svc := range services {
						builder.WriteString(fmt.Sprintf("    \"%s\";\n", svc))
					}
					builder.WriteString("  }\n")
				}
			}
		}
	}

	builder.WriteString("}\n")

	return builder.String()
}

// isInCycle checks if a service is part of any circular dependency
func (g *DependencyGraph) isInCycle(serviceName string) bool {
	for _, cycle := range g.CircularDeps {
		for _, name := range cycle {
			if name == serviceName {
				return true
			}
		}
	}
	return false
}

// isCycleEdge checks if an edge is part of a circular dependency
func (g *DependencyGraph) isCycleEdge(from, to string) bool {
	for _, cycle := range g.CircularDeps {
		for i := 0; i < len(cycle)-1; i++ {
			if cycle[i] == from && cycle[i+1] == to {
				return true
			}
		}
	}
	return false
}

// getNetworkGroups returns a map of network name to services on that network
func (g *DependencyGraph) getNetworkGroups() map[string][]string {
	groups := make(map[string][]string)
	for name, node := range g.Services {
		for _, network := range node.Networks {
			groups[network] = append(groups[network], name)
		}
	}
	// Sort services in each group for consistency
	for network := range groups {
		sort.Strings(groups[network])
	}
	return groups
}
