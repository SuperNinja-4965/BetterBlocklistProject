package terminal

import (
	"fmt"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// Function to request input from user
func RequestInput(question string, defaultValue string, handler func(string) tea.Cmd, originalCmd string) tea.Cmd {
	return func() tea.Msg {
		return inputRequest{
			question:     question,
			defaultValue: defaultValue,
			handler:      handler,
			originalCmd:  originalCmd,
		}
	}
}

// Function to open a submenu with the provided title, choices, and handler.
func Submenu(title string, choices []string, handler func(string) tea.Cmd) tea.Cmd {
	return func() tea.Msg {
		return SubmenuOpen{
			Title:   title,
			Choices: append([]string(nil), choices...),
			Handler: handler,
		}
	}
}

// Function to save output to file
func saveOutputToFile(OutputLines []string) tea.Cmd {
	return func() tea.Msg {
		// Generate filename with timestamp
		timestamp := time.Now().Format("20060102_150405")
		filename := fmt.Sprintf("manager_output_%s.txt", timestamp)

		// Join all output lines
		content := strings.Join(OutputLines, "\n")

		// Write to file
		err := os.WriteFile(filename, []byte(content), 0644)
		if err != nil {
			return saveOutput{
				success:  false,
				message:  err.Error(),
				filePath: "",
			}
		}

		return saveOutput{
			success:  true,
			message:  "Output saved successfully",
			filePath: filename,
		}
	}
}
