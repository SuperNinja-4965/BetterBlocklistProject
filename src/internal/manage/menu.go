package manage

import (
	o "better-blocklist/src/internal/operation-on-list"
	t "better-blocklist/src/internal/terminal"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

// AddMenu uses the shared Operation helper and provides an operation which
// requests terminal input and then appends the entered value to the file.
func AddMenu() tea.Cmd {
	return o.Operation("🔧 Select a list to modify", func(path string) tea.Cmd {
		// Ask user for input, then append to file
		return t.RequestInput(fmt.Sprintf("Enter value to add to %s:", path), "", func(input string) tea.Cmd {
			return func() tea.Msg {
				if input == "" {
					return t.OutputLine{Line: "No input provided; cancelled."}
				}
				if err := addToFile(path, input); err != nil {
					return t.OutputLine{Line: err.Error()}
				}
				return t.OutputLine{Line: "Done."}
			}
		}, "Add to list")
	})
}

func RemoveMenu() tea.Cmd {
	return o.Operation("🔧 Select a list to modify", func(path string) tea.Cmd {
		// Ask user for input, then append to file
		return t.RequestInput(fmt.Sprintf("Enter value to remove from %s:", path), "", func(input string) tea.Cmd {
			return func() tea.Msg {
				if input == "" {
					return t.OutputLine{Line: "No input provided; cancelled."}
				}
				if err := removeFromFile(path, input); err != nil {
					return t.OutputLine{Line: err.Error()}
				}
				return t.OutputLine{Line: "Done."}
			}
		}, "Remove from list")
	})
}
