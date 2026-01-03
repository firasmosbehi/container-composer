package tui

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

type addMenuModel struct {
	list     list.Model
	quitting bool
}

func newAddMenuModel() addMenuModel {
	items := []list.Item{
		menuItem{
			title: "🐳 Add Service",
			desc:  "Add a new container/service to docker-compose.yml",
			id:    "service",
		},
		menuItem{
			title: "🌐 Add Network",
			desc:  "Add a new network definition",
			id:    "network",
		},
		menuItem{
			title: "💾 Add Volume",
			desc:  "Add a new volume definition",
			id:    "volume",
		},
	}

	// Create list with delegate
	delegate := list.NewDefaultDelegate()
	// Set large initial height to ensure all items are visible
	l := list.New(items, delegate, 100, 50)
	l.Title = "Add Resources"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowPagination(false)
	l.SetShowHelp(true)
	l.Styles.Title = titleStyle

	return addMenuModel{list: l}
}

func (m addMenuModel) Init() tea.Cmd {
	return nil
}

func (m addMenuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit

		case "esc":
			return newMainMenuModel(), nil

		case "enter":
			if i, ok := m.list.SelectedItem().(menuItem); ok {
				switch i.id {
				case "service":
					newModel := newAddServiceModel()
					return newModel, newModel.Init()
				case "network":
					newModel := newAddNetworkModel()
					return newModel, newModel.Init()
				case "volume":
					newModel := newAddVolumeModel()
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

func (m addMenuModel) View() string {
	if m.quitting {
		return "Cancelled.\n"
	}
	return docStyle.Render(m.list.View())
}
