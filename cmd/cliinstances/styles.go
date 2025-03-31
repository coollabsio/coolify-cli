package cliinstances

import "github.com/charmbracelet/lipgloss"

var (
	HeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("99"))

	FocusedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("99"))

	BlurredStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("60"))

	ErrorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("204"))

	CursorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("99"))
)
