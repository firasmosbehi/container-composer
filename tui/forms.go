package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Validation function type
type ValidatorFunc func(string) error

// textInputForm is a reusable text input component with validation
type textInputForm struct {
	input     textinput.Model
	title     string
	help      string
	validator ValidatorFunc
	err       error
}

func newTextInputForm(title, placeholder, defaultValue, help string, validator ValidatorFunc) textInputForm {
	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.SetValue(defaultValue)
	ti.Focus()
	ti.CharLimit = 256
	ti.Width = 50

	return textInputForm{
		input:     ti,
		title:     title,
		help:      help,
		validator: validator,
	}
}

func (m textInputForm) Init() tea.Cmd {
	return textinput.Blink
}

func (m textInputForm) Update(msg tea.Msg) (textInputForm, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			value := strings.TrimSpace(m.input.Value())
			if m.validator != nil {
				if err := m.validator(value); err != nil {
					m.err = err
					return m, nil
				}
			}
			m.err = nil
		}
	}

	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m textInputForm) View() string {
	s := titleStyle.Render(m.title) + "\n\n"
	s += m.input.View() + "\n\n"

	if m.err != nil {
		errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
		s += errorStyle.Render(fmt.Sprintf("✗ %s", m.err.Error())) + "\n\n"
	}

	if m.help != "" {
		s += helpStyle.Render(m.help) + "\n\n"
	}

	s += helpStyle.Render("Press 'enter' to continue • 'esc' to go back • 'q' to quit")
	return docStyle.Render(s)
}

func (m textInputForm) Value() string {
	return strings.TrimSpace(m.input.Value())
}

// confirmForm is a Yes/No confirmation prompt
type confirmForm struct {
	message string
	choice  bool
}

func newConfirmForm(message string, defaultChoice bool) confirmForm {
	return confirmForm{
		message: message,
		choice:  defaultChoice,
	}
}

func (m confirmForm) Init() tea.Cmd {
	return nil
}

func (m confirmForm) Update(msg tea.Msg) (confirmForm, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "left", "h":
			m.choice = false
		case "right", "l":
			m.choice = true
		case "y":
			m.choice = true
		case "n":
			m.choice = false
		}
	}
	return m, nil
}

func (m confirmForm) View() string {
	s := titleStyle.Render(m.message) + "\n\n"

	yesStyle := menuItemStyle
	noStyle := menuItemStyle

	if m.choice {
		yesStyle = selectedItemStyle
	} else {
		noStyle = selectedItemStyle
	}

	s += yesStyle.Render("[ Yes ]") + "  " + noStyle.Render("[ No ]") + "\n\n"
	s += helpStyle.Render("← → or y/n to select • 'enter' to confirm • 'esc' to go back • 'q' to quit")

	return docStyle.Render(s)
}

func (m confirmForm) Value() bool {
	return m.choice
}

// multiInputForm collects multiple string inputs (e.g., ports, volumes)
type multiInputForm struct {
	title       string
	help        string
	placeholder string
	input       textinput.Model
	values      []string
	addingNew   bool
	validator   ValidatorFunc
	err         error
}

func newMultiInputForm(title, placeholder, help string, validator ValidatorFunc) multiInputForm {
	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.CharLimit = 256
	ti.Width = 50
	ti.Focus()

	return multiInputForm{
		title:       title,
		help:        help,
		placeholder: placeholder,
		input:       ti,
		values:      []string{},
		addingNew:   true,
		validator:   validator,
	}
}

func (m multiInputForm) Init() tea.Cmd {
	return textinput.Blink
}

func (m multiInputForm) Update(msg tea.Msg) (multiInputForm, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			value := strings.TrimSpace(m.input.Value())
			if value != "" {
				if m.validator != nil {
					if err := m.validator(value); err != nil {
						m.err = err
						return m, nil
					}
				}
				m.values = append(m.values, value)
				m.input.SetValue("")
				m.err = nil
			}
		}
	}

	if m.addingNew {
		m.input, cmd = m.input.Update(msg)
	}
	return m, cmd
}

func (m multiInputForm) View() string {
	s := titleStyle.Render(m.title) + "\n\n"

	if len(m.values) > 0 {
		s += "Added:\n"
		for _, v := range m.values {
			s += lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Render("  ✓ "+v) + "\n"
		}
		s += "\n"
	}

	s += m.input.View() + "\n\n"

	if m.err != nil {
		errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
		s += errorStyle.Render(fmt.Sprintf("✗ %s", m.err.Error())) + "\n\n"
	}

	if m.help != "" {
		s += helpStyle.Render(m.help) + "\n\n"
	}

	s += helpStyle.Render("Press 'enter' to add • leave empty and press 'enter' to finish • 'esc' to go back")
	return docStyle.Render(s)
}

func (m multiInputForm) Values() []string {
	return m.values
}

func (m multiInputForm) HasValues() bool {
	return len(m.values) > 0
}

func (m multiInputForm) IsDone() bool {
	// Done if input is empty (user pressed Enter with empty field)
	return strings.TrimSpace(m.input.Value()) == ""
}

func (m multiInputForm) Focus() tea.Cmd {
	return m.input.Focus()
}

// keyValueInputForm collects key-value pairs (e.g., environment variables)
type keyValueInputForm struct {
	title        string
	help         string
	keyInput     textinput.Model
	valueInput   textinput.Model
	pairs        map[string]string
	pairKeys     []string // To maintain order
	focusedInput int      // 0 = key, 1 = value
	err          error
}

func newKeyValueInputForm(title, help string) keyValueInputForm {
	keyInput := textinput.New()
	keyInput.Placeholder = "KEY"
	keyInput.CharLimit = 256
	keyInput.Width = 30
	keyInput.Focus()

	valueInput := textinput.New()
	valueInput.Placeholder = "value"
	valueInput.CharLimit = 512
	valueInput.Width = 50

	return keyValueInputForm{
		title:        title,
		help:         help,
		keyInput:     keyInput,
		valueInput:   valueInput,
		pairs:        make(map[string]string),
		pairKeys:     []string{},
		focusedInput: 0,
	}
}

func (m keyValueInputForm) Init() tea.Cmd {
	return textinput.Blink
}

func (m keyValueInputForm) Update(msg tea.Msg) (keyValueInputForm, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyTab:
			// Switch between key and value inputs
			if m.focusedInput == 0 {
				m.focusedInput = 1
				m.keyInput.Blur()
				m.valueInput.Focus()
			} else {
				m.focusedInput = 0
				m.valueInput.Blur()
				m.keyInput.Focus()
			}
			return m, nil

		case tea.KeyEnter:
			key := strings.TrimSpace(m.keyInput.Value())
			value := strings.TrimSpace(m.valueInput.Value())

			if key == "" && value == "" {
				// Both empty = done adding
				return m, nil
			}

			if key != "" {
				// Add the pair
				if _, exists := m.pairs[key]; !exists {
					m.pairKeys = append(m.pairKeys, key)
				}
				m.pairs[key] = value
				m.keyInput.SetValue("")
				m.valueInput.SetValue("")
				m.focusedInput = 0
				m.valueInput.Blur()
				m.keyInput.Focus()
				m.err = nil
			} else {
				m.err = fmt.Errorf("key cannot be empty")
			}
			return m, nil
		}
	}

	if m.focusedInput == 0 {
		m.keyInput, cmd = m.keyInput.Update(msg)
	} else {
		m.valueInput, cmd = m.valueInput.Update(msg)
	}

	return m, cmd
}

func (m keyValueInputForm) View() string {
	s := titleStyle.Render(m.title) + "\n\n"

	if len(m.pairs) > 0 {
		s += "Added:\n"
		for _, k := range m.pairKeys {
			s += lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Render(fmt.Sprintf("  ✓ %s=%s", k, m.pairs[k])) + "\n"
		}
		s += "\n"
	}

	s += m.keyInput.View() + " = " + m.valueInput.View() + "\n\n"

	if m.err != nil {
		errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
		s += errorStyle.Render(fmt.Sprintf("✗ %s", m.err.Error())) + "\n\n"
	}

	if m.help != "" {
		s += helpStyle.Render(m.help) + "\n\n"
	}

	s += helpStyle.Render("'tab' to switch fields • 'enter' to add • leave empty and press 'enter' to finish • 'esc' to go back")
	return docStyle.Render(s)
}

func (m keyValueInputForm) Values() map[string]string {
	return m.pairs
}

func (m keyValueInputForm) IsDone() bool {
	// Done if both inputs are empty (user pressed Enter with empty fields)
	return strings.TrimSpace(m.keyInput.Value()) == "" && strings.TrimSpace(m.valueInput.Value()) == ""
}

func (m keyValueInputForm) HasValues() bool {
	return len(m.pairs) > 0
}
