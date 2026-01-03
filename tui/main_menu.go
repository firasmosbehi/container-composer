package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/list"
)

type menuItem struct {
	title string
	desc  string
	id    string
}

func (i menuItem) Title() string       { return i.title }
func (i menuItem) Description() string { return i.desc }
func (i menuItem) FilterValue() string { return i.title }

type mainMenuModel struct {
	list     list.Model
	selected string
	quitting bool
}

func newMainMenuModel() mainMenuModel {
	items := []list.Item{
		menuItem{
			title: "🚀 Initialize New Project",
			desc:  "Create a new Docker Compose project from templates",
			id:    "init",
		},
		menuItem{
			title: "➕ Add Resources",
			desc:  "Add services, networks, or volumes to existing project",
			id:    "add",
		},
		menuItem{
			title: "📊 View Dependency Graph",
			desc:  "Visualize service dependencies and relationships",
			id:    "graph",
		},
	}

	// Create list with delegate
	delegate := list.NewDefaultDelegate()
	// Set large initial height to ensure all items are visible
	l := list.New(items, delegate, 100, 50)
	l.Title = "Container Composer TUI"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowPagination(false)
	l.SetShowHelp(true)
	l.Styles.Title = titleStyle

	return mainMenuModel{list: l}
}

func (m mainMenuModel) Init() tea.Cmd {
	return nil
}

func (m mainMenuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit

		case "enter":
			// Get selected item
			if i, ok := m.list.SelectedItem().(menuItem); ok {
				m.selected = i.id

				// Transition to different views based on selection
				switch i.id {
				case "init":
					return newInitModel(), nil
				case "add":
					return newAddMenuModel(), nil
				case "graph":
					newModel := newDependencyGraphModel()
					return newModel, newModel.Init()
				}
			}
		}

	case tea.WindowSizeMsg:
		// Use full screen (minus 1 line for margins)
		m.list.SetSize(msg.Width-2, msg.Height-2)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m mainMenuModel) View() string {
	if m.quitting {
		return "Goodbye! 👋\n"
	}
	return docStyle.Render(m.list.View())
}
