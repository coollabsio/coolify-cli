package cliinstances

import (
	"fmt"
	"net"
	"net/url"
	"strings"

	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/coollabsio/coolify-cli/cmd/coolTypes"
)

type addModel struct {
	inputs    []string
	values    []string
	focus     int
	err       error
	instance  coolTypes.Instance
	width     int
	height    int
	result    chan<- coolTypes.Instance
	force     bool
	isDefault bool
	cursor    int
}

// Add a new command type for sending the instance
type sendInstanceMsg struct {
	instance coolTypes.Instance
}

func newAddModel(result chan<- coolTypes.Instance, force, isDefault bool) addModel {
	return addModel{
		inputs: []string{
			"Name",
			"FQDN",
			"Token",
		},
		values:    make([]string, 3),
		focus:     0,
		result:    result,
		force:     force,
		isDefault: isDefault,
		cursor:    0,
	}
}

func (m addModel) Init() tea.Cmd {
	return nil
}

func (m addModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return m, tea.Quit
		case "ctrl+v":
			if m.focus < len(m.inputs) {
				text, err := clipboard.ReadAll()
				if err == nil {
					// Insert clipboard content at cursor position
					if m.values[m.focus] == "" {
						m.values[m.focus] = text
					} else {
						m.values[m.focus] = m.values[m.focus][:m.cursor] + text + m.values[m.focus][m.cursor:]
					}
					m.cursor += len(text)
				}
			}
		case "ctrl+c":
			return m, tea.Quit
		case "ctrl+x":
			if m.focus < len(m.inputs) {
				// Cut the current line
				clipboard.WriteAll(m.values[m.focus])
				m.values[m.focus] = ""
				m.cursor = 0
			}
		case "enter":
			if m.focus == len(m.inputs) {
				// Submit
				if err := m.validate(); err != nil {
					m.err = err
					return m, nil
				}
				m.instance = coolTypes.Instance{
					Name:    m.values[0],
					Fqdn:    m.values[1],
					Token:   m.values[2],
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
			if m.focus < len(m.inputs) {
				m.cursor = 0
			}
			if m.focus > len(m.inputs)+1 {
				m.focus = 0
			}
		case "backspace":
			if m.focus < len(m.inputs) {
				if m.cursor > 0 {
					if m.values[m.focus] == "" {
						return m, nil
					}
					m.values[m.focus] = m.values[m.focus][:m.cursor-1] + m.values[m.focus][m.cursor:]
					m.cursor--
				}
			}
		case "left":
			if m.focus < len(m.inputs) {
				if m.cursor > 0 {
					m.cursor--
				}
			}
		case "right":
			if m.focus < len(m.inputs) {
				if m.cursor < len(m.values[m.focus]) {
					m.cursor++
				}
			}
		case "tab", "shift+tab":
			if msg.String() == "tab" {
				m.focus++
			} else {
				m.focus--
			}
			if m.focus > len(m.inputs)+1 {
				m.focus = 0
			} else if m.focus < 0 {
				m.focus = len(m.inputs) + 1
			}
			if m.focus < len(m.inputs) {
				m.cursor = 0
			}
		case "up":
			m.focus--
			if m.focus < 0 {
				m.focus = len(m.inputs) + 1
			}
			if m.focus < len(m.inputs) {
				m.cursor = 0
			}
		case "down":
			m.focus++
			if m.focus > len(m.inputs)+1 {
				m.focus = 0
			}
			if m.focus < len(m.inputs) {
				m.cursor = 0
			}
		default:
			if m.focus < len(m.inputs) && len(msg.String()) == 1 {
				if m.values[m.focus] == "" {
					m.values[m.focus] = msg.String()
				} else {
					m.values[m.focus] = m.values[m.focus][:m.cursor] + msg.String() + m.values[m.focus][m.cursor:]
				}
				m.cursor++
			}
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	return m, nil
}

func (m addModel) validate() error {
	if m.values[0] == "" {
		return fmt.Errorf("name is required")
	}
	if m.values[1] == "" {
		return fmt.Errorf("FQDN is required")
	}
	if m.values[2] == "" {
		return fmt.Errorf("token is required")
	}

	// Validate FQDN format
	fqdn := m.values[1]
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
		style := BlurredStyle
		if m.focus == i {
			style = FocusedStyle
		}
		value := m.values[i]
		if m.focus == i {
			if value == "" {
				value = CursorStyle.Render("█")
			} else {
				value = value[:m.cursor] + CursorStyle.Render("█") + value[m.cursor:]
			}
		}
		s.WriteString(style.Render(fmt.Sprintf("%s: %s", m.inputs[i], value)))
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

	// Error message
	if m.err != nil {
		s.WriteString("\n\n")
		s.WriteString(ErrorStyle.Render(m.err.Error()))
	}

	return s.String()
}
