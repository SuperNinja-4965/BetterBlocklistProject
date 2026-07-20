package terminal

import tea "github.com/charmbracelet/bubbletea"

// Custom message types
type OutputLine struct {
	Line string
}

type exportResult struct {
	success bool
	message string
}

type streamOutput struct {
	line string
}

type commandComplete struct {
	success bool
	message string
}

// New message types for input system
type inputRequest struct {
	question     string
	defaultValue string
	handler      func(string) tea.Cmd // function to call with the input result
	originalCmd  string               // the original command that triggered this input
}

type inputComplete struct {
	value   string
	handler func(string) tea.Cmd
}

// Message type for command cancellation
type cancelCommand struct{}

// Message type for command status updates
type commandProgress struct {
	lines []string
	done  bool
}

// Message type for saving output
type saveOutput struct {
	success  bool
	message  string
	filePath string
}

type SubmenuOpen struct {
	Title   string
	Choices []string
	Handler func(string) tea.Cmd
}
