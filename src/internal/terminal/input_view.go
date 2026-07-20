package terminal

import (
	"strings"

	textinput "github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
)

// inputView renders a focused input prompt view with status and bottom bars
func (m model) inputView() string {
	statusBar := statusBarStyle.Width(m.width).Render("🔧 " + m.originalCommand + " - Input Required")

	// Use bubbles textinput for a polished input control
	ti := textinput.New()
	ti.Prompt = "Input: "
	ti.CharLimit = 18 // Set character limit to 18
	ti.Placeholder = ""
	ti.SetValue(m.inputValue)
	ti.SetCursor(m.inputCursor)
	ti.Focus()

	// Render the main input panel with question + text input view
	inputValues := m.inputQuestion + "\n\n" + ti.View()
	currentLines := len(strings.Split(inputValues, "\n"))
	visibleHeight := m.height - outputStyle.GetVerticalFrameSize() - 2

	for currentLines < visibleHeight {
		inputValues += "\n"
		currentLines++
	}

	panel := outputStyle.Width(m.width - 4).Render(inputValues)

	// Bottom hints
	bottomBar := bottomBarStyle.Width(m.width).Render("Enter value • Press Enter to confirm • Esc to cancel")

	return lipgloss.JoinVertical(
		lipgloss.Left,
		statusBar,
		panel,
		bottomBar,
	)
}
