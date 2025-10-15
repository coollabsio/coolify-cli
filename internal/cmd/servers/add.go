package servers

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/coollabsio/cli-coolify/internal/client"
	"github.com/coollabsio/cli-coolify/internal/tui"
	"github.com/coollabsio/cli-coolify/internal/utils"
	"github.com/spf13/cobra"
)

// addKeyMap defines keybindings for the add server form
type addKeyMap struct {
	Up    key.Binding
	Down  key.Binding
	Tab   key.Binding
	Enter key.Binding
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
		{k.Up, k.Down, k.Tab}, // first column
		{k.Enter, k.Help},     // second column
		{k.Quit},              // third column
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
	inputs     []textinput.Model
	focusIndex int
	err        error
	done       bool
	keys       addKeyMap
	help       help.Model
}

func (c *cliServers) newAddCommand() *cobra.Command {
	var validate bool

	cmd := &cobra.Command{
		Use:   "add [name] [ip] [private_key_uuid]",
		Short: "Add a new server",
		Long: `
Add a new server to your Coolify instance.
If no arguments are provided, an interactive form will be shown.`,
		SilenceUsage: true,
		Example: utils.GetCommandExample(`
%[1]s servers add "My Server" 192.168.1.100 abcd1234-uuid
%[1]s servers add "Production" 10.0.0.1 efgh5678-uuid --validate
%[1]s servers add  # Interactive mode`),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return c.runInteractiveAdd(validate)
			}

			if len(args) != 3 {
				return fmt.Errorf("requires exactly 3 arguments (name, ip, private_key_uuid) or no arguments for interactive mode")
			}

			return c.addServer(args[0], args[1], args[2], 22, "root", validate)
		},
	}

	cmd.Flags().BoolVar(&validate, "validate", false, "Validate the server after adding")

	return cmd
}

func (c *cliServers) runInteractiveAdd(validate bool) error {
	p := tea.NewProgram(initialAddModel())
	m, err := p.Run()
	if err != nil {
		return fmt.Errorf("error running form: %w", err)
	}

	finalModel := m.(addModel)
	if !finalModel.done {
		return fmt.Errorf("operation cancelled")
	}

	// Get values from the form
	name := strings.TrimSpace(finalModel.inputs[0].Value())
	ip := strings.TrimSpace(finalModel.inputs[1].Value())
	port := strings.TrimSpace(finalModel.inputs[2].Value())
	user := strings.TrimSpace(finalModel.inputs[3].Value())
	privateKeyUUID := strings.TrimSpace(finalModel.inputs[4].Value())

	// Convert port to int with default 22
	portNum := 22
	if port != "" {
		if _, err := fmt.Sscanf(port, "%d", &portNum); err != nil {
			return fmt.Errorf("invalid port number: %s", port)
		}
	}

	// Use default user if not specified
	if user == "" {
		user = "root"
	}

	return c.addServer(name, ip, privateKeyUUID, portNum, user, validate)
}

func initialAddModel() addModel {
	inputs := make([]textinput.Model, 5)

	// Initialize text inputs
	labels := []string{"Name", "IP Address", "Port (default: 22)", "User (default: root)", "Private Key UUID"}
	for i := range inputs {
		input := tui.NewBlurredInput(labels[i], "")
		inputs[i] = input
	}

	inputs[0].Focus()
	return addModel{
		inputs: inputs,
		err:    nil,
		keys:   addKeys,
		help:   help.New(),
	}
}

func (m addModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m addModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	if msg, ok := msg.(tea.KeyMsg); ok {
		if key.Matches(msg, m.keys.Quit) {
			m.done = false
			return m, tea.Quit
		}

		if key.Matches(msg, m.keys.Help) {
			m.help.ShowAll = !m.help.ShowAll
		}

		if key.Matches(msg, m.keys.Enter) {
			// Submit on enter when last input is focused
			if m.focusIndex == len(m.inputs)-1 {
				m.done = true
				return m, tea.Quit
			}
			// Otherwise move to next input
			m.focusIndex++
			if m.focusIndex >= len(m.inputs) {
				m.focusIndex = 0
			}
			m.updateFocus()
		}

		if key.Matches(msg, m.keys.Tab) {
			// Cycle focus between inputs
			if msg.String() == "shift+tab" {
				m.focusIndex--
			} else {
				m.focusIndex++
			}

			if m.focusIndex >= len(m.inputs) {
				m.focusIndex = 0
			} else if m.focusIndex < 0 {
				m.focusIndex = len(m.inputs) - 1
			}
			m.updateFocus()
		}

		if key.Matches(msg, m.keys.Up) {
			m.focusIndex--
			if m.focusIndex < 0 {
				m.focusIndex = len(m.inputs) - 1
			}
			m.updateFocus()
		}

		if key.Matches(msg, m.keys.Down) {
			m.focusIndex++
			if m.focusIndex >= len(m.inputs) {
				m.focusIndex = 0
			}
			m.updateFocus()
		}
	}

	// Handle character input
	cmd := m.updateInputs(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m *addModel) updateFocus() {
	for i := 0; i < len(m.inputs); i++ {
		if i == m.focusIndex {
			m.inputs[i].Focus()
		} else {
			m.inputs[i].Blur()
		}
	}
}

func (m *addModel) updateInputs(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd

	m.inputs[m.focusIndex], cmd = m.inputs[m.focusIndex].Update(msg)

	return cmd
}

func (m addModel) View() string {
	var b strings.Builder

	b.WriteString("Please enter server details:\n\n")

	for i, input := range m.inputs {
		b.WriteString(input.View())
		if i < len(m.inputs)-1 {
			b.WriteString("\n")
		}
	}

	button := "\n\n"
	if m.focusIndex == len(m.inputs)-1 {
		button += lipgloss.NewStyle().
			Foreground(lipgloss.Color("99")).
			Render("[ Submit ]")
	} else {
		button += "[ Submit ]"
	}

	b.WriteString(button)

	// Add help view
	if m.help.ShowAll {
		b.WriteString("\n\n")
		b.WriteString(m.help.View(m.keys))
	} else {
		b.WriteString("\n\n")
		b.WriteString(m.help.ShortHelpView(m.keys.ShortHelp()))
	}

	return b.String()
}

func (c *cliServers) addServer(name, ip, privateKeyUUID string, port int, user string, validate bool) error {
	req, err := c.coolify().Client.CreateServer(context.Background(), client.CreateServerJSONRequestBody{
		Name:            &name,
		Ip:              &ip,
		Port:            &port,
		User:            &user,
		PrivateKeyUuid:  &privateKeyUUID,
		InstantValidate: &validate,
	})
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	parsedResponse, err := client.ParseCreateServerResponse(req)
	if err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if parsedResponse.StatusCode() != http.StatusCreated {
		return fmt.Errorf("failed to add server: %s", *parsedResponse.JSON400.Message)
	}

	if validate {
		fmt.Printf("Server added successfully with uuid %s\n", *parsedResponse.JSON201.Uuid)
	} else {
		fmt.Printf("Server added successfully with uuid %s. Server is not validated. Use 'servers validate %s' to validate the server.\n", *parsedResponse.JSON201.Uuid, *parsedResponse.JSON201.Uuid)
	}

	return nil
}
