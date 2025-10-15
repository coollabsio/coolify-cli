package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Message types for success/error notifications
type clearMessageMsg struct{}

// Duration for how long to show messages
const messageTimeout = 5 * time.Second

// FilterableItem represents an item that can be filtered by name
type FilterableItem interface {
	GetFilterValue() string
}

// KeyMap defines keybindings for the table view
type KeyMap struct {
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
func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Help, k.Quit}
}

// FullHelp returns keybindings for the expanded help view
func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down}, // first column
		{k.Sensitive, k.Detail, k.Delete, k.Help}, // second column
		{k.Quit}, // third column
	}
}

// DefaultKeyMap returns the default key bindings
func DefaultKeyMap() KeyMap {
	return KeyMap{
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
			key.WithHelp("enter", "show details"),
		),
		Delete: key.NewBinding(
			key.WithKeys("delete"),
			key.WithHelp("delete", "delete item"),
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
			key.WithKeys("ctrl+c"),
			key.WithHelp("ctrl+c", "quit"),
		),
	}
}

// FilterableTable manages a filterable table view
type FilterableTable struct {
	Table             table.Model
	FilterInput       textinput.Model
	Items             []FilterableItem
	Columns           []table.Column
	RowBuilder        func(item FilterableItem) table.Row
	DetailBuilder     func(item FilterableItem, sensitive bool) string
	DeleteHandler     func(item FilterableItem) error
	KeyMap            KeyMap
	Help              help.Model
	Viewport          viewport.Model
	ViewportReady     bool
	detailHeader      string
	ConfirmDeleteMode bool
	DetailMode        bool
	Sensitive         bool
	Width             int
	Height            int
	messageTimer      *time.Timer
	Err               error
	SuccessMsg        string
}

// New creates a new FilterableTable
func NewTableFilter(items []FilterableItem, columns []table.Column, rowBuilder func(item FilterableItem) table.Row) *FilterableTable {
	t := table.New(
		table.WithColumns(columns),
		table.WithFocused(true),
		table.WithStyles(table.Styles{
			Selected: FocusedStyle,
		}),
	)

	// Create the filter input
	filterInput := NewFocusedInput("Filter by name", "Filter: ")

	// Initialize viewport
	vp := viewport.New(0, 0)
	vp.Style = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(0, 2)

	ft := &FilterableTable{
		Table:         t,
		FilterInput:   filterInput,
		Items:         items,
		Columns:       columns,
		RowBuilder:    rowBuilder,
		KeyMap:        DefaultKeyMap(),
		Help:          help.New(),
		Viewport:      vp,
		ViewportReady: true,
	}

	// Initialize table with filtered rows
	ft.updateTableRows()

	return ft
}

// WithInitialFilter sets the initial filter
func (ft *FilterableTable) WithInitialFilter(initialFilter string) *FilterableTable {
	ft.FilterInput.SetValue(initialFilter)
	return ft
}

// WithDetailView adds detail view support
func (ft *FilterableTable) WithDetailView(detailBuilder func(item FilterableItem, sensitive bool) string) *FilterableTable {
	ft.DetailBuilder = detailBuilder
	return ft
}

// WithDeleteHandler adds delete support
func (ft *FilterableTable) WithDeleteHandler(deleteHandler func(item FilterableItem) error) *FilterableTable {
	ft.DeleteHandler = deleteHandler
	return ft
}

// WithViewportHeader sets the header text for the detail view
func (ft *FilterableTable) WithDetailHeader(header string) *FilterableTable {
	ft.detailHeader = header
	return ft
}

// setError sets an error message and starts the clear timer
func (ft *FilterableTable) setError(err error) tea.Cmd {
	ft.Err = err
	ft.SuccessMsg = "" // Clear success when showing error

	// Cancel existing timer if any
	if ft.messageTimer != nil {
		ft.messageTimer.Stop()
	}

	return tea.Tick(messageTimeout, func(t time.Time) tea.Msg {
		return clearMessageMsg{}
	})
}

// setSuccess sets a success message and starts the clear timer
func (ft *FilterableTable) setSuccess(msg string) tea.Cmd {
	ft.SuccessMsg = msg
	ft.Err = nil // Clear error when showing success

	// Cancel existing timer if any
	if ft.messageTimer != nil {
		ft.messageTimer.Stop()
	}

	return tea.Tick(messageTimeout, func(t time.Time) tea.Msg {
		return clearMessageMsg{}
	})
}

func (ft *FilterableTable) Update(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd

	if _, ok := msg.(clearMessageMsg); ok {
		ft.Err = nil
		ft.SuccessMsg = ""
		return nil
	}

	var cmd tea.Cmd
	ft.FilterInput, cmd = ft.updateFilter(msg)
	cmds = append(cmds, cmd)

	ft.Table, cmd = ft.updateTable(msg)
	cmds = append(cmds, cmd)

	return tea.Batch(cmds...)
}

// UpdateFilter updates the filter and refreshes the table rows
func (ft *FilterableTable) updateFilter(msg tea.Msg) (textinput.Model, tea.Cmd) {
	// Ignore help key in filter input
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		if key.Matches(keyMsg, ft.KeyMap.Help) {
			return ft.FilterInput, nil
		}
		// If in confirm delete mode, quit updating filter input from y/n inputs
		if ft.ConfirmDeleteMode && (key.Matches(keyMsg, ft.KeyMap.Cancel) || key.Matches(keyMsg, ft.KeyMap.Confirm)) {
			return ft.FilterInput, nil
		}
	}

	prevFilter := ft.FilterInput.Value()
	var cmd tea.Cmd
	ft.FilterInput, cmd = ft.FilterInput.Update(msg)

	// Update filtered rows when filter changes
	if prevFilter != ft.FilterInput.Value() {
		ft.updateTableRows()
	}

	return ft.FilterInput, cmd
}

// UpdateTable updates the table with the given message
func (ft *FilterableTable) updateTable(msg tea.Msg) (table.Model, tea.Cmd) {
	var cmds []tea.Cmd

	if keyMsg, ok := msg.(tea.KeyMsg); ok {

		switch {
		case key.Matches(keyMsg, ft.KeyMap.Detail):
			if ft.ConfirmDeleteMode || ft.DetailMode {
				return ft.Table, nil
			}
			if ft.DetailBuilder != nil {
				// Update viewport content with current item's details
				if item, ok := ft.GetSelectedItem(); ok {
					ft.DetailMode = true
					content := ft.DetailBuilder(item, ft.Sensitive)
					ft.Viewport.SetContent(content)
					// Reset viewport position
					ft.Viewport.GotoTop()
				}
				return ft.Table, nil
			}
		case key.Matches(keyMsg, ft.KeyMap.Delete):
			// We have a delete handler and we are not in confirm delete mode or detail mode
			if ft.DeleteHandler != nil && !ft.ConfirmDeleteMode && !ft.DetailMode {
				ft.ConfirmDeleteMode = true
				return ft.Table, nil
			}
		case key.Matches(keyMsg, ft.KeyMap.Confirm):
			if ft.ConfirmDeleteMode && ft.DeleteHandler != nil {
				if item, ok := ft.GetSelectedItem(); ok {
					if err := ft.DeleteHandler(item); err != nil {
						cmds = append(cmds, ft.setError(err))
					} else {
						// Remove the item from the Items slice
						for i, existing := range ft.Items {
							if existing == item {
								ft.Items = append(ft.Items[:i], ft.Items[i+1:]...)
								break
							}
						}
						// Update the table rows to reflect the deletion
						ft.updateTableRows()
						cmds = append(cmds, ft.setSuccess(fmt.Sprintf("Successfully deleted %s", item.GetFilterValue())))
					}
				}
				ft.ConfirmDeleteMode = false
				return ft.Table, tea.Batch(cmds...)
			}
		case key.Matches(keyMsg, ft.KeyMap.Cancel):
			if ft.ConfirmDeleteMode {
				ft.ConfirmDeleteMode = false
				return ft.Table, nil
			}
			if ft.DetailMode {
				ft.DetailMode = false
				return ft.Table, nil
			}
		case key.Matches(keyMsg, ft.KeyMap.Sensitive):
			ft.Sensitive = !ft.Sensitive
			// Update viewport content if in detail mode
			if ft.DetailMode {
				if item, ok := ft.GetSelectedItem(); ok {
					content := ft.DetailBuilder(item, ft.Sensitive)
					ft.Viewport.SetContent(content)
				}
			}
			return ft.Table, nil
		case key.Matches(keyMsg, ft.KeyMap.Help):
			ft.Help.ShowAll = !ft.Help.ShowAll
			return ft.Table, nil
		case key.Matches(keyMsg, ft.KeyMap.Quit):
			return ft.Table, tea.Quit
		}
	}

	// Handle window size updates
	if msg, ok := msg.(tea.WindowSizeMsg); ok {
		ft.Width = msg.Width
		ft.Height = msg.Height

		// Calculate viewport height accounting for header, footer, and help
		viewportHeight := ft.Height - 4 // Base reduction for borders/help
		if ft.detailHeader != "" {
			viewportHeight -= 4 // Account for header and spacing
		}

		ft.Viewport.Width = msg.Width
		ft.Viewport.Height = viewportHeight

		// Update content to fit new size if in detail mode
		if ft.DetailMode {
			if item, ok := ft.GetSelectedItem(); ok {
				content := ft.DetailBuilder(item, ft.Sensitive)
				ft.Viewport.SetContent(content)
			}
		}
		return ft.Table, nil
	}

	var cmd tea.Cmd
	if ft.DetailMode {
		// Update viewport when in detail mode
		var viewportCmd tea.Cmd
		ft.Viewport, viewportCmd = ft.Viewport.Update(msg)
		return ft.Table, viewportCmd
	}

	ft.Table, cmd = ft.Table.Update(msg)
	return ft.Table, cmd
}

// FilteredItems returns the currently filtered items
func (ft *FilterableTable) FilteredItems() []FilterableItem {
	filter := strings.ToLower(ft.FilterInput.Value())
	if filter == "" {
		return ft.Items
	}

	var filtered []FilterableItem
	for _, item := range ft.Items {
		if strings.Contains(strings.ToLower(item.GetFilterValue()), filter) {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

// GetSelectedItem returns the currently selected item
func (ft *FilterableTable) GetSelectedItem() (FilterableItem, bool) {
	filtered := ft.FilteredItems()
	if len(filtered) == 0 || ft.Table.Cursor() >= len(filtered) {
		return nil, false
	}
	return filtered[ft.Table.Cursor()], true
}

// updateTableRows updates the table rows based on the current filter
func (ft *FilterableTable) updateTableRows() {
	filtered := ft.FilteredItems()
	rows := make([]table.Row, len(filtered))
	for i, item := range filtered {
		rows[i] = ft.RowBuilder(item)
	}
	ft.Table.SetRows(rows)

	// Ensure cursor stays within bounds
	switch length := len(filtered); {
	case length == 0:
		ft.Table.SetCursor(0)
	case ft.Table.Cursor() >= length:
		ft.Table.SetCursor(length - 1)
	case ft.Table.Cursor() < 0:
		ft.Table.SetCursor(0)
	}
}

// View returns the combined view of filter input, table, and help
func (ft *FilterableTable) View() string {
	var s strings.Builder

	if ft.DetailMode && ft.DetailBuilder != nil {
		if !ft.ViewportReady {
			return "Loading..."
		}
		if ft.detailHeader != "" {
			s.WriteString(FocusedStyle.Bold(true).Render(ft.detailHeader))
			s.WriteString("\n\n")
		}
		s.WriteString(ft.Viewport.View())

		footer := lipgloss.JoinHorizontal(
			lipgloss.Center,
			BlurredStyle.Render("↑/↓: scroll • ESC: back"),
			strings.Repeat(" ", 3),
			BlurredStyle.Render(fmt.Sprintf("%.f%%", ft.Viewport.ScrollPercent()*100)),
		)

		s.WriteString("\n" + footer)
		s.WriteString("\n\n")
		s.WriteString(ft.Help.View(ft.KeyMap))
		return s.String()
	}

	if ft.ConfirmDeleteMode {
		if item, ok := ft.GetSelectedItem(); ok {
			confirmStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("204")).
				BorderStyle(lipgloss.NormalBorder()).
				BorderForeground(lipgloss.Color("204")).
				Padding(0, 1)
			confirmMsg := fmt.Sprintf("Delete %s? [y/N]", item.GetFilterValue())
			return confirmStyle.Render(confirmMsg)
		}
		ft.ConfirmDeleteMode = false
	}

	s.WriteString(ft.FilterInput.View())
	s.WriteString("\n\n")
	s.WriteString(ft.Table.View())
	s.WriteString("\n\n")
	s.WriteString(ft.Help.View(ft.KeyMap))
	s.WriteString("\n\n")
	if ft.Err != nil {
		s.WriteString(ErrorStyle.Render(fmt.Sprintf("Error: %v\n", ft.Err)))
		s.WriteString("\n")
	}

	if ft.SuccessMsg != "" {
		s.WriteString(SuccessStyle.Render(ft.SuccessMsg))
		s.WriteString("\n")
	}
	return s.String()
}
