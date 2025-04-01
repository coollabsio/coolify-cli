package cliinstances

import (
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/coollabsio/cli-coolify/cmd/coolTypes"
	"github.com/coollabsio/cli-coolify/cmd/emoji"
)

// listKeyMap defines keybindings for the list instances view
type listKeyMap struct {
	Up        key.Binding
	Down      key.Binding
	Paste     key.Binding
	Sensitive key.Binding
	Help      key.Binding
	Quit      key.Binding
}

// ShortHelp returns keybindings to be shown in the mini help view
func (k listKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Help, k.Quit}
}

// FullHelp returns keybindings for the expanded help view
func (k listKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down},                 // first column
		{k.Paste, k.Sensitive, k.Help}, // second column
		{k.Quit},                       // third column
	}
}

var listKeys = listKeyMap{
	Up: key.NewBinding(
		key.WithKeys("up"),
		key.WithHelp("↑", "move up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down"),
		key.WithHelp("↓", "move down"),
	),
	Paste: key.NewBinding(
		key.WithKeys("ctrl+v"),
		key.WithHelp("ctrl+v", "paste"),
	),
	Sensitive: key.NewBinding(
		key.WithKeys("ctrl+s"),
		key.WithHelp("ctrl+s", "toggle sensitive info"),
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
	keys        listKeyMap
	help        help.Model
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
		keys:        listKeys,
		help:        help.New(),
	}
}

func (m listModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m listModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, m.keys.Help):
			m.help.ShowAll = !m.help.ShowAll
			return m, nil
		case key.Matches(msg, m.keys.Up):
			if m.selected > 0 {
				m.selected--
				if m.selected < m.cursor {
					m.cursor = m.selected
				}
			}
			return m, nil
		case key.Matches(msg, m.keys.Down):
			filtered := m.filteredInstances()
			if m.selected < len(filtered)-1 {
				m.selected++
				if m.selected >= m.cursor+m.tableHeight-1 {
					m.cursor = m.selected - m.tableHeight + 1
				}
			}
			return m, nil
		case key.Matches(msg, m.keys.Sensitive):
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
		m.help.Width = msg.Width
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

	// Help view
	s.WriteString("\n\n")
	s.WriteString(m.help.View(m.keys))

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
