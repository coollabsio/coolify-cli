package cliinit

import (
	"errors"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/coollabsio/cli-coolify/cmd/coolTypes"
	"github.com/coollabsio/cli-coolify/pkg/tui"
)

var (
	checkboxStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("99"))
	checked       = checkboxStyle.Render("[x]")
	unchecked     = checkboxStyle.Render("[ ]")
	goldStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("220")).Bold(true)
)

// initKeyMap defines keybindings for the initialization form
type initKeyMap struct {
	Up    key.Binding
	Down  key.Binding
	Tab   key.Binding
	Space key.Binding
	Enter key.Binding
	Paste key.Binding
	Help  key.Binding
	Quit  key.Binding
}

// ShortHelp returns keybindings to be shown in the mini help view
func (k initKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Help, k.Quit}
}

// FullHelp returns keybindings for the expanded help view
func (k initKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Tab},               // first column
		{k.Space, k.Enter, k.Paste, k.Help}, // second column
		{k.Quit},                            // third column
	}
}

var initKeys = initKeyMap{
	Up: key.NewBinding(
		key.WithKeys("up"),
		key.WithHelp("↑", "move up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down"),
		key.WithHelp("↓", "move down"),
	),
	Tab: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "next field"),
	),
	Space: key.NewBinding(
		key.WithKeys(" "),
		key.WithHelp("space", "toggle checkbox"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "continue"),
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

type initModel struct {
	instances     []coolTypes.Instance
	width         int
	height        int
	focus         int
	err           error
	useCloud      bool
	useSelfHost   bool
	cloudToken    textinput.Model
	selfHostName  textinput.Model
	selfHostFqdn  textinput.Model
	selfHostToken textinput.Model
	result        chan<- []coolTypes.Instance
	step          int // Current step in the initialization process
	tick          int // For rainbow effect
	keys          initKeyMap
	help          help.Model
}

// validateCloudToken validates the cloud token
func validateCloudToken(s string) error {
	trimmed := strings.TrimSpace(s)
	if trimmed == "" {
		return errors.New("token is required when using Coolify Cloud")
	}
	return nil
}

// validateInstanceName validates the instance name
func validateInstanceName(s string) error {
	trimmed := strings.TrimSpace(s)
	if trimmed == "" {
		return errors.New("name is required for self-hosted instance")
	}
	return nil
}

// validateInstanceToken validates the instance token
func validateInstanceToken(s string) error {
	trimmed := strings.TrimSpace(s)
	if trimmed == "" {
		return errors.New("token is required for self-hosted instance")
	}
	return nil
}

func newInitModel(result chan<- []coolTypes.Instance) initModel {
	cloudToken := textinput.New()
	cloudToken.Placeholder = "Enter your Coolify Cloud token"
	cloudToken.Prompt = "Cloud Token: "
	cloudToken.PromptStyle = tui.FocusedStyle
	cloudToken.TextStyle = tui.FocusedStyle
	cloudToken.Validate = tui.ValidateNotEmpty

	selfHostName := textinput.New()
	selfHostName.Placeholder = "Enter name for self-hosted instance"
	selfHostName.Prompt = "Name: "
	selfHostName.PromptStyle = tui.FocusedStyle
	selfHostName.TextStyle = tui.FocusedStyle
	selfHostName.Validate = tui.ValidateNotEmpty

	selfHostFqdn := textinput.New()
	selfHostFqdn.Placeholder = "Enter FQDN for self-hosted instance"
	selfHostFqdn.Prompt = "FQDN: "
	selfHostFqdn.PromptStyle = tui.FocusedStyle
	selfHostFqdn.TextStyle = tui.FocusedStyle
	selfHostFqdn.Validate = tui.ValidateFQDN

	selfHostToken := textinput.New()
	selfHostToken.Placeholder = "Enter token for self-hosted instance"
	selfHostToken.Prompt = "Token: "
	selfHostToken.PromptStyle = tui.FocusedStyle
	selfHostToken.TextStyle = tui.FocusedStyle
	selfHostToken.Validate = tui.ValidateNotEmpty

	return initModel{
		instances:     make([]coolTypes.Instance, 0),
		focus:         0,
		result:        result,
		step:          0,
		cloudToken:    cloudToken,
		selfHostName:  selfHostName,
		selfHostFqdn:  selfHostFqdn,
		selfHostToken: selfHostToken,
		keys:          initKeys,
		help:          help.New(),
	}
}

func (m initModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m initModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, m.keys.Help):
			m.help.ShowAll = !m.help.ShowAll
			return m, nil
		case key.Matches(msg, m.keys.Space):
			// Space toggles checkbox when on step 0 or 2
			switch m.step {
			case 0:
				m.useCloud = !m.useCloud
				return m, nil
			case 2:
				m.useSelfHost = !m.useSelfHost
				return m, nil
			}
		case key.Matches(msg, m.keys.Enter):
			if m.step == 0 {
				// Enter handles progression
				if m.useCloud {
					m.step++
					m.focus = 1
					m.cloudToken.Focus()
				} else {
					m.step += 2
					m.focus = 2
				}
			} else if m.step == 1 {
				if m.useCloud {
					// Check for validation errors
					if m.cloudToken.Err != nil {
						m.err = m.cloudToken.Err
						return m, nil
					}
					// Manual validation in case field hasn't been edited
					if m.cloudToken.Value() == "" {
						m.err = errors.New("token is required when using Coolify Cloud")
						return m, nil
					}
					m.step++
					m.focus = 2
					m.cloudToken.Blur()
				}
			} else if m.step == 2 {
				// Enter handles progression
				if m.useSelfHost {
					m.step++
					m.focus = 3
					m.selfHostName.Focus()
				} else {
					// If self-hosted is false, build instances and quit
					if m.useCloud {
						m.instances = append(m.instances, coolTypes.Instance{
							Name:    "cloud",
							Default: true,
							Fqdn:    "https://app.coolify.io",
							Token:   m.cloudToken.Value(),
						})
					}
					// Send instances back to command
					if m.result != nil {
						m.result <- m.instances
					}
					return m, tea.Quit
				}
			} else if m.step == 3 {
				cloudToken := strings.TrimSpace(m.cloudToken.Value())
				if m.useSelfHost {
					// Check for validation errors
					if m.selfHostName.Err != nil || m.selfHostFqdn.Err != nil || m.selfHostToken.Err != nil {
						m.err = errors.New("please fix all field errors before submitting")
						return m, nil
					}

					selfHostName := strings.TrimSpace(m.selfHostName.Value())
					selfHostFqdn := strings.TrimSpace(m.selfHostFqdn.Value())
					selfHostToken := strings.TrimSpace(m.selfHostToken.Value())
					// Manual validation in case fields haven't been edited
					if selfHostName == "" {
						m.err = errors.New("name is required for self-hosted instance")
						return m, nil
					}
					if selfHostFqdn == "" {
						m.err = errors.New("FQDN is required for self-hosted instance")
						return m, nil
					}
					if selfHostToken == "" {
						m.err = errors.New("token is required for self-hosted instance")
						return m, nil
					}

					// Build instances array
					if m.useCloud {
						m.instances = append(m.instances, coolTypes.Instance{
							Name:    "cloud",
							Default: true,
							Fqdn:    "https://app.coolify.io",
							Token:   cloudToken,
						})
					}
					m.instances = append(m.instances, coolTypes.Instance{
						Name:    selfHostName,
						Default: !m.useCloud,
						Fqdn:    selfHostFqdn,
						Token:   selfHostToken,
					})
					// Send instances back to command
					if m.result != nil {
						m.result <- m.instances
					}
					return m, tea.Quit
				} else {
					// If self-hosted is false, build instances and quit
					if m.useCloud {
						m.instances = append(m.instances, coolTypes.Instance{
							Name:    "cloud",
							Default: true,
							Fqdn:    "https://app.coolify.io",
							Token:   cloudToken,
						})
					}
					// Send instances back to command
					if m.result != nil {
						m.result <- m.instances
					}
					return m, tea.Quit
				}
			}
		case key.Matches(msg, m.keys.Up):
			// Only allow up/down navigation when multiple items are visible
			if m.step == 3 && m.useSelfHost {
				m.focus--
				if m.focus < 3 {
					m.focus = 5
				}
				m.updateFocus()
			}
		case key.Matches(msg, m.keys.Down), key.Matches(msg, m.keys.Tab):
			// Only allow up/down navigation when multiple items are visible
			if m.step == 3 && m.useSelfHost {
				m.focus++
				if m.focus > 5 {
					m.focus = 3
				}
				m.updateFocus()
			}
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.help.Width = msg.Width
	}

	// Handle text input updates
	if m.step == 1 && m.focus == 1 {
		m.cloudToken, cmd = m.cloudToken.Update(msg)
		cmds = append(cmds, cmd)
	} else if m.step == 3 {
		switch m.focus {
		case 3:
			m.selfHostName, cmd = m.selfHostName.Update(msg)
			cmds = append(cmds, cmd)
		case 4:
			m.selfHostFqdn, cmd = m.selfHostFqdn.Update(msg)
			cmds = append(cmds, cmd)
		case 5:
			m.selfHostToken, cmd = m.selfHostToken.Update(msg)
			cmds = append(cmds, cmd)
		}
	}

	return m, tea.Batch(cmds...)
}

func (m *initModel) updateFocus() {
	// Blur all inputs
	m.cloudToken.Blur()
	m.selfHostName.Blur()
	m.selfHostFqdn.Blur()
	m.selfHostToken.Blur()

	// Focus the selected input
	switch m.focus {
	case 1:
		m.cloudToken.Focus()
	case 3:
		m.selfHostName.Focus()
	case 4:
		m.selfHostFqdn.Focus()
	case 5:
		m.selfHostToken.Focus()
	}
}

func (m initModel) View() string {
	if m.width == 0 {
		return "loading..."
	}

	var s strings.Builder

	// Title
	s.WriteString("Initialize Coolify CLI\n\n")

	// Step 1: Cloud question
	if m.step == 0 {
		cloudStyle := tui.BlurredStyle
		if m.focus == 0 {
			cloudStyle = tui.FocusedStyle
		}
		s.WriteString(cloudStyle.Render("Do you use "))
		s.WriteString(goldStyle.Render("Coolify Cloud?"))
		s.WriteString(" ")
		if m.useCloud {
			s.WriteString(checked)
		} else {
			s.WriteString(unchecked)
		}
		s.WriteString("\n")
		s.WriteString(tui.BlurredStyle.Render("Hint: use spacebar to toggle checkbox\n"))
	}

	// Step 2: Cloud token input
	if m.step == 1 && m.useCloud {
		s.WriteString(m.cloudToken.View())
		if m.cloudToken.Err != nil {
			// Display validation error next to input
			s.WriteString(" ")
			s.WriteString(tui.ErrorStyle.Render(m.cloudToken.Err.Error()))
		}
		s.WriteString("\n")
	}

	// Step 3: Self-hosted question
	if m.step == 2 {
		selfHostStyle := tui.BlurredStyle
		if m.focus == 2 {
			selfHostStyle = tui.FocusedStyle
		}
		s.WriteString(selfHostStyle.Render("Add self-hosted instance"))
		s.WriteString(" ")
		if m.useSelfHost {
			s.WriteString(checked)
		} else {
			s.WriteString(unchecked)
		}
		s.WriteString("\n")
		s.WriteString(tui.BlurredStyle.Render("Hint: use spacebar to toggle checkbox\n"))
	}

	// Step 4: Self-hosted inputs
	if m.step == 3 && m.useSelfHost {
		// Name input
		s.WriteString(m.selfHostName.View())
		if m.selfHostName.Err != nil {
			// Display validation error next to input
			s.WriteString(" ")
			s.WriteString(tui.ErrorStyle.Render(m.selfHostName.Err.Error()))
		}
		s.WriteString("\n\n")

		// FQDN input
		s.WriteString(m.selfHostFqdn.View())
		if m.selfHostFqdn.Err != nil {
			// Display validation error next to input
			s.WriteString(" ")
			s.WriteString(tui.ErrorStyle.Render(m.selfHostFqdn.Err.Error()))
		}
		s.WriteString("\n\n")

		// Token input
		s.WriteString(m.selfHostToken.View())
		if m.selfHostToken.Err != nil {
			// Display validation error next to input
			s.WriteString(" ")
			s.WriteString(tui.ErrorStyle.Render(m.selfHostToken.Err.Error()))
		}
		s.WriteString("\n")
	}

	// Help view
	s.WriteString("\n\n")
	s.WriteString(m.help.View(m.keys))

	// Error message
	if m.err != nil {
		s.WriteString("\n\n")
		s.WriteString(tui.ErrorStyle.Render(m.err.Error()))
	}

	return s.String()
}
