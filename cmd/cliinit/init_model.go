package cliinit

import (
	"fmt"
	"net"
	"net/url"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/coollabsio/coolify-cli/cmd/coolTypes"
	"github.com/coollabsio/coolify-cli/cmd/utils"
)

var (
	checkboxStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("99"))
	checked       = checkboxStyle.Render("[x]")
	unchecked     = checkboxStyle.Render("[ ]")
	goldStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("220")).Bold(true)
)

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
	showHelp      bool
	result        chan<- []coolTypes.Instance
	step          int // Current step in the initialization process
	tick          int // For rainbow effect
}

func newInitModel(result chan<- []coolTypes.Instance) initModel {
	cloudToken := textinput.New()
	cloudToken.Placeholder = "Enter your Coolify Cloud token"
	cloudToken.Prompt = "Cloud Token: "
	cloudToken.PromptStyle = FocusedStyle
	cloudToken.TextStyle = FocusedStyle

	selfHostName := textinput.New()
	selfHostName.Placeholder = "Enter name for self-hosted instance"
	selfHostName.Prompt = "Name: "
	selfHostName.PromptStyle = FocusedStyle
	selfHostName.TextStyle = FocusedStyle

	selfHostFqdn := textinput.New()
	selfHostFqdn.Placeholder = "Enter FQDN for self-hosted instance"
	selfHostFqdn.Prompt = "FQDN: "
	selfHostFqdn.PromptStyle = FocusedStyle
	selfHostFqdn.TextStyle = FocusedStyle

	selfHostToken := textinput.New()
	selfHostToken.Placeholder = "Enter token for self-hosted instance"
	selfHostToken.Prompt = "Token: "
	selfHostToken.PromptStyle = FocusedStyle
	selfHostToken.TextStyle = FocusedStyle

	return initModel{
		instances:     make([]coolTypes.Instance, 0),
		focus:         0,
		showHelp:      false,
		result:        result,
		step:          0,
		cloudToken:    cloudToken,
		selfHostName:  selfHostName,
		selfHostFqdn:  selfHostFqdn,
		selfHostToken: selfHostToken,
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
		switch msg.String() {
		case "esc":
			return m, tea.Quit
		case "ctrl+c":
			return m, tea.Quit
		case "ctrl+v":
			// Handle paste based on the current focus
			switch m.focus {
			case 1:
				utils.PasteToInput(&m.cloudToken)
			case 3:
				utils.PasteToInput(&m.selfHostName)
			case 4:
				utils.PasteToInput(&m.selfHostFqdn)
			case 5:
				utils.PasteToInput(&m.selfHostToken)
			}
			return m, nil
		case "ctrl+x":
			// Handle cut based on the current focus
			switch m.focus {
			case 1:
				utils.CutFromInput(&m.cloudToken)
			case 3:
				utils.CutFromInput(&m.selfHostName)
			case 4:
				utils.CutFromInput(&m.selfHostFqdn)
			case 5:
				utils.CutFromInput(&m.selfHostToken)
			}
			return m, nil
		case "enter", " ":
			if m.step == 0 {
				if msg.String() == " " {
					// Space only toggles checkbox
					m.useCloud = !m.useCloud
				} else {
					// Enter handles progression
					if m.useCloud {
						m.step++
						m.focus = 1
						m.cloudToken.Focus()
					} else {
						m.step += 2
						m.focus = 2
					}
				}
			} else if m.step == 1 {
				if m.useCloud {
					// Validate cloud token
					if m.cloudToken.Value() == "" {
						m.err = fmt.Errorf("token is required when using Coolify Cloud")
						return m, nil
					}
					m.step++
					m.focus = 2
					m.cloudToken.Blur()
				}
			} else if m.step == 2 {
				if msg.String() == " " {
					// Space only toggles checkbox
					m.useSelfHost = !m.useSelfHost
				} else {
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
				}
			} else if m.step == 3 {
				if m.useSelfHost {
					// Validate self-hosted inputs
					if m.selfHostName.Value() == "" {
						m.err = fmt.Errorf("name is required for self-hosted instance")
						return m, nil
					}
					if m.selfHostFqdn.Value() == "" {
						m.err = fmt.Errorf("FQDN is required for self-hosted instance")
						return m, nil
					}
					if m.selfHostToken.Value() == "" {
						m.err = fmt.Errorf("token is required for self-hosted instance")
						return m, nil
					}

					// Validate FQDN format
					fqdn := m.selfHostFqdn.Value()
					if strings.Contains(fqdn, "://") {
						u, err := url.Parse(fqdn)
						if err != nil {
							m.err = fmt.Errorf("invalid URL format: %s", err)
							return m, nil
						}
						if u.Scheme != "http" && u.Scheme != "https" {
							m.err = fmt.Errorf("URL scheme must be http or https")
							return m, nil
						}
						if u.Host == "" {
							m.err = fmt.Errorf("URL must contain a host")
							return m, nil
						}
					} else {
						host, port, err := net.SplitHostPort(fqdn)
						if err != nil {
							m.err = fmt.Errorf("invalid IP:port format: %s", err)
							return m, nil
						}
						if net.ParseIP(host) == nil {
							m.err = fmt.Errorf("invalid IP address: %s", host)
							return m, nil
						}
						if port == "" {
							m.err = fmt.Errorf("port is required for IP address format")
							return m, nil
						}
					}
					// Build instances array
					if m.useCloud {
						m.instances = append(m.instances, coolTypes.Instance{
							Name:    "cloud",
							Default: true,
							Fqdn:    "https://app.coolify.io",
							Token:   m.cloudToken.Value(),
						})
					}
					m.instances = append(m.instances, coolTypes.Instance{
						Name:    m.selfHostName.Value(),
						Default: !m.useCloud,
						Fqdn:    m.selfHostFqdn.Value(),
						Token:   m.selfHostToken.Value(),
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
							Token:   m.cloudToken.Value(),
						})
					}
					// Send instances back to command
					if m.result != nil {
						m.result <- m.instances
					}
					return m, tea.Quit
				}
			}
		case "up":
			// Only allow up/down navigation when multiple items are visible
			if m.step == 3 && m.useSelfHost {
				m.focus--
				if m.focus < 3 {
					m.focus = 5
				}
				m.updateFocus()
			}
		case "down", "tab":
			// Only allow up/down navigation when multiple items are visible
			if m.step == 3 && m.useSelfHost {
				m.focus++
				if m.focus > 5 {
					m.focus = 3
				}
				m.updateFocus()
			}
		case "?":
			m.showHelp = !m.showHelp
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
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
		cloudStyle := BlurredStyle
		if m.focus == 0 {
			cloudStyle = FocusedStyle
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
	}

	// Step 2: Cloud token input
	if m.step == 1 && m.useCloud {
		s.WriteString(m.cloudToken.View())
		s.WriteString("\n")
	}

	// Step 3: Self-hosted question
	if m.step == 2 {
		selfHostStyle := BlurredStyle
		if m.focus == 2 {
			selfHostStyle = FocusedStyle
		}
		s.WriteString(selfHostStyle.Render("Add self-hosted instance"))
		s.WriteString(" ")
		if m.useSelfHost {
			s.WriteString(checked)
		} else {
			s.WriteString(unchecked)
		}
		s.WriteString("\n")
	}

	// Step 4: Self-hosted inputs
	if m.step == 3 && m.useSelfHost {
		// Name input
		s.WriteString(m.selfHostName.View())
		s.WriteString("\n\n")

		// FQDN input
		s.WriteString(m.selfHostFqdn.View())
		s.WriteString("\n\n")

		// Token input
		s.WriteString(m.selfHostToken.View())
		s.WriteString("\n")
	}

	// Help text
	if m.showHelp {
		s.WriteString("\n\n")
		s.WriteString(BlurredStyle.Render("Help:"))
		s.WriteString("\n")
		s.WriteString(BlurredStyle.Render("• Use arrow keys to navigate"))
		s.WriteString("\n")
		s.WriteString(BlurredStyle.Render("• Space to toggle checkbox"))
		s.WriteString("\n")
		s.WriteString(BlurredStyle.Render("• Enter to continue"))
		s.WriteString("\n")
		s.WriteString(BlurredStyle.Render("• Tab to move between fields"))
		s.WriteString("\n")
		s.WriteString(BlurredStyle.Render("• Ctrl+V to paste from clipboard"))
		s.WriteString("\n")
		s.WriteString(BlurredStyle.Render("• Ctrl+X to cut to clipboard"))
		s.WriteString("\n")
		s.WriteString(BlurredStyle.Render("• Ctrl+C to exit"))
		s.WriteString("\n")
		s.WriteString(BlurredStyle.Render("• ? to toggle help"))
	}

	// Error message
	if m.err != nil {
		s.WriteString("\n\n")
		s.WriteString(ErrorStyle.Render(m.err.Error()))
	}

	return s.String()
}
