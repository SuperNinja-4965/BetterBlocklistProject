package deduplicate

import (
	o "better-blocklist/src/internal/operation-on-list"

	tea "github.com/charmbracelet/bubbletea"
)

func Menu() tea.Cmd {
	return o.Operation("🔧 Select a list to deduplicate", SortFile)
}
