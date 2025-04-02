package cliprivatekeys

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/coollabsio/cli-coolify/cmd/runtime"
	"github.com/coollabsio/cli-coolify/cmd/utils"
	"github.com/spf13/cobra"
)

// addKeyModel is the Bubble Tea model for the interactive add key form
type addKeyModel struct {
	focusIndex int
	inputs     []textinput.Model
	done       bool
	err        error
	coolify    *runtime.Coolify
}

func initialAddKeyModel(coolify *runtime.Coolify) addKeyModel {
	m := addKeyModel{
		inputs:  make([]textinput.Model, 2),
		coolify: coolify,
	}

	// Setup name input
	nameInput := textinput.New()
	nameInput.Placeholder = "My SSH Key"
	nameInput.Focus()
	nameInput.CharLimit = 50
	nameInput.Width = 40
	nameInput.Prompt = "› "
	nameInput.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("99"))
	m.inputs[0] = nameInput

	// Setup key input (multi-line)
	keyInput := textinput.New()
	keyInput.Placeholder = "SSH private key or path to key file"
	keyInput.CharLimit = 4096
	keyInput.Width = 60
	keyInput.Prompt = "› "
	keyInput.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("60"))
	m.inputs[1] = keyInput

	return m
}

func (m addKeyModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m addKeyModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit

		case "tab", "shift+tab", "enter", "up", "down":
			// Cycle focus between inputs
			if msg.String() == "enter" && m.focusIndex == len(m.inputs)-1 {
				// Submit the form
				m.done = true
				return m, tea.Quit
			}

			// Cycle indexes
			if msg.String() == "up" || msg.String() == "shift+tab" {
				m.focusIndex--
				if m.focusIndex < 0 {
					m.focusIndex = len(m.inputs) - 1
				}
			} else {
				m.focusIndex++
				if m.focusIndex >= len(m.inputs) {
					m.focusIndex = 0
				}
			}

			cmds := make([]tea.Cmd, len(m.inputs))
			for i := range m.inputs {
				if i == m.focusIndex {
					cmds[i] = m.inputs[i].Focus()
					m.inputs[i].PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("99"))
				} else {
					m.inputs[i].Blur()
					m.inputs[i].PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("60"))
				}
			}

			return m, tea.Batch(cmds...)
		}
	}

	// Handle character input for the active input
	cmd := m.updateInputs(msg)
	return m, cmd
}

func (m *addKeyModel) updateInputs(msg tea.Msg) tea.Cmd {
	var cmds = make([]tea.Cmd, len(m.inputs))

	for i := range m.inputs {
		m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
	}

	return tea.Batch(cmds...)
}

func (m addKeyModel) View() string {
	var b strings.Builder

	// Title with Coolify branding
	title := lipgloss.NewStyle().
		Foreground(lipgloss.Color("99")).
		Bold(true).
		Render("Add New SSH Private Key")
	b.WriteString(title + "\n\n")

	// Render inputs with labels
	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("60")).
		Width(12)

	b.WriteString(labelStyle.Render("Name:") + " " + m.inputs[0].View() + "\n\n")
	b.WriteString(labelStyle.Render("Private Key:") + " " + m.inputs[1].View() + "\n\n")

	// Instructions
	b.WriteString("\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("60")).Render("(Tab/Shift+Tab to navigate, Enter to submit)"))

	return b.String()
}

func (c *cliPrivateKeys) newAddCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add [name] [private_key_or_file]",
		Short: "Add a new private key",
		Long: `Add a new SSH private key to your Coolify instance.
The key can be provided directly as a string or as a path to a file.

If no arguments are provided, an interactive form will be used.`,
		Example: utils.GetCommandExample(`
%[1]s private-keys add "My Key" /path/to/id_rsa
%[1]s private-keys add "My Key" "-----BEGIN RSA PRIVATE KEY-----..."
%[1]s private-keys add  # Interactive mode
`),
		Args: cobra.RangeArgs(0, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Interactive mode when no arguments are provided
			if len(args) == 0 {
				model := initialAddKeyModel(c.coolify())
				p := tea.NewProgram(model)
				finalModel, err := p.Run()
				if err != nil {
					return fmt.Errorf("error running interactive mode: %w", err)
				}

				// Process the final model after user submission
				finalState := finalModel.(addKeyModel)
				if !finalState.done {
					return fmt.Errorf("operation canceled")
				}

				name := finalState.inputs[0].Value()
				privateKeyInput := finalState.inputs[1].Value()

				return c.addPrivateKey(cmd.Context(), name, privateKeyInput)
			}

			// CLI mode with arguments
			if len(args) != 2 {
				return fmt.Errorf("requires both NAME and PRIVATE_KEY_OR_FILE arguments")
			}

			name := args[0]
			privateKeyInput := args[1]

			return c.addPrivateKey(cmd.Context(), name, privateKeyInput)
		},
	}

	return cmd
}

// addPrivateKey adds a private key to the Coolify instance
func (c *cliPrivateKeys) addPrivateKey(ctx context.Context, name, privateKeyInput string) error {
	// Check if input is a file path
	var privateKey string
	if _, err := os.Stat(privateKeyInput); err == nil {
		keyBytes, err := os.ReadFile(privateKeyInput)
		if err != nil {
			return fmt.Errorf("error reading private key file: %w", err)
		}
		privateKey = string(keyBytes)
	} else {
		privateKey = privateKeyInput
	}

	// Prepare request data
	data := map[string]string{
		"name":        name,
		"private_key": privateKey,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	req, err := c.coolify().NewRequest(ctx, http.MethodPost, "security/keys", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	_, err = c.coolify().DoRequest(req)
	if err != nil {
		return fmt.Errorf("failed to add private key: %w", err)
	}

	fmt.Printf("Private key '%s' added successfully\n", name)
	return nil
}
