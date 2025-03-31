package cliinstances

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/coollabsio/coolify-cli/cmd/coolTypes"
	"github.com/coollabsio/coolify-cli/cmd/emoji"
)

type listModel struct {
	instances    []coolTypes.Instance
	filter       string
	sensitive    bool
	width        int
	height       int
	cursor       int
	selected     int
	err          error
	table        table.Model
	filterCursor int
	tableHeight  int
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

	return listModel{
		instances:    instances,
		sensitive:    sensitive,
		filter:       initialFilter,
		cursor:       0,
		selected:     0,
		table:        t,
		filterCursor: len(initialFilter),
		tableHeight:  0,
	}
}

func (m listModel) Init() tea.Cmd {
	return nil
}

func (m listModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return m, tea.Quit
		case "ctrl+c":
			return m, tea.Quit
		case "up":
			if m.selected > 0 {
				m.selected--
				if m.selected < m.cursor {
					m.cursor = m.selected
				}
			}
		case "down":
			filtered := m.filteredInstances()
			if m.selected < len(filtered)-1 {
				m.selected++
				if m.selected >= m.cursor+m.tableHeight-1 {
					m.cursor = m.selected - m.tableHeight + 1
				}
			}
		case "ctrl+s":
			m.sensitive = !m.sensitive
		case "backspace":
			if len(m.filter) > 0 && m.filterCursor > 0 {
				m.filter = m.filter[:m.filterCursor-1] + m.filter[m.filterCursor:]
				m.filterCursor--
				m.selected = 0
				m.cursor = 0
			}
		case "left":
			if m.filterCursor > 0 {
				m.filterCursor--
			}
		case "right":
			if m.filterCursor < len(m.filter) {
				m.filterCursor++
			}
		default:
			if len(msg.String()) == 1 {
				// Insert character at cursor position
				m.filter = m.filter[:m.filterCursor] + msg.String() + m.filter[m.filterCursor:]
				m.filterCursor++
				m.selected = 0
				m.cursor = 0
			}
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

	return m, nil
}

func (m listModel) filteredInstances() []coolTypes.Instance {
	if m.filter == "" {
		return m.instances
	}
	filtered := make([]coolTypes.Instance, 0)
	for _, instance := range m.instances {
		if strings.Contains(strings.ToLower(instance.Name), strings.ToLower(m.filter)) {
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

	// Render filter with cursor
	filterDisplay := m.filter
	if m.filterCursor < len(filterDisplay) {
		filterDisplay = filterDisplay[:m.filterCursor] + CursorStyle.Render("█") + filterDisplay[m.filterCursor:]
	} else {
		filterDisplay = filterDisplay + CursorStyle.Render("█")
	}
	s.WriteString(fmt.Sprintf("Filter: %s\n\n", filterDisplay))

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
