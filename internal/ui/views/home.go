package views

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mbacalan/paper-mc-tui/internal/ui/components"
	"github.com/mbacalan/paper-mc-tui/internal/ui/styles"
)

// HomeView represents the main view of the application
type HomeView struct {
	list   components.List
	styles styles.DefaultStyles
	choice string
}

const (
	HomeViewID ViewID = iota
	VersionViewID
)

func NewHomeView(s styles.DefaultStyles) *HomeView {
	items := []components.Item{
		"Check latest version",
		"Quit",
	}

	return &HomeView{
		list:   components.New(items, s),
		styles: s,
	}
}

func (v *HomeView) Init() tea.Cmd {
	v.list.SetTitle("PaperMC Management CLI")

	return nil
}

func (v *HomeView) Update(msg tea.Msg) (View, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		v.list.SetWidth(msg.Width)
		return v, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return v, tea.Quit

		case "enter":
			i, ok := v.list.SelectedItem()
			if ok {
				v.choice = string(i)

				switch v.choice {
				case "Check latest version":
					return v, func() tea.Msg {
						return SwitchViewMsg{ViewID: VersionViewID}
					}
				case "Quit":
					return v, tea.Quit
				}
			}
		}
	}

	v.list, cmd = v.list.Update(msg)
	return v, cmd
}

func (v *HomeView) View() string {
	return "\n" + v.list.View()
}
