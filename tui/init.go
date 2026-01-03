package tui

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/firasmosbahi/container-composer/templates"
)

const (
	stateInitSelection = iota
	stateInitCreating
	stateInitSuccess
)

type initModel struct {
	state          int
	templates      list.Model
	projectInput   textinput.Model
	selectedTmpl   string
	projectName    string
	creationStatus string
	quitting       bool
	err            error
	focusOnInput   bool
}

// Custom list item with category badge
type templateListItem struct {
	template templates.Template
	category string
}

func (i templateListItem) FilterValue() string { return i.template.Name }
func (i templateListItem) Title() string {
	categoryBadge := getCategoryBadge(i.category)
	return fmt.Sprintf("%s  %s", categoryBadge, i.template.Name)
}
func (i templateListItem) Description() string { return i.template.Description }

func getCategoryBadge(category string) string {
	switch category {
	case templates.CategoryStarter:
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00D7D7")).
			Bold(true).
			Render("[STARTER]")
	case templates.CategoryFullStack:
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7D56F4")).
			Bold(true).
			Render("[FULL-STACK]")
	case templates.CategoryWeb:
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("#04B575")).
			Bold(true).
			Render("[WEB]")
	case templates.CategoryMicroservice:
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFD700")).
			Bold(true).
			Render("[MICROSERVICES]")
	default:
		return "[OTHER]"
	}
}

func newInitModel() initModel {
	// Load all templates at once (no category filtering)
	tmpls := templates.GetAvailableTemplates()
	items := make([]list.Item, len(tmpls))
	for i, tmpl := range tmpls {
		items[i] = templateListItem{
			template: tmpl,
			category: tmpl.Category,
		}
	}

	// Create list with delegate
	delegate := list.NewDefaultDelegate()
	// Set large initial height to ensure all items are visible
	l := list.New(items, delegate, 100, 50)
	l.Title = "🚀 Select Template"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)
	l.SetShowPagination(false)
	l.SetShowHelp(true)
	l.Styles.Title = titleStyle

	// Create project name input
	ti := textinput.New()
	ti.Placeholder = "my-project"
	ti.CharLimit = 50
	ti.Width = 50

	return initModel{
		state:        stateInitSelection,
		templates:    l,
		projectInput: ti,
		focusOnInput: false,
	}
}

func (m initModel) Init() tea.Cmd {
	return nil
}

func (m initModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case projectCreatedMsg:
		// Project created successfully, return to main menu
		return newMainMenuModel(), nil

	case tea.KeyMsg:
		switch m.state {
		case stateInitSelection:
			switch msg.String() {
			case "ctrl+c", "q":
				m.quitting = true
				return m, tea.Quit

			case "esc":
				if m.focusOnInput {
					// Unfocus input, back to list
					m.focusOnInput = false
					m.projectInput.Blur()
					return m, nil
				}
				// Go back to main menu
				return newMainMenuModel(), nil

			case "tab":
				// Toggle between list and input
				m.focusOnInput = !m.focusOnInput
				if m.focusOnInput {
					return m, m.projectInput.Focus()
				}
				m.projectInput.Blur()
				return m, nil

			case "enter":
				if m.focusOnInput {
					// Create project with current selection and project name
					return m.startCreation()
				}
				// Select template and move to input
				if i, ok := m.templates.SelectedItem().(templateListItem); ok {
					m.selectedTmpl = i.template.Name
					m.focusOnInput = true
					return m, m.projectInput.Focus()
				}
			}

		case stateInitSuccess:
			// Allow manual return to main menu even during auto-redirect
			if msg.String() == "enter" || msg.String() == "esc" || msg.String() == "q" {
				return newMainMenuModel(), nil
			}
		}

	case tea.WindowSizeMsg:
		if m.state == stateInitSelection {
			// Calculate available space for the template list
			// Reserve space for: title (3 lines) + input section (5 lines) + help (1 line)
			reservedLines := 9
			listHeight := msg.Height - reservedLines
			if listHeight < 8 {
				listHeight = 8
			}
			// Use almost full width
			listWidth := msg.Width - 2
			if listWidth < 60 {
				listWidth = 60
			}
			m.templates.SetSize(listWidth, listHeight)
		}
	}

	// Update components based on focus
	var cmd tea.Cmd
	if m.state == stateInitSelection {
		if m.focusOnInput {
			m.projectInput, cmd = m.projectInput.Update(msg)
		} else {
			m.templates, cmd = m.templates.Update(msg)
		}
	}

	return m, cmd
}

func (m initModel) startCreation() (tea.Model, tea.Cmd) {
	// Get project name from input (or use default)
	m.projectName = strings.TrimSpace(m.projectInput.Value())
	if m.projectName == "" {
		m.projectName = "my-project"
	}

	// Ensure we have a selected template
	if m.selectedTmpl == "" {
		if i, ok := m.templates.SelectedItem().(templateListItem); ok {
			m.selectedTmpl = i.template.Name
		} else {
			return m, nil // No template selected
		}
	}

	m.state = stateInitCreating
	return m.createProject()
}

type projectCreatedMsg struct {
	projectName string
	templateName string
}

func (m initModel) createProject() (tea.Model, tea.Cmd) {
	// Get the template
	tmpl, err := templates.GetTemplate(m.selectedTmpl)
	if err != nil {
		m.err = err
		m.state = stateInitSuccess // Show error state
		return m, nil
	}

	// Check if directory exists and is not empty
	projectDir := m.projectName
	if _, err := os.Stat(projectDir); err == nil {
		// Directory exists, check if empty
		entries, err := os.ReadDir(projectDir)
		if err != nil {
			m.err = fmt.Errorf("failed to read directory: %w", err)
			m.state = stateInitSuccess
			return m, nil
		}
		if len(entries) > 0 {
			m.err = fmt.Errorf("directory '%s' is not empty", projectDir)
			m.state = stateInitSuccess
			return m, nil
		}
	}

	// Generate project from template
	vars := templates.TemplateVars{ProjectName: m.projectName}
	if err := tmpl.Generate(m.projectName, vars); err != nil {
		m.err = err
		m.state = stateInitSuccess
		return m, nil
	}

	// Change to the newly created project directory
	if err := os.Chdir(m.projectName); err != nil {
		m.err = fmt.Errorf("project created but failed to navigate to directory: %w", err)
		m.state = stateInitSuccess
		return m, nil
	}

	// Success! Return to main menu after showing success message briefly
	m.creationStatus = fmt.Sprintf("✅ Project '%s' created successfully!", m.projectName)
	m.state = stateInitSuccess

	// Return command to go back to main menu after a brief delay
	return m, tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
		return projectCreatedMsg{
			projectName:  m.projectName,
			templateName: m.selectedTmpl,
		}
	})
}

func (m initModel) View() string {
	if m.quitting {
		return "\n  ✓ Cancelled\n\n"
	}

	switch m.state {
	case stateInitSelection:
		var b strings.Builder

		// Title
		title := lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7D56F4")).
			Render("📦 CREATE NEW PROJECT")
		b.WriteString(title + "\n")

		// Info text
		info := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#9CA3AF")).
			Render("Choose a template to get started")
		b.WriteString(info + "\n\n")

		// Template list - render it directly
		b.WriteString(m.templates.View())
		b.WriteString("\n\n")

		// Project name input section
		inputLabel := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00D7D7")).
			Bold(true).
			Render("📝 Project Name:")

		b.WriteString(inputLabel + " ")

		inputBox := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(getInputBorderColor(m.focusOnInput))).
			Padding(0, 1).
			Render(m.projectInput.View())

		b.WriteString(inputBox + "\n\n")

		// Help text with better formatting
		var helpText string
		if m.focusOnInput {
			helpText = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#00D7D7")).
				Render("⏎") + " " +
				lipgloss.NewStyle().Foreground(lipgloss.Color("#9CA3AF")).Render("create project  ") +
				lipgloss.NewStyle().Foreground(lipgloss.Color("#00D7D7")).Render("TAB") + " " +
				lipgloss.NewStyle().Foreground(lipgloss.Color("#9CA3AF")).Render("switch to list  ") +
				lipgloss.NewStyle().Foreground(lipgloss.Color("#00D7D7")).Render("ESC") + " " +
				lipgloss.NewStyle().Foreground(lipgloss.Color("#9CA3AF")).Render("cancel")
		} else {
			helpText = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#00D7D7")).
				Render("⏎") + " " +
				lipgloss.NewStyle().Foreground(lipgloss.Color("#9CA3AF")).Render("select  ") +
				lipgloss.NewStyle().Foreground(lipgloss.Color("#00D7D7")).Render("TAB") + " " +
				lipgloss.NewStyle().Foreground(lipgloss.Color("#9CA3AF")).Render("to name  ") +
				lipgloss.NewStyle().Foreground(lipgloss.Color("#00D7D7")).Render("/") + " " +
				lipgloss.NewStyle().Foreground(lipgloss.Color("#9CA3AF")).Render("filter  ") +
				lipgloss.NewStyle().Foreground(lipgloss.Color("#00D7D7")).Render("Q") + " " +
				lipgloss.NewStyle().Foreground(lipgloss.Color("#9CA3AF")).Render("quit")
		}
		b.WriteString(helpText)

		return docStyle.Render(b.String())

	case stateInitCreating:
		var b strings.Builder

		title := lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFD700")).
			Render("⚙️  CREATING PROJECT")
		b.WriteString(title + "\n\n")

		spinner := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7D56F4")).
			Render("◐ ◓ ◑ ◒")
		b.WriteString(spinner + "\n\n")

		info := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FAFAFA")).
			Render(fmt.Sprintf("Template:     %s\nProject Name: %s\n",
				lipgloss.NewStyle().Bold(true).Render(m.selectedTmpl),
				lipgloss.NewStyle().Bold(true).Render(m.projectName)))
		b.WriteString(info)

		b.WriteString("\nGenerating files...")

		return docStyle.Render(b.String())

	case stateInitSuccess:
		var b strings.Builder

		if m.err != nil {
			// Error state - stays on screen until user presses a key
			title := lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#FF5F87")).
				Render("❌ ERROR")
			b.WriteString("\n\n" + title + "\n\n")

			errBox := lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#FF5F87")).
				Padding(1, 2).
				Foreground(lipgloss.Color("#FF5F87")).
				Render(m.err.Error())
			b.WriteString(errBox + "\n\n")

			help := lipgloss.NewStyle().
				Foreground(lipgloss.Color("#9CA3AF")).
				Render("Press ↵, esc, or q to return to main menu")
			b.WriteString(help)

			return docStyle.Render(b.String())
		}

		// Success state - brief message before auto-redirect
		b.WriteString("\n\n\n\n")

		// Large success icon
		successIcon := lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#04B575")).
			Render("    ✓")
		b.WriteString(successIcon + "\n\n")

		// Success message
		title := lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#04B575")).
			Render("  PROJECT CREATED!")
		b.WriteString(title + "\n\n")

		// Project info
		projectInfo := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FAFAFA")).
			Render(fmt.Sprintf("  📁 %s", m.projectName))
		b.WriteString(projectInfo + "\n")

		templateInfo := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#9CA3AF")).
			Render(fmt.Sprintf("  🎨 %s template", m.selectedTmpl))
		b.WriteString(templateInfo + "\n\n")

		// Show current directory
		cwd, _ := os.Getwd()
		cwdInfo := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00D7D7")).
			Render(fmt.Sprintf("  📂 Working directory: %s", cwd))
		b.WriteString(cwdInfo + "\n\n")

		// Auto-redirect message
		redirectMsg := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#9CA3AF")).
			Italic(true).
			Render("  Returning to main menu...")
		b.WriteString(redirectMsg + "\n\n")

		// Optional skip message
		skipMsg := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262")).
			Render("  (Press any key to skip)")
		b.WriteString(skipMsg)

		return docStyle.Render(b.String())

	default:
		return "Loading..."
	}
}

func getInputBorderColor(focused bool) string {
	if focused {
		return "#7D56F4" // Purple when focused
	}
	return "#626262" // Gray when not focused
}
