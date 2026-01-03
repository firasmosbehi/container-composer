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

// State constants for add volume wizard
const (
	stateAddVolumeInit = iota
	stateAddVolumeName
	stateAddVolumeDriver
	stateAddVolumeCustomDriver
	stateAddVolumeExternal
	stateAddVolumePreview
	stateAddVolumeConfirm
	stateAddVolumeSuccess
)

type addVolumeModel struct {
	state int

	// Data
	composeFile  *core.ComposeFile
	composePath  string
	volume       core.Volume
	volumeName   string
	customDriver bool

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

func newAddVolumeModel() addVolumeModel {
	return addVolumeModel{
		state:       stateAddVolumeInit,
		composePath: "docker-compose.yml",
	}
}

func (m addVolumeModel) Init() tea.Cmd {
	return func() tea.Msg {
		return m.loadComposeFile()
	}
}

type addVolumeErrorMsg struct{ err error }
type addVolumeComposeFileLoaded struct{ file *core.ComposeFile }

func (m addVolumeModel) loadComposeFile() tea.Msg {
	cwd, _ := os.Getwd()
	// Check if docker-compose.yml exists
	if _, err := os.Stat(m.composePath); os.IsNotExist(err) {
		return addVolumeErrorMsg{fmt.Errorf("docker-compose.yml not found in %s", cwd)}
	}

	// Parse existing docker-compose.yml
	composeFile, err := core.ParseComposeFile(m.composePath)
	if err != nil {
		return addVolumeErrorMsg{fmt.Errorf("failed to parse docker-compose.yml in %s: %w", cwd, err)}
	}

	return addVolumeComposeFileLoaded{composeFile}
}

func (m addVolumeModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case addVolumeErrorMsg:
		m.err = msg.err
		m.state = stateAddVolumeSuccess
		return m, nil

	case addVolumeComposeFileLoaded:
		m.composeFile = msg.file
		m.state = stateAddVolumeName
		// Initialize volume name input
		m.textInput = newTextInputForm(
			"Volume Name",
			"my-volume",
			"",
			"Unique identifier for this volume",
			func(s string) error {
				if s == "" {
					return fmt.Errorf("volume name cannot be empty")
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
			if m.state != stateAddVolumeName && m.state != stateAddVolumeCustomDriver {
				m.quitting = true
				return m, tea.Quit
			}

		case "esc":
			return m.handleBack()

		case "enter":
			return m.handleEnter()
		}

	case tea.WindowSizeMsg:
		if m.state == stateAddVolumePreview {
			m.previewPort.Width = msg.Width - 4
			m.previewPort.Height = msg.Height - 8
		}
	}

	// Update active component
	return m.updateComponent(msg)
}

func (m addVolumeModel) updateComponent(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch m.state {
	case stateAddVolumeName, stateAddVolumeCustomDriver:
		m.textInput, cmd = m.textInput.Update(msg)

	case stateAddVolumeExternal, stateAddVolumeConfirm:
		m.confirmInput, cmd = m.confirmInput.Update(msg)

	case stateAddVolumeDriver:
		m.selectList, cmd = m.selectList.Update(msg)

	case stateAddVolumePreview:
		m.previewPort, cmd = m.previewPort.Update(msg)
	}

	return m, cmd
}

func (m addVolumeModel) handleBack() (tea.Model, tea.Cmd) {
	switch m.state {
	case stateAddVolumeName, stateAddVolumeSuccess:
		return newAddMenuModel(), nil
	case stateAddVolumeDriver:
		m.state = stateAddVolumeName
	case stateAddVolumeCustomDriver:
		m.state = stateAddVolumeDriver
	case stateAddVolumeExternal:
		if m.customDriver {
			m.state = stateAddVolumeCustomDriver
		} else {
			m.state = stateAddVolumeDriver
		}
	case stateAddVolumePreview:
		m.state = stateAddVolumeExternal
	case stateAddVolumeConfirm:
		m.state = stateAddVolumePreview
	default:
		return newAddMenuModel(), nil
	}
	return m, nil
}

func (m addVolumeModel) handleEnter() (tea.Model, tea.Cmd) {
	switch m.state {
	case stateAddVolumeName:
		volumeName := m.textInput.Value()
		if volumeName == "" {
			return m, nil
		}

		// Check for conflicts
		if m.composeFile.VolumeExists(volumeName) {
			m.err = fmt.Errorf("volume '%s' already exists", volumeName)
			return m, nil
		}

		m.volumeName = volumeName
		m.state = stateAddVolumeDriver

		// Create driver selection list
		drivers := []string{"local", "nfs", "custom"}
		items := make([]list.Item, len(drivers))
		for i, driver := range drivers {
			var desc string
			switch driver {
			case "local":
				desc = "Default Docker volume driver"
			case "nfs":
				desc = "Network File System"
			case "custom":
				desc = "Enter custom driver name"
			}
			items[i] = menuItem{title: driver, desc: desc, id: driver}
		}
		m.selectList = list.New(items, list.NewDefaultDelegate(), 0, 0)
		m.selectList.Title = "Select Volume Driver"
		m.selectList.SetShowStatusBar(false)
		m.selectList.Styles.Title = titleStyle

	case stateAddVolumeDriver:
		if i, ok := m.selectList.SelectedItem().(menuItem); ok {
			if i.id == "custom" {
				m.customDriver = true
				m.state = stateAddVolumeCustomDriver
				m.textInput = newTextInputForm(
					"Custom Driver Name",
					"",
					"",
					"Enter the name of the custom driver",
					func(s string) error {
						if s == "" {
							return fmt.Errorf("driver name cannot be empty")
						}
						return nil
					},
				)
			} else {
				m.volume.Driver = i.id
				m.state = stateAddVolumeExternal
				m.confirmInput = newConfirmForm("Is this an external volume?", false)
			}
		}

	case stateAddVolumeCustomDriver:
		customDriver := m.textInput.Value()
		if customDriver == "" {
			return m, nil
		}
		m.volume.Driver = customDriver
		m.state = stateAddVolumeExternal
		m.confirmInput = newConfirmForm("Is this an external volume?", false)

	case stateAddVolumeExternal:
		m.volume.External = m.confirmInput.Value()
		return m.generatePreview()

	case stateAddVolumePreview:
		m.state = stateAddVolumeConfirm
		m.confirmInput = newConfirmForm("Apply these changes?", true)

	case stateAddVolumeConfirm:
		if m.confirmInput.Value() {
			// Save changes
			m.composeFile.AddVolume(m.volumeName, m.volume)
			if err := m.composeFile.WriteComposeFile(m.composePath); err != nil {
				m.err = err
				m.state = stateAddVolumeSuccess
				return m, nil
			}
			m.state = stateAddVolumeSuccess
		} else {
			return newAddMenuModel(), nil
		}

	case stateAddVolumeSuccess:
		return newAddMenuModel(), nil
	}

	return m, nil
}

func (m addVolumeModel) generatePreview() (tea.Model, tea.Cmd) {
	// Add volume to compose file temporarily for preview
	tempCompose := *m.composeFile
	tempCompose.AddVolume(m.volumeName, m.volume)

	// Marshal to YAML
	yamlData, err := yaml.Marshal(&tempCompose)
	if err != nil {
		m.err = fmt.Errorf("failed to generate preview: %w", err)
		m.state = stateAddVolumeSuccess
		return m, nil
	}

	m.yamlPreview = string(yamlData)
	m.state = stateAddVolumePreview

	// Create viewport for preview
	m.previewPort = viewport.New(80, 20)
	m.previewPort.SetContent(m.yamlPreview)

	return m, nil
}

func (m addVolumeModel) View() string {
	if m.quitting {
		return "Cancelled.\n"
	}

	switch m.state {
	case stateAddVolumeInit:
		return docStyle.Render("Loading docker-compose.yml...")

	case stateAddVolumeName, stateAddVolumeCustomDriver:
		return m.textInput.View()

	case stateAddVolumeExternal, stateAddVolumeConfirm:
		return m.confirmInput.View()

	case stateAddVolumeDriver:
		return docStyle.Render(m.selectList.View())

	case stateAddVolumePreview:
		s := titleStyle.Render("Preview docker-compose.yml") + "\n\n"
		s += m.previewPort.View() + "\n\n"
		s += helpStyle.Render("↑↓ to scroll • 'enter' to continue • 'esc' to go back")
		return docStyle.Render(s)

	case stateAddVolumeSuccess:
		if m.err != nil {
			s := titleStyle.Render("Error") + "\n\n"
			s += fmt.Sprintf("❌ %v\n\n", m.err)
			s += helpStyle.Render("Press 'enter' or 'esc' to return")
			return docStyle.Render(s)
		}
		s := lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Render(fmt.Sprintf("✅ Volume '%s' added successfully!", m.volumeName)) + "\n\n"
		s += fmt.Sprintf("Volume has been added to docker-compose.yml\n\n")
		s += helpStyle.Render("Press 'enter' or 'esc' to return to menu")
		return docStyle.Render(s)

	default:
		return "Loading..."
	}
}
