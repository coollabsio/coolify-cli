package cliinstances

import (
	"errors"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/coollabsio/cli-coolify/cmd/coolTypes"
	"github.com/coollabsio/cli-coolify/pkg/tui"
)

// addKeyMap defines keybindings for the add instance form
type addKeyMap struct {
	Up    key.Binding
	Down  key.Binding
	Tab   key.Binding
	Enter key.Binding
	Paste key.Binding
	Help  key.Binding
	Quit  key.Binding
}

// ShortHelp returns keybindings to be shown in the mini help view
func (k addKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Help, k.Quit}
}

// FullHelp returns keybindings for the expanded help view
func (k addKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Tab},      // first column
		{k.Enter, k.Paste, k.Help}, // second column
		{k.Quit},                   // third column
	}
}

var addKeys = addKeyMap{
	Up: key.NewBinding(
		key.WithKeys("up"),
		key.WithHelp("↑", "move up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down"),
		key.WithHelp("↓", "move down"),
	),
	Tab: key.NewBinding(
		key.WithKeys("tab", "shift+tab"),
		key.WithHelp("tab", "next field"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "submit/select"),
	),
	Paste: key.NewBinding(
		key.WithKeys("ctrl+v"),
		key.WithHelp("ctrl+v", "paste"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "toggle help"),
	),
	Quit: key.NewBinding(
		key.WithKeys("esc", "ctrl+c"),
		key.WithHelp("esc", "quit"),
	),
}

type addModel struct {
	inputs    []textinput.Model
	focus     int
	err       error
	instance  coolTypes.Instance
	width     int
	height    int
	result    chan<- coolTypes.Instance
	force     bool
	isDefault bool
	keys      addKeyMap
	help      help.Model
}

// Add a new command type for sending the instance
type sendInstanceMsg struct {
	instance coolTypes.Instance
}

func newAddModel(result chan<- coolTypes.Instance, force, isDefault bool) addModel {
	// Create text inputs
	inputs := make([]textinput.Model, 3)
	labels := []string{"Name", "FQDN", "Token"}

	for i, label := range labels {
		input := textinput.New()
		input.Placeholder = fmt.Sprintf("Enter instance %s", label)
		input.Prompt = fmt.Sprintf("%s: ", label)
		input.PromptStyle = FocusedStyle
		input.TextStyle = FocusedStyle

		// Set up validation for each input type
		switch label {
		case "Name":
			input.Validate = tui.ValidateNotEmpty
		case "FQDN":
			input.Validate = tui.ValidateFQDN
		case "Token":
			input.Validate = tui.ValidateNotEmpty
		}

		// Focus first input by default
		if i == 0 {
			input.Focus()
		}

		inputs[i] = input
	}

	return addModel{
		inputs:    inputs,
		focus:     0,
		result:    result,
		force:     force,
		isDefault: isDefault,
		keys:      addKeys,
		help:      help.New(),
	}
}

func (m addModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m addModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, m.keys.Help):
			m.help.ShowAll = !m.help.ShowAll
		case key.Matches(msg, m.keys.Enter):
			if m.focus == len(m.inputs) {
				// Submit - first check if any field has validation errors
				for _, input := range m.inputs {
					if input.Err != nil {
						// Don't proceed if any field has validation errors
						m.err = errors.New("please fix all field errors before submitting")
						return m, nil
					}
				}

				// Also validate in case fields haven't been edited
				if err := m.validateOnSubmit(); err != nil {
					m.err = err
					return m, nil
				}

				m.instance = coolTypes.Instance{
					Name:    strings.TrimSpace(m.inputs[0].Value()),
					Fqdn:    strings.TrimSpace(m.inputs[1].Value()),
					Token:   strings.TrimSpace(m.inputs[2].Value()),
					Default: m.isDefault,
				}
				// Return a command to send the instance
				return m, func() tea.Msg {
					if m.result != nil {
						m.result <- m.instance
					}
					return tea.Quit()
				}
			} else if m.focus == len(m.inputs)+1 {
				// Cancel
				return m, tea.Quit
			}
			// Move to next input
			m.focus++
			m.updateFocus()
		case key.Matches(msg, m.keys.Tab):
			if msg.String() == "tab" {
				m.focus++
			} else {
				m.focus--
			}

			// Wrap around
			if m.focus > len(m.inputs)+1 {
				m.focus = 0
			} else if m.focus < 0 {
				m.focus = len(m.inputs) + 1
			}

			m.updateFocus()
		case key.Matches(msg, m.keys.Up):
			m.focus--
			if m.focus < 0 {
				m.focus = len(m.inputs) + 1
			}
			m.updateFocus()
		case key.Matches(msg, m.keys.Down):
			m.focus++
			if m.focus > len(m.inputs)+1 {
				m.focus = 0
			}
			m.updateFocus()
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.help.Width = msg.Width
	}

	// Handle text input updates
	if m.focus < len(m.inputs) {
		var cmd tea.Cmd
		m.inputs[m.focus], cmd = m.inputs[m.focus].Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m *addModel) updateFocus() {
	// Blur all inputs
	for i := range m.inputs {
		m.inputs[i].Blur()
	}

	// Focus current input if it's a text input
	if m.focus < len(m.inputs) {
		m.inputs[m.focus].Focus()
	}
}

// validateOnSubmit handles validation for fields that haven't been edited
func (m addModel) validateOnSubmit() error {
	// Trigger validation for all fields
	for i, input := range m.inputs {
		// If the field hasn't been edited and is empty, it hasn't triggered validation yet
		switch i {
		case 0:
			return tui.ValidateNotEmpty(input.Value())
		case 1:
			return tui.ValidateFQDN(input.Value())
		case 2:
			return tui.ValidateNotEmpty(input.Value())
		}
	}
	return nil
}

func (m addModel) View() string {
	if m.width == 0 {
		return "loading..."
	}

	var s strings.Builder

	// Title
	s.WriteString("Add New Instance\n\n")

	// Input fields with validation errors
	for _, input := range m.inputs {
		s.WriteString(input.View())
		if input.Err != nil {
			// Display the validation error next to the input
			s.WriteString(" ")
			s.WriteString(ErrorStyle.Render(input.Err.Error()))
		}
		s.WriteString("\n")
	}

	// Submit and Cancel buttons
	submitStyle := BlurredStyle
	if m.focus == len(m.inputs) {
		submitStyle = FocusedStyle
	}
	s.WriteString(submitStyle.Render("Submit"))
	s.WriteString(" ")

	cancelStyle := BlurredStyle
	if m.focus == len(m.inputs)+1 {
		cancelStyle = FocusedStyle
	}
	s.WriteString(cancelStyle.Render("Cancel"))

	// Help view at the bottom
	s.WriteString("\n\n")
	s.WriteString(m.help.View(m.keys))

	// General form error message (if any)
	if m.err != nil {
		s.WriteString("\n\n")
		s.WriteString(ErrorStyle.Render(m.err.Error()))
	}

	return s.String()
}
