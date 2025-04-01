package cliinstances

import (
	"fmt"
	"net"
	"net/url"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/coollabsio/coolify-cli/cmd/coolTypes"
	"github.com/coollabsio/coolify-cli/cmd/utils"
)

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
	}
}

func (m addModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m addModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return m, tea.Quit
		case "ctrl+c":
			return m, tea.Quit
		case "ctrl+v":
			if m.focus < len(m.inputs) {
				utils.PasteToInput(&m.inputs[m.focus])
			}
			return m, nil
		case "ctrl+x":
			if m.focus < len(m.inputs) {
				utils.CutFromInput(&m.inputs[m.focus])
			}
			return m, nil
		case "enter":
			if m.focus == len(m.inputs) {
				// Submit
				if err := m.validate(); err != nil {
					m.err = err
					return m, nil
				}
				m.instance = coolTypes.Instance{
					Name:    m.inputs[0].Value(),
					Fqdn:    m.inputs[1].Value(),
					Token:   m.inputs[2].Value(),
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
		case "tab", "shift+tab":
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
		case "up":
			m.focus--
			if m.focus < 0 {
				m.focus = len(m.inputs) + 1
			}
			m.updateFocus()
		case "down":
			m.focus++
			if m.focus > len(m.inputs)+1 {
				m.focus = 0
			}
			m.updateFocus()
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
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

func (m addModel) validate() error {
	if m.inputs[0].Value() == "" {
		return fmt.Errorf("name is required")
	}
	if m.inputs[1].Value() == "" {
		return fmt.Errorf("FQDN is required")
	}
	if m.inputs[2].Value() == "" {
		return fmt.Errorf("token is required")
	}

	// Validate FQDN format
	fqdn := m.inputs[1].Value()
	if strings.Contains(fqdn, "://") {
		// Check if it's a valid HTTP(S) URL
		u, err := url.Parse(fqdn)
		if err != nil {
			return fmt.Errorf("invalid URL format: %s", err)
		}
		if u.Scheme != "http" && u.Scheme != "https" {
			return fmt.Errorf("URL scheme must be http or https")
		}
		if u.Host == "" {
			return fmt.Errorf("URL must contain a host")
		}
	} else {
		// Check if it's a valid IP:port format
		host, port, err := net.SplitHostPort(fqdn)
		if err != nil {
			return fmt.Errorf("invalid IP:port format: %s", err)
		}
		if net.ParseIP(host) == nil {
			return fmt.Errorf("invalid IP address: %s", host)
		}
		if port == "" {
			return fmt.Errorf("port is required for IP address format")
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

	// Input fields
	for i := range m.inputs {
		s.WriteString(m.inputs[i].View())
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

	// Keyboard shortcuts help
	s.WriteString("\n\n")
	s.WriteString(BlurredStyle.Render("Shortcuts:"))
	s.WriteString("\n")
	s.WriteString(BlurredStyle.Render("• Tab/Arrows to navigate"))
	s.WriteString("\n")
	s.WriteString(BlurredStyle.Render("• Enter to submit/select"))
	s.WriteString("\n")
	s.WriteString(BlurredStyle.Render("• Ctrl+V to paste from clipboard"))
	s.WriteString("\n")
	s.WriteString(BlurredStyle.Render("• Ctrl+X to cut to clipboard"))
	s.WriteString("\n")
	s.WriteString(BlurredStyle.Render("• Ctrl+C to exit"))

	// Error message
	if m.err != nil {
		s.WriteString("\n\n")
		s.WriteString(ErrorStyle.Render(m.err.Error()))
	}

	return s.String()
}
