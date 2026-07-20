package terminal

import tea "github.com/charmbracelet/bubbletea"

type model struct {
	choices         []string             // menu options
	rootChoices     []string             // top-level menu options
	cursor          int                  // which menu item our cursor is pointing at
	width           int                  // terminal width
	height          int                  // terminal height
	showOutput      bool                 // whether we're showing output view
	menuTitle       string               // title shown in the status bar
	submenuHandler  func(string) tea.Cmd // optional submenu selection handler
	OutputLines     []string             // lines of output to display
	scrollOffset    int                  // scroll position in output
	execComplete    bool                 // whether execution is complete
	awaitingInput   bool                 // whether we're waiting for user input
	inputValue      string               // current input value
	inputCursor     int                  // cursor position in input field
	inputQuestion   string               // question being asked
	inputHandler    func(string) tea.Cmd // function to call with input result
	originalCommand string               // the original command that triggered input
	commandRunning  bool                 // whether a command is currently running
	cancelRequested bool                 // whether user requested cancellation
	saveStatus      string               // save status message
	selected        map[int]struct{}     // which to-do items are selected
	choiceFunc      func(string) tea.Cmd // Function for choice selection
}

func InitialModel(choices []string, choiceFunc func(string) tea.Cmd) model {
	rootChoices := append([]string(nil), choices...)
	rootChoices = append(rootChoices, "Exit")

	return model{
		// Menu options for Blocklist Manager
		choices:     rootChoices,
		rootChoices: rootChoices,
		menuTitle:   "🔧 Better Blocklist Manager",
		// A map which indicates which choices are selected. We're using
		// the  map like a mathematical set. The keys refer to the indexes
		// of the `choices` slice, above.
		selected:   make(map[int]struct{}),
		choiceFunc: choiceFunc,
	}
}

func (m model) Init() tea.Cmd {
	// Just return `nil`, which means "no I/O right now, please."
	return nil
}
