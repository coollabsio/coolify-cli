package privatekeys

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/coollabsio/cli-coolify/internal/client"
	"github.com/coollabsio/cli-coolify/internal/config"
	"github.com/coollabsio/cli-coolify/internal/tui"
	"github.com/spf13/cobra"
)

func buildView(item client.PrivateKey, sensitive bool) string {
	var s strings.Builder
	addSection := func(title string, value interface{}) {
		s.WriteString(tui.FocusedStyle.Bold(true).Render(title + ": "))
		if value != nil {
			switch v := value.(type) {
			case *string:
				if v != nil {
					s.WriteString(*v + "\n\n")
				}
			case *bool:
				if v != nil {
					s.WriteString(fmt.Sprintf("%v\n\n", *v))
				}
			case *int:
				if v != nil {
					s.WriteString(fmt.Sprintf("%d\n\n", *v))
				}
			}

		} else {
			s.WriteString("N/A\n\n")
		}
	}
	addSection("UUID", item.Uuid)
	addSection("Name", item.Name)
	addSection("Description", item.Description)
	addSection("Fingerprint", item.Fingerprint)

	if sensitive {
		addSection("Private Key", item.PrivateKey)
		addSection("Public Key", item.PublicKey)
	} else {
		addSection("Private Key", &config.Redacted)
		addSection("Public Key", &config.Redacted)
	}

	addSection("Git Related", item.IsGitRelated)
	addSection("Team ID", item.TeamId)
	addSection("Created At", item.CreatedAt)
	addSection("Updated At", item.UpdatedAt)

	return s.String()
}

type keyMap struct {
	Up            key.Binding
	Down          key.Binding
	PageUp        key.Binding
	PageDown      key.Binding
	Quit          key.Binding
	ShowSensitive key.Binding
}

func defaultKeyMap() keyMap {
	return keyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "move up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "move down"),
		),
		PageUp: key.NewBinding(
			key.WithKeys("pgup"),
			key.WithHelp("pgup", "page up"),
		),
		PageDown: key.NewBinding(
			key.WithKeys("pgdown"),
			key.WithHelp("pgdown", "page down"),
		),
		Quit: key.NewBinding(
			key.WithKeys("esc", "ctrl+c"),
			key.WithHelp("esc", "quit"),
		),
		ShowSensitive: key.NewBinding(
			key.WithKeys("ctrl+s"),
			key.WithHelp("ctrl+s", "show sensitive"),
		),
	}
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down},
		{k.PageUp, k.PageDown},
		{k.Quit},
		{k.ShowSensitive},
	}
}

type privateKeyModel struct {
	viewport   viewport.Model
	keymap     keyMap
	help       help.Model
	ready      bool
	privateKey client.PrivateKey
	sensitive  bool
	quitting   bool
	err        error
}

func newPrivateKeyModel(privateKey client.PrivateKey, sensitive bool) privateKeyModel {
	return privateKeyModel{
		keymap:     defaultKeyMap(),
		help:       help.New(),
		privateKey: privateKey,
		sensitive:  sensitive,
	}
}

func (m privateKeyModel) Init() tea.Cmd {
	return nil
}

func (m privateKeyModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keymap.Quit):
			m.quitting = true
			return m, tea.Quit
		case key.Matches(msg, m.keymap.Up):
			m.viewport.LineUp(1)
		case key.Matches(msg, m.keymap.Down):
			m.viewport.LineDown(1)
		case key.Matches(msg, m.keymap.PageUp):
			m.viewport.HalfViewUp()
		case key.Matches(msg, m.keymap.PageDown):
			m.viewport.HalfViewDown()
		case key.Matches(msg, m.keymap.ShowSensitive):
			m.sensitive = !m.sensitive
			m.viewport.SetContent(buildView(m.privateKey, m.sensitive))
		}

	case tea.WindowSizeMsg:
		if !m.ready {
			m.viewport = viewport.New(msg.Width, msg.Height-4)
			m.viewport.Style = lipgloss.NewStyle().
				BorderStyle(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("62")).
				Padding(0, 2)
			m.viewport.SetContent(buildView(m.privateKey, m.sensitive))
			m.help.Width = msg.Width
			m.ready = true
		} else {
			m.viewport.Width = msg.Width
			m.viewport.Height = msg.Height - 4
			m.help.Width = msg.Width
		}
	}

	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m privateKeyModel) View() string {
	if !m.ready {
		return "Initializing..."
	}
	if m.err != nil {
		return fmt.Sprintf("Error: %v\nPress esc to quit", m.err)
	}

	var s strings.Builder
	s.WriteString(m.viewport.View())
	s.WriteString("\n\n")
	s.WriteString(m.help.View(m.keymap))
	return s.String()
}

func (c *cliPrivateKeys) newGetCommand() *cobra.Command {
	var showSensitive bool

	cmd := &cobra.Command{
		Use:   "get [uuid]",
		Short: "Get private key details",
		Long:  `Get the details of a specific private key by its UUID.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			uuid := args[0]

			response, err := c.coolify().Client.GetPrivateKeyByUuid(cmd.Context(), uuid)
			if err != nil {
				return fmt.Errorf("failed to create request: %w", err)
			}

			parsedResponse, err := client.ParseGetPrivateKeyByUuidResponse(response)
			if err != nil {
				return fmt.Errorf("failed to parse response: %w", err)
			}
			if parsedResponse.StatusCode() != http.StatusOK {
				return fmt.Errorf("failed to fetch private key: %s", string(parsedResponse.Body))
			}

			key := *parsedResponse.JSON200

			format, err := cmd.Flags().GetString("format")
			if err != nil {
				return fmt.Errorf("failed to get format: %w", err)
			}
			if format == "json" {
				// Redact sensitive data if --show-sensitive is not set
				if !showSensitive {
					// Create a copy with redacted sensitive fields
					redactedKey := key
					redactedKey.PrivateKey = &config.Redacted
					redactedKey.PublicKey = &config.Redacted
					key = redactedKey
				}

				// For JSON output, directly encode to stdout
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				return encoder.Encode(key)
			}

			// Initialize and run Bubble Tea program
			m := newPrivateKeyModel(key, showSensitive)
			p := tea.NewProgram(m)
			if _, err := p.Run(); err != nil {
				return fmt.Errorf("error running program: %w", err)
			}

			return nil
		},
	}

	// Add flags
	flags := cmd.Flags()
	flags.BoolVarP(&showSensitive, "show-sensitive", "s", false, "Show sensitive information like key contents")

	return cmd
}
