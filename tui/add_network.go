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

// State constants for add network wizard
const (
	stateAddNetworkInit = iota
	stateAddNetworkName
	stateAddNetworkDriver
	stateAddNetworkExternal
	stateAddNetworkPreview
	stateAddNetworkConfirm
	stateAddNetworkSuccess
)

type addNetworkModel struct {
	state int

	// Data
	composeFile *core.ComposeFile
	composePath string
	network     core.Network
	networkName string

	// Form components
	textInput    textInputForm
	confirmInput confirmForm
	selectList   list.Model
	previewPort  viewport.Model

	// Preview
	yamlPreview string

	// Status
	err      error
	quitting bool
}

func newAddNetworkModel() addNetworkModel {
	return addNetworkModel{
		state:       stateAddNetworkInit,
		composePath: "docker-compose.yml",
	}
}

func (m addNetworkModel) Init() tea.Cmd {
	return func() tea.Msg {
		return m.loadComposeFile()
	}
}

type addNetworkErrorMsg struct{ err error }
type addNetworkComposeFileLoaded struct{ file *core.ComposeFile }

func (m addNetworkModel) loadComposeFile() tea.Msg {
	cwd, _ := os.Getwd()
	// Check if docker-compose.yml exists
	if _, err := os.Stat(m.composePath); os.IsNotExist(err) {
		return addNetworkErrorMsg{fmt.Errorf("docker-compose.yml not found in %s", cwd)}
	}

	// Parse existing docker-compose.yml
	composeFile, err := core.ParseComposeFile(m.composePath)
	if err != nil {
		return addNetworkErrorMsg{fmt.Errorf("failed to parse docker-compose.yml in %s: %w", cwd, err)}
	}

	return addNetworkComposeFileLoaded{composeFile}
}

func (m addNetworkModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case addNetworkErrorMsg:
		m.err = msg.err
		m.state = stateAddNetworkSuccess
		return m, nil

	case addNetworkComposeFileLoaded:
		m.composeFile = msg.file
		m.state = stateAddNetworkName
		// Initialize network name input
		m.textInput = newTextInputForm(
			"Network Name",
			"my-network",
			"",
			"Unique identifier for this network",
			func(s string) error {
				if s == "" {
					return fmt.Errorf("network name cannot be empty")
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
			if m.state != stateAddNetworkName {
				m.quitting = true
				return m, tea.Quit
			}

		case "esc":
			return m.handleBack()

		case "enter":
			return m.handleEnter()
		}

	case tea.WindowSizeMsg:
		if m.state == stateAddNetworkPreview {
			m.previewPort.Width = msg.Width - 4
			m.previewPort.Height = msg.Height - 8
		}
	}

	// Update active component
	return m.updateComponent(msg)
}

func (m addNetworkModel) updateComponent(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch m.state {
	case stateAddNetworkName:
		m.textInput, cmd = m.textInput.Update(msg)

	case stateAddNetworkExternal, stateAddNetworkConfirm:
		m.confirmInput, cmd = m.confirmInput.Update(msg)

	case stateAddNetworkDriver:
		m.selectList, cmd = m.selectList.Update(msg)

	case stateAddNetworkPreview:
		m.previewPort, cmd = m.previewPort.Update(msg)
	}

	return m, cmd
}

func (m addNetworkModel) handleBack() (tea.Model, tea.Cmd) {
	switch m.state {
	case stateAddNetworkName, stateAddNetworkSuccess:
		return newAddMenuModel(), nil
	case stateAddNetworkDriver:
		m.state = stateAddNetworkName
	case stateAddNetworkExternal:
		m.state = stateAddNetworkDriver
	case stateAddNetworkPreview:
		m.state = stateAddNetworkExternal
	case stateAddNetworkConfirm:
		m.state = stateAddNetworkPreview
	default:
		return newAddMenuModel(), nil
	}
	return m, nil
}

func (m addNetworkModel) handleEnter() (tea.Model, tea.Cmd) {
	switch m.state {
	case stateAddNetworkName:
		networkName := m.textInput.Value()
		if networkName == "" {
			return m, nil
		}

		// Check for conflicts
		if m.composeFile.NetworkExists(networkName) {
			m.err = fmt.Errorf("network '%s' already exists", networkName)
			return m, nil
		}

		m.networkName = networkName
		m.state = stateAddNetworkDriver

		// Create driver selection list
		drivers := []string{"bridge", "host", "overlay", "macvlan", "none"}
		items := make([]list.Item, len(drivers))
		for i, driver := range drivers {
			var desc string
			switch driver {
			case "bridge":
				desc = "Default Docker network driver"
			case "host":
				desc = "Use host's network stack"
			case "overlay":
				desc = "Multi-host networking"
			case "macvlan":
				desc = "Assign MAC address to container"
			case "none":
				desc = "Disable networking"
			}
			items[i] = menuItem{title: driver, desc: desc, id: driver}
		}
		m.selectList = list.New(items, list.NewDefaultDelegate(), 0, 0)
		m.selectList.Title = "Select Network Driver"
		m.selectList.SetShowStatusBar(false)
		m.selectList.Styles.Title = titleStyle

	case stateAddNetworkDriver:
		if i, ok := m.selectList.SelectedItem().(menuItem); ok {
			m.network.Driver = i.id
		}
		m.state = stateAddNetworkExternal
		m.confirmInput = newConfirmForm("Is this an external network?", false)

	case stateAddNetworkExternal:
		m.network.External = m.confirmInput.Value()
		return m.generatePreview()

	case stateAddNetworkPreview:
		m.state = stateAddNetworkConfirm
		m.confirmInput = newConfirmForm("Apply these changes?", true)

	case stateAddNetworkConfirm:
		if m.confirmInput.Value() {
			// Save changes
			m.composeFile.AddNetwork(m.networkName, m.network)
			if err := m.composeFile.WriteComposeFile(m.composePath); err != nil {
				m.err = err
				m.state = stateAddNetworkSuccess
				return m, nil
			}
			m.state = stateAddNetworkSuccess
		} else {
			return newAddMenuModel(), nil
		}

	case stateAddNetworkSuccess:
		return newAddMenuModel(), nil
	}

	return m, nil
}

func (m addNetworkModel) generatePreview() (tea.Model, tea.Cmd) {
	// Add network to compose file temporarily for preview
	tempCompose := *m.composeFile
	tempCompose.AddNetwork(m.networkName, m.network)

	// Marshal to YAML
	yamlData, err := yaml.Marshal(&tempCompose)
	if err != nil {
		m.err = fmt.Errorf("failed to generate preview: %w", err)
		m.state = stateAddNetworkSuccess
		return m, nil
	}

	m.yamlPreview = string(yamlData)
	m.state = stateAddNetworkPreview

	// Create viewport for preview
	m.previewPort = viewport.New(80, 20)
	m.previewPort.SetContent(m.yamlPreview)

	return m, nil
}

func (m addNetworkModel) View() string {
	if m.quitting {
		return "Cancelled.\n"
	}

	switch m.state {
	case stateAddNetworkInit:
		return docStyle.Render("Loading docker-compose.yml...")

	case stateAddNetworkName:
		return m.textInput.View()

	case stateAddNetworkExternal, stateAddNetworkConfirm:
		return m.confirmInput.View()

	case stateAddNetworkDriver:
		return docStyle.Render(m.selectList.View())

	case stateAddNetworkPreview:
		s := titleStyle.Render("Preview docker-compose.yml") + "\n\n"
		s += m.previewPort.View() + "\n\n"
		s += helpStyle.Render("↑↓ to scroll • 'enter' to continue • 'esc' to go back")
		return docStyle.Render(s)

	case stateAddNetworkSuccess:
		if m.err != nil {
			s := titleStyle.Render("Error") + "\n\n"
			s += fmt.Sprintf("❌ %v\n\n", m.err)
			s += helpStyle.Render("Press 'enter' or 'esc' to return")
			return docStyle.Render(s)
		}
		s := lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Render(fmt.Sprintf("✅ Network '%s' added successfully!", m.networkName)) + "\n\n"
		s += fmt.Sprintf("Network has been added to docker-compose.yml\n\n")
		s += helpStyle.Render("Press 'enter' or 'esc' to return to menu")
		return docStyle.Render(s)

	default:
		return "Loading..."
	}
}
