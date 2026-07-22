package operationonlist

import (
	h "better-blocklist/src/internal/helpers"
	t "better-blocklist/src/internal/terminal"

	"fmt"
	"os"
	"sort"

	tea "github.com/charmbracelet/bubbletea"
)

func getLists() ([]string, error) {
	dir := fmt.Sprintf("%sLists", h.GetCurrentDir())

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var lists []string

	for _, entry := range entries {
		if !entry.IsDir() {
			lists = append(lists, entry.Name())
		}
	}

	sort.Strings(lists)

	// Add "All" at the top
	return append([]string{"All"}, lists...), nil
}

// Operation runs an operation that returns a tea.Cmd. This allows callers to
// either perform synchronous work (wrap an error-returning function) or to
// return a command that requests input from the terminal UI.
func Operation(title string, operation func(string) tea.Cmd) tea.Cmd {
	lists, err := getLists()
	if err != nil {
		return func() tea.Msg {
			return t.OutputLine{Line: "Failed to load lists: " + err.Error()}
		}
	}

	choices := append([]string(nil), lists...)
	choices = append(choices, "Back")

	return t.Submenu(title, choices, func(selected string) tea.Cmd {
		switch selected {
		case "All":
			return func() tea.Msg {
				var output []string
				for _, list := range lists {
					if list == "All" {
						continue
					}

					file := fmt.Sprintf("%sLists/%s", h.GetCurrentDir(), list)
					cmd := operation(file)
					if cmd == nil {
						output = append(output, fmt.Sprintf("%s: skipped", list))
						continue
					}

					// Execute the returned command synchronously and gather result
					msg := cmd()
					switch m := msg.(type) {
					case t.OutputLine:
						// Use the message as-is if it contains informative text
						output = append(output, fmt.Sprintf("%s: %s", list, m.Line))
					default:
						output = append(output, fmt.Sprintf("%s: done", list))
					}
				}

				if len(output) == 0 {
					output = append(output, "No lists were processed.")
				}

				return t.OutputLine{Line: fmt.Sprintf("%s\n\nDone.", h.JoinLines(output))}
			}
		case "Back":
			return nil
		default:
			file := fmt.Sprintf("%sLists/%s", h.GetCurrentDir(), selected)
			exists, err := h.FileExists(file)
			if !exists || err != nil {
				return func() tea.Msg {
					return t.OutputLine{Line: "Unknown list: " + selected}
				}
			}

			return func() tea.Msg {
				cmd := operation(file)
				if cmd == nil {
					return t.OutputLine{Line: selected + ": skipped"}
				}
				return cmd()
			}
		}
	})
}
