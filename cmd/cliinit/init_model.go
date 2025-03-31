package cliinit

import (
	"fmt"
	"net"
	"net/url"
	"strings"

	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/coollabsio/coolify-cli/cmd/coolTypes"
)

type initModel struct {
	instances     []coolTypes.Instance
	width         int
	height        int
	focus         int
	err           error
	useCloud      bool
	cloudToken    string
	useSelfHost   bool
	selfHostFqdn  string
	selfHostName  string
	selfHostToken string
	cursor        int
	showHelp      bool
	result        chan<- []coolTypes.Instance
	step          int // Current step in the initialization process
	tick          int // For rainbow effect
}

var (
	checkboxStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("99"))
	checked       = checkboxStyle.Render("[x]")
	unchecked     = checkboxStyle.Render("[ ]")
	goldStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("220")).Bold(true)
)

func newInitModel(result chan<- []coolTypes.Instance) initModel {
	return initModel{
		instances: make([]coolTypes.Instance, 0),
		focus:     0,
		cursor:    0,
		showHelp:  false,
		result:    result,
		step:      0,
	}
}

func (m initModel) Init() tea.Cmd {
	return nil
}

func (m initModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return m, tea.Quit
		case "ctrl+c":
			return m, tea.Quit
		case "ctrl+v":
			text, err := clipboard.ReadAll()
			if err == nil {
				switch m.focus {
				case 1:
					m.cloudToken = text
				case 3:
					m.selfHostName = text
				case 4:
					m.selfHostFqdn = text
				case 5:
					m.selfHostToken = text
				}
				m.cursor = len(text)
			}
		case "ctrl+x":
			switch m.focus {
			case 1:
				clipboard.WriteAll(m.cloudToken)
			case 3:
				clipboard.WriteAll(m.selfHostName)
			case 4:
				clipboard.WriteAll(m.selfHostFqdn)
			case 5:
				clipboard.WriteAll(m.selfHostToken)
			}
			m.cursor = 0
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
						m.cursor = 0
					} else {
						m.step += 2
						m.focus = 2
					}
				}
			} else if m.step == 1 {
				if m.useCloud {
					// Validate cloud token
					if m.cloudToken == "" {
						m.err = fmt.Errorf("token is required when using Coolify Cloud")
						return m, nil
					}
					m.step++
					m.focus = 2
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
						m.cursor = 0
					} else {
						// If self-hosted is false, build instances and quit
						if m.useCloud {
							m.instances = append(m.instances, coolTypes.Instance{
								Name:    "cloud",
								Default: true,
								Fqdn:    "https://app.coolify.io",
								Token:   m.cloudToken,
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
					if m.selfHostName == "" {
						m.err = fmt.Errorf("name is required for self-hosted instance")
						return m, nil
					}
					if m.selfHostFqdn == "" {
						m.err = fmt.Errorf("FQDN is required for self-hosted instance")
						return m, nil
					}
					if m.selfHostToken == "" {
						m.err = fmt.Errorf("token is required for self-hosted instance")
						return m, nil
					}

					// Validate FQDN format
					if strings.Contains(m.selfHostFqdn, "://") {
						u, err := url.Parse(m.selfHostFqdn)
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
						host, port, err := net.SplitHostPort(m.selfHostFqdn)
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
							Token:   m.cloudToken,
						})
					}
					m.instances = append(m.instances, coolTypes.Instance{
						Name:    m.selfHostName,
						Default: !m.useCloud,
						Fqdn:    m.selfHostFqdn,
						Token:   m.selfHostToken,
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
							Token:   m.cloudToken,
						})
					}
					// Send instances back to command
					if m.result != nil {
						m.result <- m.instances
					}
					return m, tea.Quit
				}
			}
		case "backspace":
			if m.focus == 1 || m.focus == 3 || m.focus == 4 || m.focus == 5 {
				switch m.focus {
				case 1:
					if m.cursor > 0 {
						m.cloudToken = m.cloudToken[:m.cursor-1] + m.cloudToken[m.cursor:]
						m.cursor--
					}
				case 3:
					if m.cursor > 0 {
						m.selfHostName = m.selfHostName[:m.cursor-1] + m.selfHostName[m.cursor:]
						m.cursor--
					}
				case 4:
					if m.cursor > 0 {
						m.selfHostFqdn = m.selfHostFqdn[:m.cursor-1] + m.selfHostFqdn[m.cursor:]
						m.cursor--
					}
				case 5:
					if m.cursor > 0 {
						m.selfHostToken = m.selfHostToken[:m.cursor-1] + m.selfHostToken[m.cursor:]
						m.cursor--
					}
				}
			}
		case "left":
			if m.focus == 1 || m.focus == 3 || m.focus == 4 || m.focus == 5 {
				switch m.focus {
				case 1:
					if m.cursor > 0 {
						m.cursor--
					}
				case 3:
					if m.cursor > 0 {
						m.cursor--
					}
				case 4:
					if m.cursor > 0 {
						m.cursor--
					}
				case 5:
					if m.cursor > 0 {
						m.cursor--
					}
				}
			}
		case "right":
			if m.focus == 1 || m.focus == 3 || m.focus == 4 || m.focus == 5 {
				switch m.focus {
				case 1:
					if m.cursor < len(m.cloudToken) {
						m.cursor++
					}
				case 3:
					if m.cursor < len(m.selfHostName) {
						m.cursor++
					}
				case 4:
					if m.cursor < len(m.selfHostFqdn) {
						m.cursor++
					}
				case 5:
					if m.cursor < len(m.selfHostToken) {
						m.cursor++
					}
				}
			}
		case "up":
			// Only allow up/down navigation when multiple items are visible
			if m.step == 3 && m.useSelfHost {
				m.focus--
				if m.focus < 0 {
					m.focus = 5
				}
				if m.focus == 3 || m.focus == 4 || m.focus == 5 {
					m.cursor = 0
				}
			}
		case "down":
			// Only allow up/down navigation when multiple items are visible
			if m.step == 3 && m.useSelfHost {
				m.focus++
				if m.focus > 5 {
					m.focus = 0
				}
				if m.focus == 3 || m.focus == 4 || m.focus == 5 {
					m.cursor = 0
				}
			}
		case "?":
			m.showHelp = !m.showHelp
		default:
			if m.focus >= 1 && m.focus <= 5 && len(msg.String()) == 1 {
				switch m.focus {
				case 1:
					m.cloudToken = m.cloudToken[:m.cursor] + msg.String() + m.cloudToken[m.cursor:]
					m.cursor++
				case 3:
					m.selfHostName = m.selfHostName[:m.cursor] + msg.String() + m.selfHostName[m.cursor:]
					m.cursor++
				case 4:
					m.selfHostFqdn = m.selfHostFqdn[:m.cursor] + msg.String() + m.selfHostFqdn[m.cursor:]
					m.cursor++
				case 5:
					m.selfHostToken = m.selfHostToken[:m.cursor] + msg.String() + m.selfHostToken[m.cursor:]
					m.cursor++
				}
			}
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}
	return m, nil
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
		tokenStyle := BlurredStyle
		if m.focus == 1 {
			tokenStyle = FocusedStyle
		}
		tokenValue := m.cloudToken
		if m.focus == 1 {
			if tokenValue == "" {
				tokenValue = CursorStyle.Render("█")
			} else {
				tokenValue = tokenValue[:m.cursor] + CursorStyle.Render("█") + tokenValue[m.cursor:]
			}
		}
		s.WriteString(tokenStyle.Render(fmt.Sprintf("Cloud Token: %s", tokenValue)))
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
		nameStyle := BlurredStyle
		if m.focus == 3 {
			nameStyle = FocusedStyle
		}
		nameValue := m.selfHostName
		if m.focus == 3 {
			if nameValue == "" {
				nameValue = CursorStyle.Render("█")
			} else {
				nameValue = nameValue[:m.cursor] + CursorStyle.Render("█") + nameValue[m.cursor:]
			}
		}
		s.WriteString(nameStyle.Render(fmt.Sprintf("Name: %s", nameValue)))
		s.WriteString("\n\n")

		// FQDN input
		fqdnStyle := BlurredStyle
		if m.focus == 4 {
			fqdnStyle = FocusedStyle
		}
		fqdnValue := m.selfHostFqdn
		if m.focus == 4 {
			if fqdnValue == "" {
				fqdnValue = CursorStyle.Render("█")
			} else {
				fqdnValue = fqdnValue[:m.cursor] + CursorStyle.Render("█") + fqdnValue[m.cursor:]
			}
		}
		s.WriteString(fqdnStyle.Render(fmt.Sprintf("FQDN: %s", fqdnValue)))
		s.WriteString("\n\n")

		// Token input
		tokenStyle := BlurredStyle
		if m.focus == 5 {
			tokenStyle = FocusedStyle
		}
		tokenValue := m.selfHostToken
		if m.focus == 5 {
			if tokenValue == "" {
				tokenValue = CursorStyle.Render("█")
			} else {
				tokenValue = tokenValue[:m.cursor] + CursorStyle.Render("█") + tokenValue[m.cursor:]
			}
		}
		s.WriteString(tokenStyle.Render(fmt.Sprintf("Token: %s", tokenValue)))
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
		s.WriteString(BlurredStyle.Render("• Ctrl+V to paste"))
		s.WriteString("\n")
		s.WriteString(BlurredStyle.Render("• Ctrl+X to cut"))
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
