package main

import (
	"better-blocklist/src/internal/deduplicate"
	t "better-blocklist/src/internal/terminal"
	"fmt"
	"os"
	"sort"

	tea "github.com/charmbracelet/bubbletea"
)

var choices = []string{
	"De-duplicate list",
}

func main() {
	p := tea.NewProgram(t.InitialModel(choices, getCommandForChoice), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}

func fileExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// Helper function to get current directory
func getCurrentDir() string {
	if cwd, err := os.Getwd(); err == nil {
		return cwd + "/"
	}
	return "./"
}

func getLists() ([]string, error) {
	dir := fmt.Sprintf("%sLists", getCurrentDir())

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

// Get the appropriate command for each menu choice
func getCommandForChoice(choice string) tea.Cmd {
	switch choice {
	case "De-duplicate list":
		lists, err := getLists()
		if err != nil {
			return func() tea.Msg {
				return t.OutputLine{Line: "Failed to load lists: " + err.Error()}
			}
		}

		choices := append([]string(nil), lists...)
		choices = append(choices, "Back")

		return t.Submenu("🔧 Select a list to deduplicate", choices, func(selected string) tea.Cmd {
			switch selected {
			case "All":
				return func() tea.Msg {
					var output []string
					for _, list := range lists {
						if list == "All" {
							continue
						}

						file := fmt.Sprintf("%sLists/%s", getCurrentDir(), list)
						if err := deduplicate.SortFile(file); err != nil {
							output = append(output, fmt.Sprintf("%s: failed: %v", list, err))
							continue
						}

						output = append(output, fmt.Sprintf("%s: done", list))
					}

					if len(output) == 0 {
						output = append(output, "No lists were processed.")
					}

					return t.OutputLine{Line: fmt.Sprintf("%s\n\nDone.", joinLines(output))}
				}
			case "Back":
				return nil
			default:
				file := fmt.Sprintf("%sLists/%s", getCurrentDir(), selected)
				exists, err := fileExists(file)
				if !exists || err != nil {
					return func() tea.Msg {
						return t.OutputLine{Line: "Unknown list: " + selected}
					}
				}

				return func() tea.Msg {
					if err := deduplicate.SortFile(file); err != nil {
						return t.OutputLine{Line: err.Error()}
					}

					return t.OutputLine{Line: selected + ": done"}
				}
			}
		})

	default:
		return func() tea.Msg {
			return t.OutputLine{Line: "Unknown menu choice: " + choice}
		}
	}
}

func joinLines(lines []string) string {
	if len(lines) == 0 {
		return ""
	}

	result := lines[0]
	for i := 1; i < len(lines); i++ {
		result += "\n" + lines[i]
	}

	return result
}
