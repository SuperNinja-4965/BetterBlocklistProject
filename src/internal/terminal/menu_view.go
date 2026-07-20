package terminal

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m model) menuView() string {
	statusBar := statusBarStyle.Width(m.width).Render(m.menuTitle)

	var listContent string
	for i, choice := range m.choices {
		if m.cursor == i {
			listContent += cursorStyle.Render("▶ ") + selectedItemStyle.Render(choice)
		} else {
			listContent += "  " + normalItemStyle.Render(choice)
		}
		if i < len(m.choices)-1 {
			listContent += "\n"
		}
	}

	currentLines := len(strings.Split(listContent, "\n"))
	visibleHeight := m.height - outputStyle.GetVerticalFrameSize() - 2

	for currentLines < visibleHeight {
		listContent += "\n"
		currentLines++
	}

	list := listStyle.Width(m.width - 4).Render(listContent)
	footer := footerStyle.Render("Press ↑/↓ or j/k to move • Enter to select • q to quit")

	return lipgloss.JoinVertical(
		lipgloss.Left,
		statusBar,
		list,
		footer,
	)
}
