package terminal

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m model) outputView() string {
	// Output view shows command output only; input is handled in input_view.go
	statusBarText := "🔧 Better Blocklist Manager - Command Output"

	statusBar := statusBarStyle.Width(m.width).Render(statusBarText)

	visibleHeight := m.height - outputStyle.GetVerticalFrameSize() - 5
	start := m.scrollOffset
	end := start + visibleHeight

	var outputContent string
	for i := start; i < end && i < len(m.OutputLines); i++ {
		outputContent += m.OutputLines[i] + "\n"
	}

	currentLines := len(strings.Split(outputContent, "\n")) - 3
	for currentLines < visibleHeight {
		outputContent += "\n"
		currentLines++
	}

	output := outputStyle.Width(m.width - 4).Render(outputContent)

	var bottomBar string
	if m.execComplete {
		bottomBar = bottomBarStyle.Width(m.width).Render("Execution completed. Press 'q' to return to menu • 's' to save output")
	} else if m.commandRunning {
		if m.cancelRequested {
			bottomBar = bottomBarStyle.Width(m.width).Render("Cancelling... Please wait")
		} else {
			bottomBar = bottomBarStyle.Width(m.width).Render("Executing... Press 'c' to cancel")
		}
	} else {
		bottomBar = bottomBarStyle.Width(m.width).Render("Executing... Please wait")
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		statusBar,
		output,
		bottomBar,
	)
}
