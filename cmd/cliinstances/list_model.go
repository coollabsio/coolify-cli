package cliinstances

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/coollabsio/coolify-cli/cmd/coolTypes"
	"github.com/coollabsio/coolify-cli/cmd/emoji"
	"github.com/coollabsio/coolify-cli/cmd/utils"
)

type listModel struct {
	instances   []coolTypes.Instance
	filterInput textinput.Model
	sensitive   bool
	width       int
	height      int
	cursor      int
	selected    int
	err         error
	table       table.Model
	tableHeight int
}

func newListModel(instances []coolTypes.Instance, sensitive bool, initialFilter string) listModel {
	t := table.New(
		table.WithColumns([]table.Column{
			{Title: "Name", Width: 30},
			{Title: "URL", Width: 40},
			{Title: "Token", Width: 20},
			{Title: "Default", Width: 8},
		}),
		table.WithFocused(true),
		table.WithHeight(0),
		table.WithStyles(table.Styles{
			Selected: lipgloss.NewStyle().
				Foreground(lipgloss.Color("99")),
		}),
	)

	// Create the filter input
	filterInput := textinput.New()
	filterInput.Placeholder = "Filter by name"
	filterInput.Prompt = "Filter: "
	filterInput.PromptStyle = FocusedStyle
	filterInput.TextStyle = FocusedStyle
	filterInput.Focus()
	filterInput.SetValue(initialFilter)

	return listModel{
		instances:   instances,
		sensitive:   sensitive,
		filterInput: filterInput,
		cursor:      0,
		selected:    0,
		table:       t,
		tableHeight: 0,
	}
}

func (m listModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m listModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return m, tea.Quit
		case "ctrl+c":
			return m, tea.Quit
		case "ctrl+v":
			// Add paste functionality to the filter input
			utils.PasteToInput(&m.filterInput)
			return m, nil
		case "ctrl+x":
			// Add cut functionality to the filter input
			utils.CutFromInput(&m.filterInput)
			return m, nil
		case "up":
			if m.selected > 0 {
				m.selected--
				if m.selected < m.cursor {
					m.cursor = m.selected
				}
			}
			return m, nil
		case "down":
			filtered := m.filteredInstances()
			if m.selected < len(filtered)-1 {
				m.selected++
				if m.selected >= m.cursor+m.tableHeight-1 {
					m.cursor = m.selected - m.tableHeight + 1
				}
			}
			return m, nil
		case "ctrl+s":
			m.sensitive = !m.sensitive
			return m, nil
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// Calculate available height for the table
		// Reserve space for title (2 lines), filter (2 lines), and footer (1 line)
		availableHeight := m.height - 5
		if availableHeight < 5 {
			availableHeight = 5 // Minimum height
		}
		m.tableHeight = availableHeight
		m.table.SetHeight(availableHeight)
	}

	// Handle filter input updates
	m.filterInput, cmd = m.filterInput.Update(msg)

	// Reset cursor and selection position when filter changes
	if m.filterInput.Value() != m.getFilterValue() {
		m.selected = 0
		m.cursor = 0
	}

	return m, cmd
}

// Helper to get the previous filter value for change detection
func (m listModel) getFilterValue() string {
	return m.filterInput.Value()
}

func (m listModel) filteredInstances() []coolTypes.Instance {
	filter := m.filterInput.Value()
	if filter == "" {
		return m.instances
	}
	filtered := make([]coolTypes.Instance, 0)
	for _, instance := range m.instances {
		if strings.Contains(strings.ToLower(instance.Name), strings.ToLower(filter)) {
			filtered = append(filtered, instance)
		}
	}
	return filtered
}

func (m listModel) View() string {
	if m.width == 0 {
		return "loading..."
	}

	var s strings.Builder

	// Title and filter
	s.WriteString("List Instances\n\n")

	// Render filter input
	s.WriteString(m.filterInput.View())
	s.WriteString("\n\n")

	// Update table rows
	rows := make([]table.Row, 0)
	filtered := m.filteredInstances()
	start := m.cursor
	end := min(m.cursor+m.tableHeight, len(filtered))

	if end > start {
		for i := start; i < end; i++ {
			instance := filtered[i]
			token := instance.Token
			if !m.sensitive && token != "" {
				token = "********"
			}

			e := emoji.CrossMark
			if instance.Default {
				e = emoji.CheckMarkButton
			}

			rows = append(rows, table.Row{
				instance.Name,
				instance.Fqdn,
				token,
				e,
			})
		}
	}

	m.table.SetRows(rows)
	// Set the selected row to be relative to the current viewport
	m.table.SetCursor(m.selected - m.cursor)

	// Render table and remove the height indicator
	tableView := m.table.View()
	tableView = strings.TrimSuffix(tableView, "\n10")
	s.WriteString(tableView)

	// Footer
	s.WriteString("\n")
	s.WriteString(fmt.Sprintf("Sensitive: %v (press 'ctrl+s' to toggle)", m.sensitive))
	s.WriteString("\n")
	s.WriteString(BlurredStyle.Render("Shortcuts: ↑/↓ to navigate • Ctrl+V to paste • Ctrl+X to cut • Ctrl+C to exit"))

	// Error message
	if m.err != nil {
		s.WriteString("\n\n")
		s.WriteString(ErrorStyle.Render(m.err.Error()))
	}

	return s.String()
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
