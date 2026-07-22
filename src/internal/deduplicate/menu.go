package deduplicate

import (
	o "better-blocklist/src/internal/operation-on-list"
	t "better-blocklist/src/internal/terminal"

	tea "github.com/charmbracelet/bubbletea"
)

func Menu() tea.Cmd {
	// Wrap SortFile (which returns error) into a tea.Cmd
	return o.Operation("🔧 Select a list to deduplicate", func(path string) tea.Cmd {
		return func() tea.Msg {
			if err := SortFile(path); err != nil {
				return t.OutputLine{Line: err.Error()}
			}
			return t.OutputLine{Line: "Done."}
		}
	})
}
