package terminal

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Define styles
var (
	// List styles
	listStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#874BFD")).
			Padding(0, 2).
			Margin(0, 0)

	// Item styles
	selectedItemStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#EE6FF8")).
				Bold(true)

	cursorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF06B7")).
			Bold(true)

	normalItemStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FAFAFA"))

	// Footer style
	footerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#AD58B4")).
			Italic(true).
			Margin(0, 0)

	// Status bar style
	statusBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#874BFD")).
			Bold(true).
			Padding(0, 1).
			Width(100) // Ensure minimum width

	// Output view styles
	outputStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#874BFD")).
			Padding(0, 0).
			Margin(0, 0)

	bottomBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#42BF71")).
			Bold(true).
			Padding(0, 1)
)

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	// Handle export result
	case exportResult:
		m.OutputLines = []string{msg.message}
		if !msg.success {
			m.OutputLines = append(m.OutputLines, "", "Export failed. Please check the directory path and permissions.")
		} else {
			m.OutputLines = append(m.OutputLines, "", "Export completed successfully!")
		}
		m.execComplete = true
		m.awaitingInput = false // Reset input state
		return m, nil

	// Handle input request
	case inputRequest:
		m.showOutput = true
		m.awaitingInput = true
		m.inputQuestion = msg.question
		m.inputValue = msg.defaultValue
		m.inputCursor = len(msg.defaultValue) // Set cursor to end of default value
		m.inputHandler = msg.handler
		m.originalCommand = msg.originalCmd
		m.OutputLines = []string{msg.question}
		m.execComplete = false
		m.scrollOffset = 0
		return m, nil

	// Handle submenu opening
	case SubmenuOpen:
		m.choices = append([]string(nil), msg.Choices...)
		m.cursor = 0
		m.menuTitle = msg.Title
		m.submenuHandler = msg.Handler
		m.showOutput = false
		m.awaitingInput = false
		m.execComplete = false
		m.OutputLines = nil
		m.scrollOffset = 0
		return m, nil

	// Handle input completion
	case inputComplete:
		m.awaitingInput = false
		m.OutputLines = append(m.OutputLines, "", "Processing input: "+msg.value)
		return m, msg.handler(msg.value)

	// Handle streaming output
	case streamOutput:
		m.OutputLines = append(m.OutputLines, msg.line)
		return m, nil

	// Handle command completion
	case commandComplete:
		if !msg.success {
			m.OutputLines = append(m.OutputLines, "", "Command failed: "+msg.message)
		} else {
			m.OutputLines = append(m.OutputLines, "", "Command completed successfully!")
		}
		m.execComplete = true
		return m, nil

	// Handle custom messages
	case OutputLine:
		// Split the joined output back into lines
		m.OutputLines = strings.Split(msg.Line, "\n")
		m.execComplete = true
		m.commandRunning = false
		m.cancelRequested = false
		return m, nil

	// Handle command progress updates
	case commandProgress:
		m.OutputLines = msg.lines
		if !msg.done {
			m.commandRunning = true
		} else {
			m.execComplete = true
			m.commandRunning = false
			m.cancelRequested = false
		}
		return m, nil

	// Handle cancellation
	case cancelCommand:
		m.OutputLines = append(m.OutputLines, "", "Command execution cancelled by user.")
		m.execComplete = true
		m.commandRunning = false
		m.cancelRequested = false
		return m, nil

	// Handle save output result
	case saveOutput:
		if msg.success {
			m.OutputLines = append(m.OutputLines, "", "Output saved to: "+msg.filePath)
			m.saveStatus = "Saved to " + msg.filePath
		} else {
			m.OutputLines = append(m.OutputLines, "", "Failed to save output: "+msg.message)
			m.saveStatus = "Save failed: " + msg.message
		}

	// Handle window size changes
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	// Is it a key press?
	case tea.KeyMsg:
		// Handle input (renamed from directory input)
		if m.awaitingInput {
			switch msg.Type {
			case tea.KeyEnter:
				// Process the input
				inputVal := strings.TrimSpace(m.inputValue)
				m.awaitingInput = false
				m.execComplete = false
				m.scrollOffset = 0
				return m, func() tea.Msg {
					return inputComplete{
						value:   inputVal,
						handler: m.inputHandler,
					}
				}
			case tea.KeyCtrlC:
				return m, tea.Quit
			case tea.KeyEscape:
				// Cancel input and return to menu
				m.awaitingInput = false
				m.showOutput = false
				m.inputValue = ""
				m.inputCursor = 0
				return m, nil
			case tea.KeyLeft:
				// Move cursor left
				if m.inputCursor > 0 {
					m.inputCursor--
				}
			case tea.KeyRight:
				// Move cursor right
				if m.inputCursor < len(m.inputValue) {
					m.inputCursor++
				}
			case tea.KeyBackspace:
				// Delete character before cursor
				if m.inputCursor > 0 && len(m.inputValue) > 0 {
					m.inputValue = m.inputValue[:m.inputCursor-1] + m.inputValue[m.inputCursor:]
					m.inputCursor--
				}
			case tea.KeyDelete:
				// Delete character at cursor
				if m.inputCursor < len(m.inputValue) {
					m.inputValue = m.inputValue[:m.inputCursor] + m.inputValue[m.inputCursor+1:]
				}
			case tea.KeyHome:
				// Move cursor to beginning
				m.inputCursor = 0
			case tea.KeyEnd:
				// Move cursor to end
				m.inputCursor = len(m.inputValue)
			default:
				// Handle different backspace sequences that some terminals send
				keyStr := msg.String()
				if keyStr == "\b" || keyStr == "\x7f" || keyStr == "backspace" {
					// Alternative backspace handling for different terminals
					if m.inputCursor > 0 && len(m.inputValue) > 0 {
						m.inputValue = m.inputValue[:m.inputCursor-1] + m.inputValue[m.inputCursor:]
						m.inputCursor--
					}
				} else if len(keyStr) == 1 {
					// Add character at cursor position
					m.inputValue = m.inputValue[:m.inputCursor] + keyStr + m.inputValue[m.inputCursor:]
					m.inputCursor++
				}
			}
			return m, nil
		}

		// Handle keys differently based on current view
		if m.showOutput {
			// Output view key handling
			switch msg.String() {
			case "q", "esc", "enter":
				if m.execComplete {
					// Return to menu
					m.showOutput = false
					m.OutputLines = nil
					m.execComplete = false
				}
			case "c":
				// Cancel running command
				if m.commandRunning && !m.cancelRequested {
					m.cancelRequested = true
					m.OutputLines = append(m.OutputLines, "", "Cancellation requested... Please wait")
					// Cancel the currently running command
					cancelCurrentCommand()
					return m, func() tea.Msg {
						return cancelCommand{}
					}
				}
			case "s":
				// Save output to file
				if len(m.OutputLines) > 0 {
					return m, saveOutputToFile(m.OutputLines)
				}
			case "up", "k":
				if m.scrollOffset > 0 {
					m.scrollOffset--
				}
			case "down", "j":
				maxScroll := len(m.OutputLines) - (m.height - 12)
				if maxScroll > 0 && m.scrollOffset < maxScroll {
					m.scrollOffset++
				}
			case "ctrl+c":
				return m, tea.Quit
			}
		} else {
			// Menu view key handling
			switch msg.String() {
			case "ctrl+c", "q", "esc":
				return m, tea.Quit
			case "up", "k":
				if m.cursor > 0 {
					m.cursor--
				}
			case "down", "j":
				if m.cursor < len(m.choices)-1 {
					m.cursor++
				}
			case "enter":
				if m.choices[m.cursor] == "Exit" {
					return m, tea.Quit
				}
				if m.choices[m.cursor] == "Back" {
					m.choices = append([]string(nil), m.rootChoices...)
					m.cursor = 0
					m.menuTitle = "🔧 Better Blocklist Manager"
					m.submenuHandler = nil
					return m, nil
				}
				if m.submenuHandler != nil {
					selected := m.choices[m.cursor]
					if selected == "All" {
						m.showOutput = true
						m.OutputLines = []string{}
						m.execComplete = false
						m.commandRunning = false
						m.cancelRequested = false
						m.scrollOffset = 0
						return m, m.submenuHandler(selected)
					}
					if selected != "Back" {
						m.showOutput = true
						m.OutputLines = []string{}
						m.execComplete = false
						m.commandRunning = false
						m.cancelRequested = false
						m.scrollOffset = 0
						return m, m.submenuHandler(selected)
					}
				}
				// Execute selected submenu command.
				m.showOutput = true
				m.OutputLines = []string{}
				m.execComplete = false
				m.commandRunning = false // Will be set to true when command starts
				m.cancelRequested = false
				m.scrollOffset = 0
				return m, m.choiceFunc(m.choices[m.cursor])
			}
		}
	}

	return m, nil
}
