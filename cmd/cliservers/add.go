package cliservers

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/coollabsio/cli-coolify/cmd/utils"
	"github.com/coollabsio/cli-coolify/pkg/gen/openapi"
	"github.com/coollabsio/cli-coolify/pkg/tui"
	"github.com/spf13/cobra"
)

type addModel struct {
	inputs     []textinput.Model
	focusIndex int
	err        error
	done       bool
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
		case "ctrl+c", "esc":
			m.done = false
			return m, tea.Quit
		case "tab", "shift+tab", "enter", "up", "down":
			// Cycle between inputs
			s := msg.String()

			if s == "enter" && m.focusIndex == len(m.inputs)-1 {
				m.done = true
				return m, tea.Quit
			}

			if s == "up" || s == "shift+tab" {
				m.focusIndex--
			} else {
				m.focusIndex++
			}

			if m.focusIndex >= len(m.inputs) {
				m.focusIndex = 0
			} else if m.focusIndex < 0 {
				m.focusIndex = len(m.inputs) - 1
			}

			for i := 0; i < len(m.inputs); i++ {
				if i == m.focusIndex {
					cmds = append(cmds, m.inputs[i].Focus())
				} else {
					m.inputs[i].Blur()
				}
			}

			return m, tea.Batch(cmds...)
		}
	}

	// Handle character input
	cmd := m.updateInputs(msg)

	return m, cmd
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
	b.WriteString("\n\nPress esc to cancel • tab/shift+tab to navigate\n")

	return b.String()
}

func (c *cliServers) addServer(name, ip, privateKeyUUID string, port int, user string, validate bool) error {
	req, err := c.coolify().Client.CreateServer(context.Background(), openapi.CreateServerJSONRequestBody{
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

	parsedResponse, err := openapi.ParseCreateServerResponse(req)
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
