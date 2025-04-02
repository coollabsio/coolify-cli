package cliprivatekeys

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Styles
var (
	FocusedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("99"))
	BlurredStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("60"))
	ErrorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("204"))
	SuccessStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("40"))
)

// Messages for BubbleTea
type deleteMsg struct{}
type confirmDeleteMsg struct{}
type cancelDeleteMsg struct{}
type deleteSuccessMsg struct{ UUID string }

// listKeyMap defines keybindings for the private keys list view
type listKeyMap struct {
	Up        key.Binding
	Down      key.Binding
	Sensitive key.Binding
	Delete    key.Binding
	Confirm   key.Binding
	Cancel    key.Binding
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
		{k.Up, k.Down},                  // first column
		{k.Sensitive, k.Delete, k.Help}, // second column
		{k.Quit},                        // third column
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
	Sensitive: key.NewBinding(
		key.WithKeys("ctrl+s"),
		key.WithHelp("ctrl+s", "toggle sensitive info"),
	),
	Delete: key.NewBinding(
		key.WithKeys("d", "delete"),
		key.WithHelp("d", "delete selected key"),
	),
	Confirm: key.NewBinding(
		key.WithKeys("y", "Y"),
		key.WithHelp("y", "confirm"),
	),
	Cancel: key.NewBinding(
		key.WithKeys("n", "N", "esc"),
		key.WithHelp("n", "cancel"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "toggle help"),
	),
	Quit: key.NewBinding(
		key.WithKeys("esc", "ctrl+c", "q"),
		key.WithHelp("esc", "quit"),
	),
}

type listModel struct {
	keys              []PrivateKey
	filterInput       textinput.Model
	sensitive         bool
	width             int
	height            int
	cursor            int
	selected          int
	err               error
	successMsg        string
	successTimer      int
	table             table.Model
	tableHeight       int
	keymap            listKeyMap
	help              help.Model
	confirmDeleteMode bool
	deleteFunc        func(uuid string) error
}

func newListModel(keys []PrivateKey, sensitive bool, initialFilter string, deleteFunc func(uuid string) error) listModel {
	t := table.New(
		table.WithColumns([]table.Column{
			{Title: "UUID", Width: 30},
			{Title: "Name", Width: 40},
			{Title: "Public Key", Width: 20},
		}),
		table.WithFocused(true),
		table.WithHeight(0),
		table.WithStyles(table.Styles{
			Selected: FocusedStyle,
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
		keys:        keys,
		sensitive:   sensitive,
		filterInput: filterInput,
		cursor:      0,
		selected:    0,
		table:       t,
		tableHeight: 0,
		keymap:      listKeys,
		help:        help.New(),
		deleteFunc:  deleteFunc,
	}
}

func deleteKey(uuid string, deleteFunc func(uuid string) error) tea.Cmd {
	return func() tea.Msg {
		err := deleteFunc(uuid)
		if err != nil {
			return err
		}
		return deleteSuccessMsg{UUID: uuid}
	}
}

func (m listModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m listModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// If we're in confirm delete mode, handle those keys first
		if m.confirmDeleteMode {
			switch {
			case key.Matches(msg, m.keymap.Confirm):
				filtered := m.filteredKeys()
				if len(filtered) == 0 || m.selected >= len(filtered) {
					m.confirmDeleteMode = false
					return m, nil
				}

				uuid := filtered[m.selected].UUID
				m.confirmDeleteMode = false

				// Execute deletion as a command to handle async
				return m, deleteKey(uuid, m.deleteFunc)

			case key.Matches(msg, m.keymap.Cancel):
				m.confirmDeleteMode = false
				return m, nil
			}
			return m, nil
		}

		// Normal mode key handling
		switch {
		case key.Matches(msg, m.keymap.Quit):
			return m, tea.Quit

		case key.Matches(msg, m.keymap.Help):
			m.help.ShowAll = !m.help.ShowAll
			return m, nil

		case key.Matches(msg, m.keymap.Up):
			if m.selected > 0 {
				m.selected--
				if m.selected < m.cursor {
					m.cursor = m.selected
				}
			}
			return m, nil

		case key.Matches(msg, m.keymap.Down):
			filtered := m.filteredKeys()
			if m.selected < len(filtered)-1 {
				m.selected++
				if m.selected >= m.cursor+m.tableHeight-1 {
					m.cursor = m.selected - m.tableHeight + 1
				}
			}
			return m, nil

		case key.Matches(msg, m.keymap.Sensitive):
			m.sensitive = !m.sensitive
			return m, nil

		case key.Matches(msg, m.keymap.Delete):
			filtered := m.filteredKeys()
			if len(filtered) > 0 && m.selected < len(filtered) {
				m.confirmDeleteMode = true
			}
			return m, nil
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// Calculate available height for the table
		// Reserve space for title (2 lines), filter (2 lines), help (2 lines), and possible confirmation (2 lines)
		availableHeight := m.height - 8
		if availableHeight < 5 {
			availableHeight = 5 // Minimum height
		}
		m.tableHeight = availableHeight
		m.table.SetHeight(availableHeight)
		m.help.Width = msg.Width

	case error:
		m.err = msg
		return m, nil

	case deleteSuccessMsg:
		// Remove the deleted key from the keys slice
		for i, key := range m.keys {
			if key.UUID == msg.UUID {
				// Remove this key from the slice
				m.keys = append(m.keys[:i], m.keys[i+1:]...)

				// Set success message
				m.successMsg = fmt.Sprintf("Private key deleted successfully")
				m.successTimer = 20 // Show for a short time

				// Adjust selection if needed
				filtered := m.filteredKeys()
				if len(filtered) == 0 {
					m.selected = 0
					m.cursor = 0
				} else if m.selected >= len(filtered) {
					m.selected = len(filtered) - 1
				}
				break
			}
		}
		return m, nil
	}

	// Handle filter input updates
	m.filterInput, cmd = m.filterInput.Update(msg)

	// Reset cursor and selection position when filter changes
	if m.filterInput.Value() != m.getFilterValue() {
		m.selected = 0
		m.cursor = 0
	}

	// Tick down the success message timer if it's active
	if m.successTimer > 0 {
		m.successTimer--
		if m.successTimer == 0 {
			m.successMsg = ""
		}
	}

	return m, cmd
}

// Helper to get the previous filter value for change detection
func (m listModel) getFilterValue() string {
	return m.filterInput.Value()
}

func (m listModel) filteredKeys() []PrivateKey {
	filter := m.filterInput.Value()
	if filter == "" {
		return m.keys
	}
	filtered := make([]PrivateKey, 0)
	for _, key := range m.keys {
		if strings.Contains(strings.ToLower(key.Name), strings.ToLower(filter)) {
			filtered = append(filtered, key)
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
	s.WriteString(FocusedStyle.Copy().Bold(true).Render("SSH Private Keys") + "\n\n")

	// Render filter input
	s.WriteString(m.filterInput.View())
	s.WriteString("\n\n")

	// Update table rows
	rows := make([]table.Row, 0)
	filtered := m.filteredKeys()
	start := m.cursor
	end := min(m.cursor+m.tableHeight, len(filtered))

	if end > start {
		for i := start; i < end; i++ {
			key := filtered[i]
			publicKey := key.PublicKey

			if !m.sensitive && publicKey != "" {
				publicKey = "********"
			}

			rows = append(rows, table.Row{
				key.UUID,
				key.Name,
				publicKey,
			})
		}
	}

	m.table.SetRows(rows)
	// Set the selected row to be relative to the current viewport
	if len(filtered) > 0 {
		m.table.SetCursor(m.selected - m.cursor)
	}

	// Render table
	tableView := m.table.View()
	s.WriteString(tableView)

	// Show confirmation dialog if in delete mode
	if m.confirmDeleteMode && len(filtered) > 0 && m.selected < len(filtered) {
		key := filtered[m.selected]
		confirmStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("204")).
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("204")).
			Padding(0, 1)

		confirmMsg := fmt.Sprintf("Delete private key '%s'? [y/N]", key.Name)
		s.WriteString("\n\n" + confirmStyle.Render(confirmMsg))
	}

	// Success message (if any)
	if m.successMsg != "" {
		s.WriteString("\n\n")
		s.WriteString(SuccessStyle.Render(m.successMsg))
	}

	// Error message (if any)
	if m.err != nil {
		s.WriteString("\n\n")
		s.WriteString(ErrorStyle.Render(m.err.Error()))
	}

	// Help
	s.WriteString("\n\n")
	s.WriteString(m.help.View(m.keymap))

	return s.String()
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
