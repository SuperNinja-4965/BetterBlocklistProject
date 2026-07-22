package main

import (
	"better-blocklist/src/internal/deduplicate"
	"better-blocklist/src/internal/manage"
	t "better-blocklist/src/internal/terminal"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

var choices = []string{
	"De-duplicate list",
	"Add to list",
	"Remove from list",
}

func main() {
	p := tea.NewProgram(t.InitialModel(choices, getCommandForChoice), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}

// Get the appropriate command for each menu choice
func getCommandForChoice(choice string) tea.Cmd {
	switch choice {
	case "De-duplicate list":
		return deduplicate.Menu()

	case "Add to list":
		return manage.AddMenu()

	case "Remove from list":
		return manage.RemoveMenu()

	default:
		return func() tea.Msg {
			return t.OutputLine{Line: "Unknown menu choice: " + choice}
		}
	}
}
