package views

import tea "github.com/charmbracelet/bubbletea"

// backToHome is a tea.Cmd that switches back to the home menu.
func backToHome() tea.Msg {
	return SwitchViewMsg{ViewID: HomeViewID}
}
