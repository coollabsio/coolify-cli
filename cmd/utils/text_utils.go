package utils

import (
	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/textinput"
)

// PasteToInput pastes text from clipboard into a textinput at the current cursor position
func PasteToInput(input *textinput.Model) {
	text, err := clipboard.ReadAll()
	if err == nil {
		currentVal := input.Value()
		cursorPos := input.Position()
		newVal := currentVal[:cursorPos] + text + currentVal[cursorPos:]
		input.SetValue(newVal)
		input.SetCursor(cursorPos + len(text))
	}
}

// CutFromInput cuts the current text from a textinput to the clipboard
func CutFromInput(input *textinput.Model) {
	clipboard.WriteAll(input.Value())
	input.SetValue("")
}
