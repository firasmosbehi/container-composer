package tui

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/firasmosbahi/container-composer/core"
)

// State constants
const (
	stateGraphInit = iota
	stateGraphMain
	stateGraphServiceDetails
	stateGraphFilter
	stateGraphExport
)

type highlightType int

const (
	highlightNone highlightType = iota
	highlightSelected
	highlightDependency
	highlightDependent
	highlightNetwork
	highlightVolume
)

// dependencyGraphModel is the main TUI model
type dependencyGraphModel struct {
	state int

	// Data
	composeFile   *core.ComposeFile
	composePath   string
	graph         *core.DependencyGraph
	filteredGraph *core.DependencyGraph

	// UI state
	serviceList     list.Model
	selectedService string
	viewport        viewport.Model
	filterInput     textinput.Model

	// Display options
	showNetworks     bool
	showVolumes      bool
	showHealthChecks bool
	highlightMode    bool

	// Highlighted services
	highlightedServices map[string]highlightType

	// Export
	exportPath   string
	exportFormat string

	// Status
	err      error
	message  string
	quitting bool
}

func newDependencyGraphModel() dependencyGraphModel {
	return dependencyGraphModel{
		state:               stateGraphInit,
		composePath:         "docker-compose.yml",
		showNetworks:        true,
		showVolumes:         true,
		showHealthChecks:    true,
		highlightedServices: make(map[string]highlightType),
	}
}

func (m dependencyGraphModel) Init() tea.Cmd {
	return func() tea.Msg {
		return m.loadGraph()
	}
}

func (m dependencyGraphModel) loadGraph() tea.Msg {
	// Check if docker-compose.yml exists
	if _, err := os.Stat(m.composePath); os.IsNotExist(err) {
		return graphErrorMsg{fmt.Errorf("docker-compose.yml not found")}
	}

	// Parse compose file
	composeFile, err := core.ParseComposeFile(m.composePath)
	if err != nil {
		return graphErrorMsg{fmt.Errorf("failed to parse: %w", err)}
	}

	// Build graph
	graph, err := composeFile.BuildDependencyGraph()
	if err != nil {
		return graphErrorMsg{fmt.Errorf("failed to build graph: %w", err)}
	}

	return graphLoadedMsg{composeFile, graph}
}

type graphErrorMsg struct{ err error }
type graphLoadedMsg struct {
	composeFile *core.ComposeFile
	graph       *core.DependencyGraph
}

func (m dependencyGraphModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case graphErrorMsg:
		m.err = msg.err
		m.state = stateGraphMain
		return m, nil

	case graphLoadedMsg:
		m.composeFile = msg.composeFile
		m.graph = msg.graph
		m.filteredGraph = msg.graph
		m.state = stateGraphMain

		// Build service list
		m.serviceList = m.createServiceList()

		// Create viewport for displaying graph
		m.viewport = viewport.New(80, 20)
		m.viewport.SetContent(m.renderGraph())

		return m, nil

	case tea.KeyMsg:
		return m.handleKeyPress(msg)

	case tea.WindowSizeMsg:
		m.viewport.Width = msg.Width - 4
		m.viewport.Height = msg.Height - 10
		m.viewport.SetContent(m.renderGraph())
		return m, nil
	}

	return m.updateComponents(msg)
}

func (m dependencyGraphModel) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.state {
	case stateGraphMain:
		switch msg.String() {
		case "ctrl+c", "q":
			if m.highlightMode {
				// First q clears highlight
				m.highlightMode = false
				m.selectedService = ""
				m.highlightedServices = make(map[string]highlightType)
				m.viewport.SetContent(m.renderGraph())
				m.message = "Highlight cleared"
				return m, nil
			}
			m.quitting = true
			return m, tea.Quit

		case "esc":
			if m.highlightMode {
				m.highlightMode = false
				m.selectedService = ""
				m.highlightedServices = make(map[string]highlightType)
				m.viewport.SetContent(m.renderGraph())
				m.message = "Highlight cleared"
				return m, nil
			}
			return newMainMenuModel(), nil

		case "enter":
			// Show service details
			if i, ok := m.serviceList.SelectedItem().(menuItem); ok {
				m.selectedService = i.id
				m.state = stateGraphServiceDetails
				m.viewport.SetContent(m.renderServiceDetails())
				return m, nil
			}

		case "h":
			// Toggle highlight mode
			if i, ok := m.serviceList.SelectedItem().(menuItem); ok {
				m.selectedService = i.id
				m.highlightMode = !m.highlightMode
				if m.highlightMode {
					m.calculateHighlights()
					m.message = fmt.Sprintf("Highlighting: %s", m.selectedService)
				} else {
					m.highlightedServices = make(map[string]highlightType)
					m.message = "Highlight cleared"
				}
				m.viewport.SetContent(m.renderGraph())
				return m, nil
			}

		case "f":
			// Enter filter mode
			m.state = stateGraphFilter
			m.filterInput = textinput.New()
			m.filterInput.Placeholder = "service name"
			m.filterInput.Focus()
			return m, nil

		case "e":
			// Export
			m.state = stateGraphExport
			return m, nil

		case "n":
			// Toggle networks
			m.showNetworks = !m.showNetworks
			m.viewport.SetContent(m.renderGraph())
			if m.showNetworks {
				m.message = "Networks: ON"
			} else {
				m.message = "Networks: OFF"
			}
			return m, nil

		case "v":
			// Toggle volumes
			m.showVolumes = !m.showVolumes
			m.viewport.SetContent(m.renderGraph())
			if m.showVolumes {
				m.message = "Volumes: ON"
			} else {
				m.message = "Volumes: OFF"
			}
			return m, nil

		case "c":
			// Toggle health checks
			m.showHealthChecks = !m.showHealthChecks
			m.viewport.SetContent(m.renderGraph())
			if m.showHealthChecks {
				m.message = "Health checks: ON"
			} else {
				m.message = "Health checks: OFF"
			}
			return m, nil
		}

	case stateGraphServiceDetails:
		switch msg.String() {
		case "ctrl+c", "q", "esc", "enter":
			m.state = stateGraphMain
			m.viewport.SetContent(m.renderGraph())
			return m, nil
		}

	case stateGraphFilter:
		switch msg.String() {
		case "esc":
			// Cancel filter
			m.state = stateGraphMain
			m.filterInput.Blur()
			m.filteredGraph = m.graph
			m.viewport.SetContent(m.renderGraph())
			m.message = "Filter cancelled"
			return m, nil

		case "enter":
			// Apply filter
			serviceName := strings.TrimSpace(m.filterInput.Value())
			if serviceName != "" {
				filtered, err := m.graph.FilterByService(serviceName, -1)
				if err != nil {
					m.message = fmt.Sprintf("Error: %v", err)
				} else {
					m.filteredGraph = filtered
					m.viewport.SetContent(m.renderGraph())
					m.message = fmt.Sprintf("Filtered to service: %s", serviceName)
				}
			} else {
				// Empty = reset filter
				m.filteredGraph = m.graph
				m.viewport.SetContent(m.renderGraph())
				m.message = "Filter cleared"
			}
			m.state = stateGraphMain
			m.filterInput.Blur()
			return m, nil
		}

	case stateGraphExport:
		switch msg.String() {
		case "1":
			// Export as ASCII
			if err := m.exportGraph("ascii", "graph.txt"); err != nil {
				m.message = fmt.Sprintf("Error: %v", err)
			} else {
				m.message = "Exported to graph.txt"
			}
			m.state = stateGraphMain
			return m, nil

		case "2":
			// Export as DOT
			if err := m.exportGraph("dot", "graph.dot"); err != nil {
				m.message = fmt.Sprintf("Error: %v", err)
			} else {
				m.message = "Exported to graph.dot"
			}
			m.state = stateGraphMain
			return m, nil

		case "esc":
			m.state = stateGraphMain
			m.message = "Export cancelled"
			return m, nil
		}
	}

	return m, nil
}

func (m dependencyGraphModel) updateComponents(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch m.state {
	case stateGraphMain:
		m.serviceList, cmd = m.serviceList.Update(msg)
		m.viewport, _ = m.viewport.Update(msg)

	case stateGraphFilter:
		m.filterInput, cmd = m.filterInput.Update(msg)

	case stateGraphServiceDetails:
		m.viewport, cmd = m.viewport.Update(msg)
	}

	return m, cmd
}

func (m dependencyGraphModel) createServiceList() list.Model {
	items := []list.Item{}

	// Sort services for consistent display
	serviceNames := []string{}
	for name := range m.graph.Services {
		serviceNames = append(serviceNames, name)
	}
	sort.Strings(serviceNames)

	for _, name := range serviceNames {
		node := m.graph.Services[name]
		desc := fmt.Sprintf("Deps: %d", len(node.DependsOn))
		if node.HasHealthCheck {
			desc += " | ⚡HC"
		}

		items = append(items, menuItem{
			title: name,
			desc:  desc,
			id:    name,
		})
	}

	delegate := list.NewDefaultDelegate()
	l := list.New(items, delegate, 30, 20)
	l.Title = "Services"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)
	l.Styles.Title = titleStyle

	return l
}

func (m dependencyGraphModel) calculateHighlights() {
	m.highlightedServices = make(map[string]highlightType)

	node := m.graph.Services[m.selectedService]
	if node == nil {
		return
	}

	// Mark selected service
	m.highlightedServices[m.selectedService] = highlightSelected

	// Mark dependencies
	for _, dep := range node.DependsOn {
		m.highlightedServices[dep.Name] = highlightDependency
	}

	// Mark dependents
	for _, dependent := range node.DependedBy {
		m.highlightedServices[dependent.Name] = highlightDependent
	}

	// Mark network peers
	for _, peers := range node.NetworkPeers {
		for _, peer := range peers {
			if peer.Name != m.selectedService {
				if _, exists := m.highlightedServices[peer.Name]; !exists {
					m.highlightedServices[peer.Name] = highlightNetwork
				}
			}
		}
	}

	// Mark volume peers
	for _, peers := range node.VolumePeers {
		for _, peer := range peers {
			if peer.Name != m.selectedService {
				if _, exists := m.highlightedServices[peer.Name]; !exists {
					m.highlightedServices[peer.Name] = highlightVolume
				}
			}
		}
	}
}

func (m dependencyGraphModel) renderGraph() string {
	options := core.ASCIIOptions{
		ShowNetworks:     m.showNetworks,
		ShowVolumes:      m.showVolumes,
		ShowHealthChecks: m.showHealthChecks,
	}

	output := m.filteredGraph.FormatASCII(options)

	// Apply highlighting if in highlight mode
	if m.highlightMode {
		output = m.applyHighlighting(output)
	}

	return output
}

func (m dependencyGraphModel) applyHighlighting(content string) string {
	// Color service names based on highlight type
	lines := strings.Split(content, "\n")
	var highlighted []string

	for _, line := range lines {
		newLine := line
		for serviceName, hlType := range m.highlightedServices {
			if strings.Contains(line, serviceName) {
				style := lipgloss.NewStyle()
				switch hlType {
				case highlightSelected:
					style = style.Foreground(lipgloss.Color("#7D56F4")).Bold(true)
				case highlightDependency:
					style = style.Foreground(lipgloss.Color("#00D7D7"))
				case highlightDependent:
					style = style.Foreground(lipgloss.Color("#04B575"))
				case highlightNetwork:
					style = style.Foreground(lipgloss.Color("214"))
				case highlightVolume:
					style = style.Foreground(lipgloss.Color("208"))
				}
				newLine = strings.Replace(newLine, serviceName,
					style.Render(serviceName), 1)
			}
		}
		highlighted = append(highlighted, newLine)
	}

	return strings.Join(highlighted, "\n")
}

func (m dependencyGraphModel) renderServiceDetails() string {
	node := m.graph.Services[m.selectedService]
	if node == nil {
		return "Service not found\n"
	}

	var builder strings.Builder

	builder.WriteString(titleStyle.Render(fmt.Sprintf("Service: %s", m.selectedService)))
	builder.WriteString("\n\n")

	// Image/Build
	if node.Service.Image != "" {
		builder.WriteString(fmt.Sprintf("Image: %s\n", node.Service.Image))
	} else if node.Service.Build != nil {
		builder.WriteString(fmt.Sprintf("Build Context: %s\n", node.Service.Build.Context))
		if node.Service.Build.Dockerfile != "" {
			builder.WriteString(fmt.Sprintf("Dockerfile: %s\n", node.Service.Build.Dockerfile))
		}
	}

	builder.WriteString("\n")

	// Dependencies
	if len(node.DependsOn) > 0 {
		builder.WriteString("Dependencies:\n")
		for _, dep := range node.DependsOn {
			builder.WriteString(fmt.Sprintf("  • %s\n", dep.Name))
		}
		builder.WriteString("\n")
	}

	// Dependents
	if len(node.DependedBy) > 0 {
		builder.WriteString("Required By:\n")
		for _, dep := range node.DependedBy {
			builder.WriteString(fmt.Sprintf("  • %s\n", dep.Name))
		}
		builder.WriteString("\n")
	}

	// Networks
	if len(node.Networks) > 0 {
		builder.WriteString("Networks:\n")
		for _, network := range node.Networks {
			builder.WriteString(fmt.Sprintf("  🌐 %s", network))
			if peers, ok := node.NetworkPeers[network]; ok && len(peers) > 0 {
				peerNames := []string{}
				for _, peer := range peers {
					if peer.Name != m.selectedService {
						peerNames = append(peerNames, peer.Name)
					}
				}
				if len(peerNames) > 0 {
					builder.WriteString(fmt.Sprintf(" (shared with: %s)", strings.Join(peerNames, ", ")))
				}
			}
			builder.WriteString("\n")
		}
		builder.WriteString("\n")
	}

	// Volumes
	if len(node.Volumes) > 0 {
		builder.WriteString("Volumes:\n")
		for _, volume := range node.Volumes {
			builder.WriteString(fmt.Sprintf("  💾 %s\n", volume))
		}
		builder.WriteString("\n")
	}

	// Ports
	if len(node.Service.Ports) > 0 {
		builder.WriteString("Ports:\n")
		for _, port := range node.Service.Ports {
			builder.WriteString(fmt.Sprintf("  • %s\n", port))
		}
		builder.WriteString("\n")
	}

	// Health check
	if node.HasHealthCheck {
		builder.WriteString("⚡ Health Check:\n")
		if len(node.HealthCheck.Test) > 0 {
			builder.WriteString(fmt.Sprintf("  Test: %v\n", node.HealthCheck.Test))
		}
		if node.HealthCheck.Interval != "" {
			builder.WriteString(fmt.Sprintf("  Interval: %s\n", node.HealthCheck.Interval))
		}
		if node.HealthCheck.Timeout != "" {
			builder.WriteString(fmt.Sprintf("  Timeout: %s\n", node.HealthCheck.Timeout))
		}
		if node.HealthCheck.Retries > 0 {
			builder.WriteString(fmt.Sprintf("  Retries: %d\n", node.HealthCheck.Retries))
		}
		builder.WriteString("\n")
	}

	builder.WriteString(helpStyle.Render("Press 'enter' or 'esc' to go back"))

	return builder.String()
}

func (m dependencyGraphModel) exportGraph(format, filename string) error {
	var content string

	switch format {
	case "ascii":
		options := core.ASCIIOptions{
			ShowNetworks:     m.showNetworks,
			ShowVolumes:      m.showVolumes,
			ShowHealthChecks: m.showHealthChecks,
		}
		content = m.filteredGraph.FormatASCII(options)

	case "dot":
		options := core.DOTOptions{
			ShowNetworks:     m.showNetworks,
			ShowVolumes:      m.showVolumes,
			ShowHealthChecks: m.showHealthChecks,
			HighlightCycles:  true,
		}
		content = m.filteredGraph.FormatDOT(options)
	}

	return os.WriteFile(filename, []byte(content), 0644)
}

func (m dependencyGraphModel) View() string {
	if m.quitting {
		return "Goodbye! 👋\n"
	}

	switch m.state {
	case stateGraphInit:
		return "Loading dependency graph...\n"

	case stateGraphMain:
		return m.renderMainView()

	case stateGraphServiceDetails:
		return docStyle.Render(m.viewport.View())

	case stateGraphFilter:
		return m.renderFilterView()

	case stateGraphExport:
		return m.renderExportView()
	}

	return ""
}

func (m dependencyGraphModel) renderMainView() string {
	if m.err != nil {
		s := titleStyle.Render("Error") + "\n\n"
		s += fmt.Sprintf("❌ %v\n\n", m.err)
		s += helpStyle.Render("Press 'esc' to go back")
		return docStyle.Render(s)
	}

	// Create two-column layout
	leftColumn := m.serviceList.View()
	rightColumn := m.viewport.View()

	layout := lipgloss.JoinHorizontal(
		lipgloss.Top,
		lipgloss.NewStyle().Width(35).Render(leftColumn),
		lipgloss.NewStyle().Width(80).Render(rightColumn),
	)

	// Status bar
	statusBar := m.renderStatusBar()

	// Help text
	help := m.renderHelpText()

	return docStyle.Render(
		titleStyle.Render("Dependency Graph") + "\n" +
			layout + "\n\n" +
			statusBar + "\n" +
			help,
	)
}

func (m dependencyGraphModel) renderStatusBar() string {
	parts := []string{}

	if m.showNetworks {
		parts = append(parts, "🌐 Networks")
	}
	if m.showVolumes {
		parts = append(parts, "💾 Volumes")
	}
	if m.showHealthChecks {
		parts = append(parts, "⚡ Health")
	}
	if m.highlightMode {
		parts = append(parts, fmt.Sprintf("🎯 Highlighting: %s", m.selectedService))
	}

	status := strings.Join(parts, " | ")
	if m.message != "" {
		status += " | " + m.message
	}

	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Render(status)
}

func (m dependencyGraphModel) renderHelpText() string {
	if m.highlightMode {
		legend := lipgloss.NewStyle().Foreground(lipgloss.Color("#7D56F4")).Render("█ Selected") + " " +
			lipgloss.NewStyle().Foreground(lipgloss.Color("#00D7D7")).Render("█ Dependency") + " " +
			lipgloss.NewStyle().Foreground(lipgloss.Color("#04B575")).Render("█ Dependent") + " " +
			lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Render("█ Network") + " " +
			lipgloss.NewStyle().Foreground(lipgloss.Color("208")).Render("█ Volume")

		return helpStyle.Render(
			"↑↓ navigate • 'h' clear highlight • 'enter' details • 'f' filter • 'e' export\n" +
				"'n' toggle networks • 'v' toggle volumes • 'c' toggle health • 'esc' back • 'q' quit\n" +
				legend,
		)
	}

	return helpStyle.Render(
		"↑↓ navigate • 'h' highlight • 'enter' details • 'f' filter • 'e' export\n" +
			"'n' toggle networks • 'v' toggle volumes • 'c' toggle health • 'esc' back • 'q' quit",
	)
}

func (m dependencyGraphModel) renderFilterView() string {
	s := titleStyle.Render("Filter by Service") + "\n\n"
	s += "Enter service name to show only that service and its dependencies:\n\n"
	s += m.filterInput.View() + "\n\n"
	s += helpStyle.Render("Leave empty to show all • 'enter' to apply • 'esc' to cancel")
	return docStyle.Render(s)
}

func (m dependencyGraphModel) renderExportView() string {
	s := titleStyle.Render("Export Graph") + "\n\n"
	s += "Choose export format:\n\n"
	s += "  [1] ASCII text (graph.txt)\n"
	s += "  [2] Graphviz DOT (graph.dot)\n\n"
	s += helpStyle.Render("Press number to export • 'esc' to cancel")
	return docStyle.Render(s)
}
