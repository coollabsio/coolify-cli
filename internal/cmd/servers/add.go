package servers

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
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

var (
	checkboxStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("99"))
	checked       = checkboxStyle.Render("[x]")
	unchecked     = checkboxStyle.Render("[ ]")
)

// addKeyMap defines keybindings for the add server form
type addKeyMap struct {
	Up    key.Binding
	Down  key.Binding
	Tab   key.Binding
	Enter key.Binding
	Space key.Binding
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
		{k.Enter, k.Space, k.Help}, // second column
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
		key.WithHelp("enter", "submit/next"),
	),
	Space: key.NewBinding(
		key.WithKeys(" "),
		key.WithHelp("space", "toggle"),
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
	inputs           []textinput.Model
	focusIndex       int
	err              error
	done             bool
	validateCheckbox bool
	keys             addKeyMap
	help             help.Model
}

func (c *cliServers) newAddCommand() *cobra.Command {
	var validate bool
	var port int
	var user string

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
				return c.runInteractiveAdd(port, user, validate)
			}

			if len(args) != 3 {
				return fmt.Errorf("requires exactly 3 arguments (name, ip, private_key_uuid) or no arguments for interactive mode")
			}

			return c.addServer(args[0], args[1], args[2], port, user, validate)
		},
	}

	cmd.Flags().BoolVar(&validate, "validate", false, "Validate the server after adding")
	cmd.Flags().IntVar(&port, "port", 22, "SSH port (default: 22)")
	cmd.Flags().StringVar(&user, "user", "root", "SSH user (default: root)")

	return cmd
}

func (c *cliServers) runInteractiveAdd(port int, user string, validate bool) error {
	p := tea.NewProgram(initialAddModel(port, user, validate))
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
	portString := strings.TrimSpace(finalModel.inputs[2].Value())
	userValue := strings.TrimSpace(finalModel.inputs[3].Value())
	privateKeyUUID := strings.TrimSpace(finalModel.inputs[4].Value())
	validateServer := finalModel.validateCheckbox

	// Convert port to int with default 22
	portNum := 22
	if portString != "" {
		portNum, err = strconv.Atoi(portString)
		if err != nil {
			return fmt.Errorf("invalid port number: %s", portString)
		}
	}

	// Use default user if not specified
	if userValue == "" {
		userValue = "root"
	}

	return c.addServer(name, ip, privateKeyUUID, portNum, userValue, validateServer)
}

func initialAddModel(port int, user string, validate bool) addModel {
	inputs := make([]textinput.Model, 5)

	// Initialize text inputs with autofill support
	labels := []string{"Name", "IP Address", "Port", "User", "Private Key UUID"}

	// Set default values from flags
	portStr := ""
	if port != 0 && port != 22 {
		portStr = fmt.Sprintf("%d", port)
	}

	userStr := ""
	if user != "" && user != "root" {
		userStr = user
	}

	values := []string{"", "", portStr, userStr, ""}
	placeholders := []string{"My Server", "192.168.1.100", "22", "root", "uuid-1234-5678"}

	for i := range inputs {
		inputs[i] = tui.NewBlurredInput(placeholders[i], labels[i]+": ")
		inputs[i].Placeholder = placeholders[i]
		inputs[i].SetValue(values[i])

		// Add validation
		switch i {
		case 0: // Name
			inputs[i].Validate = func(s string) error {
				return tui.ValidateNotEmpty(s)
			}
		case 1: // IP Address
			inputs[i].Validate = func(s string) error {
				if err := tui.ValidateNotEmpty(s); err != nil {
					return err
				}
				return tui.ValidateIPOrHostname(s)
			}
		case 2: // Port
			inputs[i].Validate = func(s string) error {
				if strings.TrimSpace(s) == "" {
					return nil // Allow empty for default
				}
				return tui.ValidatePort(s)
			}
		case 4: // Private Key UUID
			inputs[i].Validate = func(s string) error {
				return tui.ValidateNotEmpty(s)
			}
		}
	}

	inputs[0].Focus()
	return addModel{
		inputs:           inputs,
		validateCheckbox: validate,
		err:              nil,
		keys:             addKeys,
		help:             help.New(),
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

		// Handle space for checkbox toggle
		if key.Matches(msg, m.keys.Space) && m.focusIndex == len(m.inputs) {
			m.validateCheckbox = !m.validateCheckbox
			return m, nil
		}

		if key.Matches(msg, m.keys.Enter) {
			// Check if we're on the checkbox field
			if m.focusIndex == len(m.inputs) {
				// Move to submit button
				m.focusIndex++
				return m, nil
			}

			// Submit on enter when submit button is focused
			if m.focusIndex == len(m.inputs)+1 {
				// Validate all fields before submitting
				for i, input := range m.inputs {
					if input.Validate != nil {
						value := strings.TrimSpace(input.Value())
						if err := input.Validate(value); err != nil {
							m.err = fmt.Errorf("validation failed for %s: %w", []string{"Name", "IP", "Port", "User", "Private Key"}[i], err)
							return m, nil
						}
					}
				}

				m.done = true
				return m, tea.Quit
			}

			// Otherwise move to next field
			m.focusIndex++
			if m.focusIndex > len(m.inputs)+1 {
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

			if m.focusIndex > len(m.inputs)+1 {
				m.focusIndex = 0
			} else if m.focusIndex < 0 {
				m.focusIndex = len(m.inputs) + 1
			}
			m.updateFocus()
		}

		if key.Matches(msg, m.keys.Up) {
			m.focusIndex--
			if m.focusIndex < 0 {
				m.focusIndex = len(m.inputs) + 1
			}
			m.updateFocus()
		}

		if key.Matches(msg, m.keys.Down) {
			m.focusIndex++
			if m.focusIndex > len(m.inputs)+1 {
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
	for i := range m.inputs {
		if i == m.focusIndex {
			m.inputs[i].Focus()
		} else {
			m.inputs[i].Blur()
		}
	}
}

func (m *addModel) updateInputs(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd

	// Only update the focused input if it's a text input
	if m.focusIndex < len(m.inputs) {
		m.inputs[m.focusIndex], cmd = m.inputs[m.focusIndex].Update(msg)
	}

	return cmd
}

func (m addModel) View() string {
	var b strings.Builder

	b.WriteString("Add New Server\n\n")

	// Render text inputs with validation errors
	for i, input := range m.inputs {
		b.WriteString(input.View())

		// Show validation error if present and focused
		if i == m.focusIndex && input.Err != nil {
			b.WriteString("\n")
			b.WriteString(tui.ErrorStyle.Render("  ✗ " + input.Err.Error()))
		}

		b.WriteString("\n")
	}

	// Render validate checkbox
	checkboxLabel := "Validate server after adding: "
	checkbox := unchecked
	if m.validateCheckbox {
		checkbox = checked
	}

	if m.focusIndex == len(m.inputs) {
		b.WriteString(tui.FocusedStyle.Render(checkboxLabel + checkbox))
	} else {
		b.WriteString(checkboxLabel + checkbox)
	}
	b.WriteString("\n")

	// Render submit button
	submitButton := "\n"
	if m.focusIndex == len(m.inputs)+1 {
		submitButton += tui.FocusedStyle.Render("[ Submit ]")
	} else {
		submitButton += "[ Submit ]"
	}

	b.WriteString(submitButton)

	// Add help view
	if m.help.ShowAll {
		b.WriteString("\n\n")
		b.WriteString(m.help.View(m.keys))
	} else {
		b.WriteString("\n\n")
		b.WriteString(m.help.ShortHelpView(m.keys.ShortHelp()))
	}

	// Show general error if any
	if m.err != nil {
		b.WriteString("\n\n")
		b.WriteString(tui.ErrorStyle.Render("Error: " + m.err.Error()))
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
		if parsedResponse.JSON400 != nil && parsedResponse.JSON400.Message != nil {
			return fmt.Errorf("failed to add server: %s", *parsedResponse.JSON400.Message)
		}
		return fmt.Errorf("failed to add server: unexpected status code %d, body: %s", parsedResponse.StatusCode(), string(parsedResponse.Body))
	}

	if validate {
		fmt.Printf("Server added successfully with uuid %s\n", *parsedResponse.JSON201.Uuid)
	} else {
		fmt.Printf("Server added successfully with uuid %s. Server is not validated. Use 'servers validate %s' to validate the server.\n", *parsedResponse.JSON201.Uuid, *parsedResponse.JSON201.Uuid)
	}

	return nil
}
