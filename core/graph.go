package core

import (
	"fmt"
	"sort"
)

// DependencyGraph represents the complete dependency graph
type DependencyGraph struct {
	Services         map[string]*ServiceNode
	CircularDeps     [][]string
	TopologicalOrder []string
}

// ServiceNode represents a service with all its relationships
type ServiceNode struct {
	Name        string
	Service     *Service
	DependsOn   []*ServiceNode
	DependedBy  []*ServiceNode
	Networks    []string
	NetworkPeers map[string][]*ServiceNode
	Volumes     []string
	VolumePeers  map[string][]*ServiceNode
	HasHealthCheck bool
	HealthCheck *HealthCheck
}

// RelationshipType categorizes different types of relationships
type RelationshipType int

const (
	RelationshipDependsOn RelationshipType = iota
	RelationshipNetwork
	RelationshipVolume
	RelationshipHealthCheck
)

// Relationship represents a connection between two services
type Relationship struct {
	From     string
	To       string
	Type     RelationshipType
	Metadata string
}

// BuildDependencyGraph creates a complete dependency graph
func (cf *ComposeFile) BuildDependencyGraph() (*DependencyGraph, error) {
	graph := &DependencyGraph{
		Services: make(map[string]*ServiceNode),
	}

	// Step 1: Create ServiceNode for each service
	for name, service := range cf.Services {
		serviceCopy := service
		node := &ServiceNode{
			Name:           name,
			Service:        &serviceCopy,
			DependsOn:      []*ServiceNode{},
			DependedBy:     []*ServiceNode{},
			Networks:       service.Networks,
			NetworkPeers:   make(map[string][]*ServiceNode),
			Volumes:        service.Volumes,
			VolumePeers:    make(map[string][]*ServiceNode),
			HasHealthCheck: service.HealthCheck != nil,
			HealthCheck:    service.HealthCheck,
		}
		graph.Services[name] = node
	}

	// Step 2: Build dependency relationships
	for name, node := range graph.Services {
		for _, depName := range node.Service.DependsOn {
			depNode, exists := graph.Services[depName]
			if !exists {
				return nil, fmt.Errorf("service '%s' depends on non-existent service '%s'", name, depName)
			}
			node.DependsOn = append(node.DependsOn, depNode)
			depNode.DependedBy = append(depNode.DependedBy, node)
		}
	}

	// Step 3: Build network relationships
	graph.buildNetworkRelationships()

	// Step 4: Build volume relationships
	graph.buildVolumeRelationships()

	// Step 5: Detect circular dependencies
	graph.CircularDeps = graph.detectCircularDependencies()

	// Step 6: Calculate topological order (if no cycles)
	if len(graph.CircularDeps) == 0 {
		order, err := graph.topologicalSort()
		if err == nil {
			graph.TopologicalOrder = order
		}
	}

	return graph, nil
}

// buildNetworkRelationships builds network-based peer relationships
func (g *DependencyGraph) buildNetworkRelationships() {
	// Build a map of network -> services
	networkMap := make(map[string][]*ServiceNode)
	for _, node := range g.Services {
		for _, network := range node.Networks {
			networkMap[network] = append(networkMap[network], node)
		}
	}

	// For each service, find peers on the same networks
	for _, node := range g.Services {
		for _, network := range node.Networks {
			peers := networkMap[network]
			node.NetworkPeers[network] = peers
		}
	}
}

// buildVolumeRelationships builds volume-based peer relationships
func (g *DependencyGraph) buildVolumeRelationships() {
	// Build a map of volume -> services
	volumeMap := make(map[string][]*ServiceNode)
	for _, node := range g.Services {
		for _, volume := range node.Volumes {
			// Extract volume name (format can be "volume:path" or just "path")
			volumeName := extractVolumeName(volume)
			if volumeName != "" {
				volumeMap[volumeName] = append(volumeMap[volumeName], node)
			}
		}
	}

	// For each service, find peers sharing the same volumes
	for _, node := range g.Services {
		for _, volume := range node.Volumes {
			volumeName := extractVolumeName(volume)
			if volumeName != "" {
				peers := volumeMap[volumeName]
				node.VolumePeers[volumeName] = peers
			}
		}
	}
}

// extractVolumeName extracts the named volume from a volume mount string
// Handles formats like "volume-name:/path" or "/host/path:/container/path"
func extractVolumeName(volumeMount string) string {
	// Simple extraction - if it starts with /, it's a bind mount, not a named volume
	if len(volumeMount) == 0 {
		return ""
	}
	if volumeMount[0] == '/' || volumeMount[0] == '.' {
		return "" // bind mount or relative path
	}

	// Named volume - extract the part before the colon
	for i, ch := range volumeMount {
		if ch == ':' {
			return volumeMount[:i]
		}
	}

	return volumeMount
}

// detectCircularDependencies finds all circular dependency chains using DFS
func (g *DependencyGraph) detectCircularDependencies() [][]string {
	var cycles [][]string
	visited := make(map[string]bool)
	recStack := make(map[string]bool)

	for serviceName := range g.Services {
		if !visited[serviceName] {
			g.dfsDetectCycle(serviceName, visited, recStack, []string{}, &cycles)
		}
	}

	return cycles
}

// dfsDetectCycle performs DFS to detect cycles
func (g *DependencyGraph) dfsDetectCycle(
	current string,
	visited map[string]bool,
	recStack map[string]bool,
	path []string,
	cycles *[][]string,
) {
	visited[current] = true
	recStack[current] = true
	path = append(path, current)

	node := g.Services[current]
	for _, dep := range node.DependsOn {
		if !visited[dep.Name] {
			g.dfsDetectCycle(dep.Name, visited, recStack, path, cycles)
		} else if recStack[dep.Name] {
			// Found cycle - extract the cycle from path
			cycleStart := -1
			for i, s := range path {
				if s == dep.Name {
					cycleStart = i
					break
				}
			}
			if cycleStart >= 0 {
				cycle := make([]string, len(path)-cycleStart)
				copy(cycle, path[cycleStart:])
				cycle = append(cycle, dep.Name) // Complete the cycle
				*cycles = append(*cycles, cycle)
			}
		}
	}

	recStack[current] = false
}

// topologicalSort orders services by dependency using Kahn's algorithm
func (g *DependencyGraph) topologicalSort() ([]string, error) {
	// Calculate in-degree for each node
	inDegree := make(map[string]int)
	for name := range g.Services {
		inDegree[name] = 0
	}
	for _, node := range g.Services {
		for _, dep := range node.DependsOn {
			inDegree[dep.Name]++
		}
	}

	// Queue of nodes with in-degree 0
	var queue []string
	for name, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, name)
		}
	}

	// Sort queue for consistent ordering
	sort.Strings(queue)

	var result []string

	for len(queue) > 0 {
		// Pop from queue
		current := queue[0]
		queue = queue[1:]
		result = append(result, current)

		// For each dependent, decrease in-degree
		node := g.Services[current]
		var newZeroDegree []string
		for _, dependent := range node.DependedBy {
			inDegree[dependent.Name]--
			if inDegree[dependent.Name] == 0 {
				newZeroDegree = append(newZeroDegree, dependent.Name)
			}
		}

		// Sort before adding to queue for consistency
		sort.Strings(newZeroDegree)
		queue = append(queue, newZeroDegree...)
	}

	// If result doesn't contain all nodes, there's a cycle
	if len(result) != len(g.Services) {
		return nil, fmt.Errorf("circular dependencies detected")
	}

	return result, nil
}

// FilterByService returns a subgraph containing only the specified service and its dependencies
func (g *DependencyGraph) FilterByService(serviceName string, depth int) (*DependencyGraph, error) {
	node, exists := g.Services[serviceName]
	if !exists {
		return nil, fmt.Errorf("service '%s' not found", serviceName)
	}

	filtered := &DependencyGraph{
		Services: make(map[string]*ServiceNode),
	}

	visited := make(map[string]bool)
	g.collectDependencies(node, filtered, visited, depth, 0)

	// Rebuild relationships in the filtered graph
	filtered.rebuildRelationships()

	// Detect cycles in filtered graph
	filtered.CircularDeps = filtered.detectCircularDependencies()

	return filtered, nil
}

// collectDependencies recursively collects a service and its dependencies
func (g *DependencyGraph) collectDependencies(
	node *ServiceNode,
	filtered *DependencyGraph,
	visited map[string]bool,
	maxDepth int,
	currentDepth int,
) {
	if visited[node.Name] {
		return
	}
	if maxDepth >= 0 && currentDepth > maxDepth {
		return
	}

	visited[node.Name] = true

	// Create a copy of the node for the filtered graph
	nodeCopy := &ServiceNode{
		Name:           node.Name,
		Service:        node.Service,
		DependsOn:      []*ServiceNode{},
		DependedBy:     []*ServiceNode{},
		Networks:       node.Networks,
		NetworkPeers:   make(map[string][]*ServiceNode),
		Volumes:        node.Volumes,
		VolumePeers:    make(map[string][]*ServiceNode),
		HasHealthCheck: node.HasHealthCheck,
		HealthCheck:    node.HealthCheck,
	}
	filtered.Services[node.Name] = nodeCopy

	// Recursively collect dependencies
	for _, dep := range node.DependsOn {
		g.collectDependencies(dep, filtered, visited, maxDepth, currentDepth+1)
	}

	// Also collect dependents (services that depend on this one)
	for _, dependent := range node.DependedBy {
		g.collectDependencies(dependent, filtered, visited, maxDepth, currentDepth+1)
	}
}

// rebuildRelationships rebuilds dependency relationships in a filtered graph
func (g *DependencyGraph) rebuildRelationships() {
	// Clear existing relationships
	for _, node := range g.Services {
		node.DependsOn = []*ServiceNode{}
		node.DependedBy = []*ServiceNode{}
	}

	// Rebuild depends_on relationships
	for _, node := range g.Services {
		for _, depName := range node.Service.DependsOn {
			if depNode, exists := g.Services[depName]; exists {
				node.DependsOn = append(node.DependsOn, depNode)
				depNode.DependedBy = append(depNode.DependedBy, node)
			}
		}
	}

	// Rebuild network and volume relationships
	g.buildNetworkRelationships()
	g.buildVolumeRelationships()
}

// GetAllRelationships returns all relationships in the graph
func (g *DependencyGraph) GetAllRelationships() []Relationship {
	var relationships []Relationship

	// Add dependency relationships
	for _, node := range g.Services {
		for _, dep := range node.DependsOn {
			relationships = append(relationships, Relationship{
				From:     node.Name,
				To:       dep.Name,
				Type:     RelationshipDependsOn,
				Metadata: "",
			})
		}
	}

	// Add network relationships
	for _, node := range g.Services {
		for network, peers := range node.NetworkPeers {
			for _, peer := range peers {
				if peer.Name != node.Name {
					relationships = append(relationships, Relationship{
						From:     node.Name,
						To:       peer.Name,
						Type:     RelationshipNetwork,
						Metadata: network,
					})
				}
			}
		}
	}

	// Add volume relationships
	for _, node := range g.Services {
		for volume, peers := range node.VolumePeers {
			for _, peer := range peers {
				if peer.Name != node.Name {
					relationships = append(relationships, Relationship{
						From:     node.Name,
						To:       peer.Name,
						Type:     RelationshipVolume,
						Metadata: volume,
					})
				}
			}
		}
	}

	return relationships
}

// HasCircularDependencies checks if the graph has any circular dependencies
func (g *DependencyGraph) HasCircularDependencies() bool {
	return len(g.CircularDeps) > 0
}

// GetRootServices returns services with no dependencies
func (g *DependencyGraph) GetRootServices() []string {
	var roots []string
	for name, node := range g.Services {
		if len(node.DependsOn) == 0 {
			roots = append(roots, name)
		}
	}
	sort.Strings(roots)
	return roots
}
