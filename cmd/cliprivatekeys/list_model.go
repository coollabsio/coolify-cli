package cliprivatekeys

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/coollabsio/cli-coolify/pkg/tui"
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
	Detail    key.Binding
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
		{k.Up, k.Down}, // first column
		{k.Sensitive, k.Detail, k.Delete, k.Help}, // second column
		{k.Quit}, // third column
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
	Detail: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "show details on selected key"),
	),
	Delete: key.NewBinding(
		key.WithKeys("delete"),
		key.WithHelp("delete", "delete selected key"),
	),
	Confirm: key.NewBinding(
		key.WithKeys("y", "Y"),
		key.WithHelp("y", "confirm"),
	),
	Cancel: key.NewBinding(
		key.WithKeys("n", "N", "esc"),
		key.WithHelp("n", "cancel deletion"),
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
	keys              []PrivateKey
	filterInput       textinput.Model
	sensitive         bool
	width             int
	height            int
	err               error
	successMsg        string
	successTimer      int
	table             table.Model
	tableHeight       int
	keymap            listKeyMap
	help              help.Model
	confirmDeleteMode bool
	detailMode        bool
	deleteFunc        func(uuid string) error
	viewport          viewport.Model // Add viewport for scrollable content
	viewportReady     bool           // Track if viewport is ready
}

func newListModel(keys []PrivateKey, sensitive bool, initialFilter string, deleteFunc func(uuid string) error) listModel {
	t := table.New(
		table.WithColumns([]table.Column{
			{Title: "UUID", Width: 30},
			{Title: "Name", Width: 40},
			{Title: "Description", Width: 30},
		}),
		table.WithFocused(true),
		table.WithHeight(0),
		table.WithStyles(table.Styles{
			Selected: tui.FocusedStyle,
		}),
	)

	// Create the filter input
	filterInput := tui.NewFocusedInput("Filter by name", "Filter: ")
	filterInput.SetValue(initialFilter)

	return listModel{
		keys:        keys,
		sensitive:   sensitive,
		filterInput: filterInput,
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

// Helper method to update table rows and maintain cursor state
func (m *listModel) updateTableRows() {
	filtered := m.filteredKeys()
	rows := m.getTableRows(filtered)
	m.table.SetRows(rows)

	// Ensure cursor stays within bounds
	if len(filtered) == 0 {
		m.table.SetCursor(0)
	} else {
		// If cursor is out of bounds or no item is selected, select the first item
		if len(filtered) <= m.table.Cursor() || m.table.Cursor() < 0 {
			m.table.SetCursor(0)
		}
	}
}

func (m *listModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Common key handlers across all modes
		switch {
		case key.Matches(msg, m.keymap.Help):
			m.help.ShowAll = !m.help.ShowAll
			return m, nil // Return immediately to prevent help key from reaching filter
		case key.Matches(msg, m.keymap.Sensitive):
			m.sensitive = !m.sensitive
			if m.detailMode && m.viewportReady {
				// Update viewport content when toggling sensitive info
				m.updateViewportContent()
			}
		}

		// Mode-specific handlers
		if m.detailMode {
			// Detail mode handlers
			if key.Matches(msg, m.keymap.Quit) {
				m.detailMode = false
				m.viewportReady = false
			} else {
				// Let viewport handle scrolling
				var cmd tea.Cmd
				m.viewport, cmd = m.viewport.Update(msg)
				cmds = append(cmds, cmd)
			}
			return m, tea.Batch(cmds...)
		}

		// Only handle filter input updates in normal mode
		if !m.detailMode && !m.confirmDeleteMode {
			prevFilter := m.filterInput.Value()
			newFilterInput, cmd := m.filterInput.Update(msg)
			m.filterInput = newFilterInput
			if cmd != nil {
				cmds = append(cmds, cmd)
			}

			// Update filtered rows when filter changes
			if prevFilter != m.filterInput.Value() {
				m.updateTableRows()
			}
		}

		if m.confirmDeleteMode {
			// Confirm delete mode handlers
			switch {
			case key.Matches(msg, m.keymap.Confirm):
				filtered := m.filteredKeys()
				if len(filtered) > 0 && m.table.Cursor() < len(filtered) {
					uuid := filtered[m.table.Cursor()].UUID
					m.confirmDeleteMode = false
					cmds = append(cmds, deleteKey(uuid, m.deleteFunc))
				} else {
					m.confirmDeleteMode = false
				}
			case key.Matches(msg, m.keymap.Cancel):
				m.confirmDeleteMode = false
			}
			return m, tea.Batch(cmds...)
		}

		// Normal mode handlers
		switch {
		case key.Matches(msg, m.keymap.Quit):
			cmds = append(cmds, tea.Quit)

		case key.Matches(msg, m.keymap.Up), key.Matches(msg, m.keymap.Down):
			if len(m.filteredKeys()) > 0 {
				if key.Matches(msg, m.keymap.Up) {
					m.table.MoveUp(1)
				} else {
					m.table.MoveDown(1)
				}
			}

		case key.Matches(msg, m.keymap.Detail):
			filtered := m.filteredKeys()
			if len(filtered) > 0 && m.table.Cursor() < len(filtered) {
				m.detailMode = true
				// Initialize viewport when entering detail mode
				if !m.viewportReady {
					m.initializeViewport()
				}
			}

		case key.Matches(msg, m.keymap.Delete):
			filtered := m.filteredKeys()
			if len(filtered) > 0 && m.table.Cursor() < len(filtered) {
				m.confirmDeleteMode = true
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		if m.detailMode {
			headerHeight := 4 // Adjust based on your header
			footerHeight := 4 // Adjust based on your footer
			m.viewport.Width = msg.Width
			m.viewport.Height = msg.Height - headerHeight - footerHeight
		} else {
			// Calculate available height for the table
			availableHeight := m.height - 8
			if availableHeight < 5 {
				availableHeight = 5 // Minimum height
			}
			m.tableHeight = availableHeight
			m.table.SetHeight(availableHeight)
		}
		m.help.Width = msg.Width

		// Update table rows after resize
		if !m.detailMode {
			m.updateTableRows()
		}

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

				// Update filtered rows and adjust cursor
				m.updateTableRows()
				break
			}
		}
	}

	// Tick down the success message timer if it's active
	if m.successTimer > 0 {
		m.successTimer--
		if m.successTimer == 0 {
			m.successMsg = ""
		}
	}

	return m, tea.Batch(cmds...)
}

// Helper function to convert filtered keys to table rows
func (m listModel) getTableRows(filtered []PrivateKey) []table.Row {
	rows := make([]table.Row, 0, len(filtered))
	for _, key := range filtered {
		rows = append(rows, table.Row{
			key.UUID,
			key.Name,
			key.Description,
		})
	}
	return rows
}

func (m listModel) View() string {
	if m.width == 0 {
		return "loading..."
	}

	var s strings.Builder

	// If in detail mode, show the detailed view
	if m.detailMode {
		filtered := m.filteredKeys()
		if len(filtered) > 0 && m.table.Cursor() < len(filtered) {
			// Title
			s.WriteString(tui.FocusedStyle.Bold(true).Render("SSH Private Key Details") + "\n")

			// Render viewport content
			if m.viewportReady {
				s.WriteString(m.viewport.View())
			}

			// Footer with help text and scroll percentage
			footerStyle := lipgloss.NewStyle().
				PaddingTop(1).
				BorderStyle(lipgloss.NormalBorder()).
				BorderTop(true)

			footer := lipgloss.JoinHorizontal(
				lipgloss.Center,
				tui.BlurredStyle.Render("↑/↓: scroll • ESC: back"),
				strings.Repeat(" ", 3),
				tui.BlurredStyle.Render(fmt.Sprintf("%.f%%", m.viewport.ScrollPercent()*100)),
			)

			s.WriteString("\n" + footerStyle.Render(footer))
			return s.String()
		}
		return "No key selected"
	}

	// Title and filter
	s.WriteString(tui.FocusedStyle.Bold(true).Render("SSH Private Keys") + "\n\n")

	// Render filter input
	s.WriteString(m.filterInput.View())
	s.WriteString("\n\n")

	// Render table
	s.WriteString(m.table.View())

	// Show confirmation dialog if in delete mode
	filtered := m.filteredKeys()
	if m.confirmDeleteMode && len(filtered) > 0 && m.table.Cursor() < len(filtered) {
		key := filtered[m.table.Cursor()]
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
		s.WriteString(tui.SuccessStyle.Render(m.successMsg))
	}

	// Error message (if any)
	if m.err != nil {
		s.WriteString("\n\n")
		s.WriteString(tui.ErrorStyle.Render(m.err.Error()))
	}

	// Help
	s.WriteString("\n\n")
	s.WriteString(m.help.View(m.keymap))

	return s.String()
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

// Helper method to initialize viewport
func (m *listModel) initializeViewport() {
	m.viewport = viewport.New(m.width, m.height-8)
	m.viewport.Style = lipgloss.NewStyle().Padding(1, 2)
	m.updateViewportContent()
	m.viewportReady = true
}

// Helper method to update viewport content
func (m *listModel) updateViewportContent() {
	filtered := m.filteredKeys()
	if len(filtered) > 0 && m.table.Cursor() < len(filtered) {
		key := filtered[m.table.Cursor()]
		var content strings.Builder

		// Helper function to add a detail section
		addSection := func(title, value string) {
			content.WriteString(tui.FocusedStyle.Bold(true).Render(title + ": "))
			content.WriteString(value + "\n\n")
		}

		addSection("UUID", key.UUID)
		addSection("Name", key.Name)
		addSection("Description", key.Description)

		if m.sensitive {
			addSection("Public Key", "\n"+key.PublicKey)
			addSection("Private Key", "\n"+key.PrivateKey)
		} else {
			addSection("Public Key", "********")
			addSection("Private Key", "********")
		}

		addSection("Git Related", fmt.Sprintf("%v", key.IsGitRelated))
		addSection("Team ID", fmt.Sprintf("%d", key.TeamID))
		addSection("Created At", key.CreatedAt)
		addSection("Updated At", key.UpdatedAt)

		m.viewport.SetContent(content.String())
	}
}
