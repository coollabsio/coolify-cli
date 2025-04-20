package cliservers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/coollabsio/cli-coolify/cmd/runtime"
	"github.com/coollabsio/cli-coolify/cmd/utils"
	"github.com/coollabsio/cli-coolify/pkg/gen/openapi"
	"github.com/coollabsio/cli-coolify/pkg/tui"
	"github.com/spf13/cobra"
)

type validateModel struct {
	spinner  spinner.Model
	uuid     string
	done     bool
	err      error
	response string
	coolify  runtime.Getter
	ctx      context.Context
}

type validateSuccessMsg struct {
	message string
}

type validateErrorMsg struct {
	err error
}

func (c *cliServers) newValidateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validate [uuid]",
		Short: "Validate server connection",
		Long: `
Validate the connection to a server in your Coolify instance.
This will check if the server is reachable and usable.`,
		Example: utils.GetCommandExample(`
%[1]s servers validate 123e4567-e89b-12d3-a456-426614174000`),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			uuid := args[0]

			p := tea.NewProgram(initialValidateModel(uuid, c.coolify, cmd.Context()))
			model, err := p.Run()
			if err != nil {
				return fmt.Errorf("error running validation: %w", err)
			}

			finalModel := model.(validateModel)
			if finalModel.err != nil {
				return finalModel.err
			}

			return nil
		},
	}

	return cmd
}

func initialValidateModel(uuid string, coolify runtime.Getter, ctx context.Context) validateModel {
	s := spinner.New()
	s.Spinner = spinner.Points
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("99"))

	return validateModel{
		spinner: s,
		uuid:    uuid,
		coolify: coolify,
		ctx:     ctx,
	}
}

func (m validateModel) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		m.validateServer,
	)
}

func (m validateModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case validateSuccessMsg:
		m.done = true
		m.response = msg.message
		return m, tea.Quit

	case validateErrorMsg:
		m.done = true
		m.err = msg.err
		return m, tea.Quit
	}

	return m, nil
}

func (m validateModel) View() string {
	if m.done {
		if m.err != nil {
			return tui.ErrorStyle.Render(fmt.Sprintf("Error: %v\n", m.err))
		}
		return tui.SuccessStyle.Render(m.response + "\n")
	}

	return fmt.Sprintf("%s Validating server...\n", m.spinner.View())
}

func (m validateModel) validateServer() tea.Msg {
	// Simulate network delay for better UX
	time.Sleep(500 * time.Millisecond)

	server, err := m.coolify().Client.ValidateServerByUuid(m.ctx, m.uuid)
	if err != nil {
		return validateErrorMsg{err: fmt.Errorf("failed to validate server: %w", err)}
	}

	parsedResponse, err := openapi.ParseValidateServerByUuidResponse(server)
	if err != nil {
		return validateErrorMsg{err: fmt.Errorf("failed to parse server response: %w", err)}
	}

	if parsedResponse.StatusCode() != http.StatusCreated {
		switch parsedResponse.StatusCode() {
		case http.StatusBadRequest:
			return validateErrorMsg{err: fmt.Errorf("failed to validate server: %s", *parsedResponse.JSON400.Message)}
		case http.StatusNotFound:
			return validateErrorMsg{err: fmt.Errorf("failed to validate server: %s", *parsedResponse.JSON404.Message)}
		default:
			return validateErrorMsg{err: fmt.Errorf("failed to validate server: %s", string(parsedResponse.Body))}
		}
	}

	return validateSuccessMsg{message: string(*parsedResponse.JSON201.Message)}
}
