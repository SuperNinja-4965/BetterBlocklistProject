package main

import (
	"better-blocklist/src/internal/deduplicate"
	t "better-blocklist/src/internal/terminal"
	"fmt"
	"os"

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

// Get the appropriate command for each menu choice
func getCommandForChoice(choice string) tea.Cmd {
	switch choice {
	case "De-duplicate list":
		return deduplicate.Menu()

	default:
		return func() tea.Msg {
			return t.OutputLine{Line: "Unknown menu choice: " + choice}
		}
	}
}
