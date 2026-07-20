package terminal

func (m model) View() string {
	if m.width <= 11 || m.height <= 11 {
		return "Terminal is too small to display content.\nPlease resize your terminal.\nPress 'q' or 'esc' to quit."
	}
	// Prioritize input view when awaiting input
	if m.awaitingInput {
		return m.inputView()
	}
	if m.showOutput {
		return m.outputView()
	}
	return m.menuView()
}
