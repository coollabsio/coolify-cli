package tui

// TUI is the package for TUI components of Coolify cli.

import (
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
)

var (
	FocusedStyle = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{
		Light: "#6B2E85", // Darker purple for light theme
		Dark:  "#875FFF", // ANSI color 99
	})
	BlurredStyle = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{
		Light: "#6B6B87", // Muted grayish-purple for light theme (similar to ANSI 60)
		Dark:  "#5F5F87", // ANSI color 60
	})
	ErrorStyle = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{
		Light: "#B22B31", // Darker red for light theme
		Dark:  "#FF5C5C", // Bright red for dark theme
	})
	SuccessStyle = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{
		Light: "#2D5A27", // Darker green for light theme
		Dark:  "#4CAF50", // Bright green for dark theme
	})
)

func NewTextInput(placeholder, prompt string, style lipgloss.Style) textinput.Model {
	t := textinput.New()
	t.Placeholder = placeholder
	t.Prompt = prompt
	t.PromptStyle = style
	t.TextStyle = style
	return t
}

func NewFocusedInput(placeholder, prompt string) textinput.Model {
	t := NewTextInput(placeholder, prompt, FocusedStyle)
	t.Focus()
	return t
}

func NewBlurredInput(placeholder, prompt string) textinput.Model {
	return NewTextInput(placeholder, prompt, BlurredStyle)
}
