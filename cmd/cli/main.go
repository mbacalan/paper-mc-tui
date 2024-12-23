package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mbacalan/paper-mc-tui/internal/ui/components"
	"github.com/mbacalan/paper-mc-tui/internal/ui/views"
)

type model struct {
	list    components.List
	version string
}

func (m model) View() string {
	return "\n" + m.list.View()
}

func main() {
	manager := views.NewManager()

	if _, err := tea.NewProgram(manager).Run(); err != nil {
		fmt.Printf("Uh oh, there was an error: %v\n", err)
		os.Exit(1)
	}
}
