package tui

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/firasmosbahi/container-composer/core"
	"gopkg.in/yaml.v3"
)

// State constants for add service wizard
const (
	stateAddServiceInit = iota
	stateAddServiceName
	stateAddServiceImageOrBuild
	stateAddServiceImageName
	stateAddServiceBuildContext
	stateAddServiceBuildDockerfile
	stateAddServicePortsConfirm
	stateAddServicePorts
	stateAddServiceEnvConfirm
	stateAddServiceEnv
	stateAddServiceVolumesConfirm
	stateAddServiceVolumes
	stateAddServiceNetworksConfirm
	stateAddServiceNetworks
	stateAddServiceDepsConfirm
	stateAddServiceDeps
	stateAddServiceRestart
	stateAddServiceAdvancedConfirm
	stateAddServiceAdvancedCommand
	stateAddServiceAdvancedWorkdir
	stateAddServiceAdvancedUser
	stateAddServiceAdvancedHostname
	stateAddServicePreview
	stateAddServiceConfirm
	stateAddServiceSuccess
)

type addServiceModel struct {
	state int

	// Data
	composeFile *core.ComposeFile
	composePath string
	service     core.Service

	// Form components
	textInput    textInputForm
	confirmInput confirmForm
	multiInput   multiInputForm
	kvInput      keyValueInputForm
	selectList   list.Model
	previewPort  viewport.Model

	// State flags
	useImage         bool
	hasAdvanced      bool
	selectedServices []string // For dependencies
	selectedNetworks []string

	// Preview
	yamlPreview string

	// Status
	err      error
	quitting bool
}

func newAddServiceModel() addServiceModel {
	return addServiceModel{
		state:       stateAddServiceInit,
		composePath: "docker-compose.yml",
	}
}

func (m addServiceModel) Init() tea.Cmd {
	return func() tea.Msg {
		return m.loadComposeFile()
	}
}

func (m addServiceModel) loadComposeFile() tea.Msg {
	// Get current directory for debugging
	cwd, _ := os.Getwd()

	// Check if docker-compose.yml exists
	if _, err := os.Stat(m.composePath); os.IsNotExist(err) {
		return addServiceErrorMsg{fmt.Errorf("docker-compose.yml not found in %s", cwd)}
	}

	// Parse existing docker-compose.yml
	composeFile, err := core.ParseComposeFile(m.composePath)
	if err != nil {
		return addServiceErrorMsg{fmt.Errorf("failed to parse docker-compose.yml in %s: %w", cwd, err)}
	}

	return addServiceComposeFileLoaded{composeFile}
}

type addServiceErrorMsg struct{ err error }
type addServiceComposeFileLoaded struct{ file *core.ComposeFile }

func (m addServiceModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case addServiceErrorMsg:
		m.err = msg.err
		m.state = stateAddServiceSuccess
		return m, nil

	case addServiceComposeFileLoaded:
		m.composeFile = msg.file
		m.state = stateAddServiceName
		// Initialize service name input
		m.textInput = newTextInputForm(
			"Service Name",
			"my-service",
			"",
			"Unique identifier for this service",
			func(s string) error {
				if s == "" {
					return fmt.Errorf("service name cannot be empty")
				}
				return nil
			},
		)
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			m.quitting = true
			return m, tea.Quit

		case "q":
			if m.state != stateAddServiceName &&
				m.state != stateAddServiceImageName &&
				m.state != stateAddServiceBuildContext &&
				m.state != stateAddServiceBuildDockerfile &&
				m.state != stateAddServicePorts &&
				m.state != stateAddServiceEnv &&
				m.state != stateAddServiceVolumes &&
				m.state != stateAddServiceNetworks &&
				m.state != stateAddServiceAdvancedCommand &&
				m.state != stateAddServiceAdvancedWorkdir &&
				m.state != stateAddServiceAdvancedUser &&
				m.state != stateAddServiceAdvancedHostname {
				m.quitting = true
				return m, tea.Quit
			}

		case "esc":
			return m.handleBack()

		case "enter":
			// Let components handle Enter for multi-input states (ports, env, volumes, networks)
			// All other states should call handleEnter to advance
			if m.state != stateAddServicePorts &&
				m.state != stateAddServiceEnv &&
				m.state != stateAddServiceVolumes &&
				m.state != stateAddServiceNetworks {
				return m.handleEnter()
			}
		}

	case tea.WindowSizeMsg:
		if m.state == stateAddServicePreview {
			m.previewPort.Width = msg.Width - 4
			m.previewPort.Height = msg.Height - 8
		}
	}

	// Update active component
	return m.updateComponent(msg)
}

func (m addServiceModel) updateComponent(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch m.state {
	case stateAddServiceName, stateAddServiceImageName,
		stateAddServiceBuildContext, stateAddServiceBuildDockerfile,
		stateAddServiceAdvancedCommand, stateAddServiceAdvancedWorkdir,
		stateAddServiceAdvancedUser, stateAddServiceAdvancedHostname:
		m.textInput, cmd = m.textInput.Update(msg)

	case stateAddServiceImageOrBuild, stateAddServicePortsConfirm,
		stateAddServiceEnvConfirm, stateAddServiceVolumesConfirm,
		stateAddServiceNetworksConfirm, stateAddServiceDepsConfirm,
		stateAddServiceAdvancedConfirm, stateAddServiceConfirm:
		m.confirmInput, cmd = m.confirmInput.Update(msg)

	case stateAddServicePorts:
		m.multiInput, cmd = m.multiInput.Update(msg)
		// Check if user is done (pressed Enter with empty field)
		if keyMsg, ok := msg.(tea.KeyMsg); ok && keyMsg.String() == "enter" && m.multiInput.IsDone() {
			m.service.Ports = m.multiInput.Values()
			m.state = stateAddServiceEnvConfirm
			m.confirmInput = newConfirmForm("Add environment variables?", false)
			return m, nil
		}

	case stateAddServiceVolumes:
		m.multiInput, cmd = m.multiInput.Update(msg)
		// Check if user is done (pressed Enter with empty field)
		if keyMsg, ok := msg.(tea.KeyMsg); ok && keyMsg.String() == "enter" && m.multiInput.IsDone() {
			m.service.Volumes = m.multiInput.Values()
			m.state = stateAddServiceNetworksConfirm
			m.confirmInput = newConfirmForm("Connect to networks?", false)
			return m, nil
		}

	case stateAddServiceNetworks:
		m.multiInput, cmd = m.multiInput.Update(msg)
		// Check if user is done (pressed Enter with empty field)
		if keyMsg, ok := msg.(tea.KeyMsg); ok && keyMsg.String() == "enter" && m.multiInput.IsDone() {
			m.service.Networks = m.multiInput.Values()
			m.state = stateAddServiceDepsConfirm
			m.confirmInput = newConfirmForm("Add service dependencies (depends_on)?", false)
			return m, nil
		}

	case stateAddServiceEnv:
		m.kvInput, cmd = m.kvInput.Update(msg)
		// Check if user is done (pressed Enter with empty fields)
		if keyMsg, ok := msg.(tea.KeyMsg); ok && keyMsg.String() == "enter" && m.kvInput.IsDone() {
			// Save values and move to next state
			m.service.Environment = m.kvInput.Values()
			m.state = stateAddServiceVolumesConfirm
			m.confirmInput = newConfirmForm("Mount volumes?", false)
			return m, nil
		}

	case stateAddServiceDeps, stateAddServiceRestart:
		m.selectList, cmd = m.selectList.Update(msg)

	case stateAddServicePreview:
		m.previewPort, cmd = m.previewPort.Update(msg)
	}

	return m, cmd
}

func (m addServiceModel) handleBack() (tea.Model, tea.Cmd) {
	switch m.state {
	case stateAddServiceName, stateAddServiceSuccess:
		return newAddMenuModel(), nil
	case stateAddServiceImageOrBuild:
		m.state = stateAddServiceName
	case stateAddServiceImageName:
		m.state = stateAddServiceImageOrBuild
	case stateAddServiceBuildContext:
		m.state = stateAddServiceImageOrBuild
	case stateAddServiceBuildDockerfile:
		m.state = stateAddServiceBuildContext
	case stateAddServicePortsConfirm:
		if m.useImage {
			m.state = stateAddServiceImageName
		} else {
			m.state = stateAddServiceBuildDockerfile
		}
	case stateAddServicePorts:
		m.state = stateAddServicePortsConfirm
	case stateAddServiceEnvConfirm:
		m.state = stateAddServicePortsConfirm
		if !m.confirmInput.Value() {
			m.state = stateAddServicePorts
		}
	case stateAddServiceEnv:
		m.state = stateAddServiceEnvConfirm
	case stateAddServiceVolumesConfirm:
		m.state = stateAddServiceEnvConfirm
	case stateAddServiceVolumes:
		m.state = stateAddServiceVolumesConfirm
	case stateAddServiceNetworksConfirm:
		m.state = stateAddServiceVolumesConfirm
	case stateAddServiceNetworks:
		m.state = stateAddServiceNetworksConfirm
	case stateAddServiceDepsConfirm:
		m.state = stateAddServiceNetworksConfirm
	case stateAddServiceDeps:
		m.state = stateAddServiceDepsConfirm
	case stateAddServiceRestart:
		m.state = stateAddServiceDepsConfirm
	case stateAddServiceAdvancedConfirm:
		m.state = stateAddServiceRestart
	case stateAddServicePreview:
		m.state = stateAddServiceAdvancedConfirm
	case stateAddServiceConfirm:
		m.state = stateAddServicePreview
	default:
		return newAddMenuModel(), nil
	}
	return m, nil
}

func (m addServiceModel) handleEnter() (tea.Model, tea.Cmd) {
	switch m.state {
	case stateAddServiceName:
		serviceName := m.textInput.Value()
		if serviceName == "" {
			return m, nil
		}

		// Check for conflicts
		if m.composeFile.ServiceExists(serviceName) {
			m.err = fmt.Errorf("service '%s' already exists", serviceName)
			return m, nil
		}

		m.service.Name = serviceName
		m.state = stateAddServiceImageOrBuild
		m.confirmInput = newConfirmForm("Use a pre-built image? (No = build from Dockerfile)", true)

	case stateAddServiceImageOrBuild:
		m.useImage = m.confirmInput.Value()
		if m.useImage {
			m.state = stateAddServiceImageName
			m.textInput = newTextInputForm(
				"Docker Image",
				"nginx:latest",
				"",
				"e.g., nginx:latest, postgres:15, node:20-alpine",
				func(s string) error {
					if s == "" {
						return fmt.Errorf("image cannot be empty")
					}
					return nil
				},
			)
		} else {
			m.state = stateAddServiceBuildContext
			m.textInput = newTextInputForm(
				"Build Context Path",
				".",
				".",
				"Path to directory containing Dockerfile",
				nil,
			)
		}

	case stateAddServiceImageName:
		m.service.Image = m.textInput.Value()
		m.state = stateAddServicePortsConfirm
		m.confirmInput = newConfirmForm("Expose ports?", true)

	case stateAddServiceBuildContext:
		if m.service.Build == nil {
			m.service.Build = &core.BuildConfig{}
		}
		m.service.Build.Context = m.textInput.Value()
		m.state = stateAddServiceBuildDockerfile
		m.textInput = newTextInputForm(
			"Dockerfile Name",
			"Dockerfile",
			"Dockerfile",
			"Name of the Dockerfile",
			nil,
		)

	case stateAddServiceBuildDockerfile:
		m.service.Build.Dockerfile = m.textInput.Value()
		m.state = stateAddServicePortsConfirm
		m.confirmInput = newConfirmForm("Expose ports?", true)

	case stateAddServicePortsConfirm:
		if m.confirmInput.Value() {
			m.state = stateAddServicePorts
			m.multiInput = newMultiInputForm(
				"Port Mappings",
				"8080:80 or 3000",
				"Format: host:container or just container port",
				nil,
			)
		} else {
			m.state = stateAddServiceEnvConfirm
			m.confirmInput = newConfirmForm("Add environment variables?", false)
		}

	case stateAddServicePorts:
		if len(m.multiInput.Values()) > 0 || !m.multiInput.HasValues() {
			m.service.Ports = m.multiInput.Values()
			m.state = stateAddServiceEnvConfirm
			m.confirmInput = newConfirmForm("Add environment variables?", false)
		}

	case stateAddServiceEnvConfirm:
		if m.confirmInput.Value() {
			m.state = stateAddServiceEnv
			m.kvInput = newKeyValueInputForm(
				"Environment Variables",
				"Enter key-value pairs. Tab to switch fields.",
			)
		} else {
			m.state = stateAddServiceVolumesConfirm
			m.confirmInput = newConfirmForm("Mount volumes?", false)
		}

	case stateAddServiceEnv:
		if len(m.kvInput.Values()) > 0 || (m.kvInput.Values() == nil) {
			m.service.Environment = m.kvInput.Values()
			m.state = stateAddServiceVolumesConfirm
			m.confirmInput = newConfirmForm("Mount volumes?", false)
		}

	case stateAddServiceVolumesConfirm:
		if m.confirmInput.Value() {
			m.state = stateAddServiceVolumes
			m.multiInput = newMultiInputForm(
				"Volume Mounts",
				"./app:/app or data:/var/lib/data",
				"Format: host:container or volume:container",
				nil,
			)
		} else {
			m.state = stateAddServiceNetworksConfirm
			m.confirmInput = newConfirmForm("Connect to networks?", false)
		}

	case stateAddServiceVolumes:
		if len(m.multiInput.Values()) > 0 || !m.multiInput.HasValues() {
			m.service.Volumes = m.multiInput.Values()
			m.state = stateAddServiceNetworksConfirm
			m.confirmInput = newConfirmForm("Connect to networks?", false)
		}

	case stateAddServiceNetworksConfirm:
		if m.confirmInput.Value() {
			m.state = stateAddServiceNetworks
			m.multiInput = newMultiInputForm(
				"Network Names",
				"network-name",
				"Enter network names to connect this service to",
				nil,
			)
		} else {
			m.state = stateAddServiceDepsConfirm
			m.confirmInput = newConfirmForm("Add service dependencies (depends_on)?", false)
		}

	case stateAddServiceNetworks:
		if len(m.multiInput.Values()) > 0 || !m.multiInput.HasValues() {
			m.service.Networks = m.multiInput.Values()
			m.state = stateAddServiceDepsConfirm
			m.confirmInput = newConfirmForm("Add service dependencies (depends_on)?", false)
		}

	case stateAddServiceDepsConfirm:
		if m.confirmInput.Value() && len(m.composeFile.Services) > 0 {
			m.state = stateAddServiceDeps
			// Create list of existing services
			var services []string
			for name := range m.composeFile.Services {
				if name != m.service.Name {
					services = append(services, name)
				}
			}
			if len(services) == 0 {
				// No other services, skip
				m.state = stateAddServiceRestart
				return m.createRestartPolicySelect()
			}
			items := make([]list.Item, len(services))
			for i, svc := range services {
				items[i] = menuItem{title: svc, desc: "Service dependency", id: svc}
			}
			m.selectList = list.New(items, list.NewDefaultDelegate(), 0, 0)
			m.selectList.Title = "Select Dependencies (space to select, enter to continue)"
			m.selectList.SetShowStatusBar(false)
			m.selectList.Styles.Title = titleStyle
		} else {
			m.state = stateAddServiceRestart
			return m.createRestartPolicySelect()
		}

	case stateAddServiceDeps:
		// Note: For simplicity, we'll skip multi-select for now and move to restart policy
		m.state = stateAddServiceRestart
		return m.createRestartPolicySelect()

	case stateAddServiceRestart:
		if i, ok := m.selectList.SelectedItem().(menuItem); ok {
			m.service.Restart = i.id
		}
		m.state = stateAddServiceAdvancedConfirm
		m.confirmInput = newConfirmForm("Configure advanced options? (command, working_dir, etc.)", false)

	case stateAddServiceAdvancedConfirm:
		if m.confirmInput.Value() {
			m.hasAdvanced = true
			m.state = stateAddServiceAdvancedCommand
			m.textInput = newTextInputForm(
				"Override Default Command",
				"",
				"",
				"Command to run when container starts (leave empty to skip)",
				nil,
			)
		} else {
			return m.generatePreview()
		}

	case stateAddServiceAdvancedCommand:
		if m.textInput.Value() != "" {
			m.service.Command = m.textInput.Value()
		}
		m.state = stateAddServiceAdvancedWorkdir
		m.textInput = newTextInputForm(
			"Working Directory",
			"/app",
			"",
			"Working directory in the container (leave empty to skip)",
			nil,
		)

	case stateAddServiceAdvancedWorkdir:
		if m.textInput.Value() != "" {
			m.service.WorkingDir = m.textInput.Value()
		}
		m.state = stateAddServiceAdvancedUser
		m.textInput = newTextInputForm(
			"User",
			"",
			"",
			"User to run as (uid:gid or username, leave empty to skip)",
			nil,
		)

	case stateAddServiceAdvancedUser:
		if m.textInput.Value() != "" {
			m.service.User = m.textInput.Value()
		}
		m.state = stateAddServiceAdvancedHostname
		m.textInput = newTextInputForm(
			"Custom Hostname",
			"",
			"",
			"Container hostname (leave empty to skip)",
			nil,
		)

	case stateAddServiceAdvancedHostname:
		if m.textInput.Value() != "" {
			m.service.Hostname = m.textInput.Value()
		}
		return m.generatePreview()

	case stateAddServicePreview:
		m.state = stateAddServiceConfirm
		m.confirmInput = newConfirmForm("Apply these changes?", true)

	case stateAddServiceConfirm:
		if m.confirmInput.Value() {
			// Save changes
			m.composeFile.AddService(m.service)
			if err := m.composeFile.WriteComposeFile(m.composePath); err != nil {
				m.err = err
				m.state = stateAddServiceSuccess
				return m, nil
			}
			m.state = stateAddServiceSuccess
		} else {
			return newAddMenuModel(), nil
		}

	case stateAddServiceSuccess:
		return newAddMenuModel(), nil
	}

	return m, nil
}

func (m addServiceModel) createRestartPolicySelect() (tea.Model, tea.Cmd) {
	policies := []string{"no", "always", "on-failure", "unless-stopped"}
	items := make([]list.Item, len(policies))
	for i, policy := range policies {
		items[i] = menuItem{title: policy, desc: "Restart policy", id: policy}
	}
	m.selectList = list.New(items, list.NewDefaultDelegate(), 0, 0)
	m.selectList.Title = "Select Restart Policy"
	m.selectList.SetShowStatusBar(false)
	m.selectList.Styles.Title = titleStyle
	return m, nil
}

func (m addServiceModel) generatePreview() (tea.Model, tea.Cmd) {
	// Add service to compose file temporarily for preview
	tempCompose := *m.composeFile
	tempCompose.AddService(m.service)

	// Marshal to YAML
	yamlData, err := yaml.Marshal(&tempCompose)
	if err != nil {
		m.err = fmt.Errorf("failed to generate preview: %w", err)
		m.state = stateAddServiceSuccess
		return m, nil
	}

	m.yamlPreview = string(yamlData)
	m.state = stateAddServicePreview

	// Create viewport for preview
	m.previewPort = viewport.New(80, 20)
	m.previewPort.SetContent(m.yamlPreview)

	return m, nil
}

func (m addServiceModel) View() string {
	if m.quitting {
		return "Cancelled.\n"
	}

	switch m.state {
	case stateAddServiceInit:
		cwd, _ := os.Getwd()
		s := "Loading docker-compose.yml...\n\n"
		s += fmt.Sprintf("Current directory: %s\n", cwd)
		return docStyle.Render(s)

	case stateAddServiceName, stateAddServiceImageName,
		stateAddServiceBuildContext, stateAddServiceBuildDockerfile,
		stateAddServiceAdvancedCommand, stateAddServiceAdvancedWorkdir,
		stateAddServiceAdvancedUser, stateAddServiceAdvancedHostname:
		return m.textInput.View()

	case stateAddServiceImageOrBuild, stateAddServicePortsConfirm,
		stateAddServiceEnvConfirm, stateAddServiceVolumesConfirm,
		stateAddServiceNetworksConfirm, stateAddServiceDepsConfirm,
		stateAddServiceAdvancedConfirm, stateAddServiceConfirm:
		return m.confirmInput.View()

	case stateAddServicePorts, stateAddServiceVolumes, stateAddServiceNetworks:
		return m.multiInput.View()

	case stateAddServiceEnv:
		return m.kvInput.View()

	case stateAddServiceDeps, stateAddServiceRestart:
		return docStyle.Render(m.selectList.View())

	case stateAddServicePreview:
		s := titleStyle.Render("Preview docker-compose.yml") + "\n\n"
		s += m.previewPort.View() + "\n\n"
		s += helpStyle.Render("↑↓ to scroll • 'enter' to continue • 'esc' to go back")
		return docStyle.Render(s)

	case stateAddServiceSuccess:
		if m.err != nil {
			s := titleStyle.Render("Error") + "\n\n"
			s += fmt.Sprintf("❌ %v\n\n", m.err)
			s += helpStyle.Render("Press 'enter' or 'esc' to return")
			return docStyle.Render(s)
		}
		s := lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Render(fmt.Sprintf("✅ Service '%s' added successfully!", m.service.Name)) + "\n\n"
		s += fmt.Sprintf("Service has been added to docker-compose.yml\n\n")
		s += helpStyle.Render("Press 'enter' or 'esc' to return to menu")
		return docStyle.Render(s)

	default:
		return "Loading..."
	}
}
