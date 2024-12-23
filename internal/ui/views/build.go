package views

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mbacalan/paper-mc-tui/internal/ui/components"
	"github.com/mbacalan/paper-mc-tui/internal/utils"
)

type BuildView struct {
	version string
	build   string
}

func NewBuildView() *BuildView {
	return &BuildView{}
}

func (v *BuildView) Init() tea.Cmd {
	version, _ := utils.GetLatestStableVersion()
	build, _ := utils.GetLatestBuild(version)
	v.build = build
	return nil
}

func (v *BuildView) Update(msg tea.Msg) (View, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return v, tea.Quit

		case "esc":
			return v, func() tea.Msg {
				return SwitchViewMsg{ViewID: HomeViewID}
			}
		}
	}

	return v, cmd
}

func (v *BuildView) View() string {
	style := lipgloss.NewStyle().Margin(1, 2)
	help := components.NewHelp()

	if v.build != "" {
		buildText := style.Render(fmt.Sprintf("Latest available build is %s\n\n", v.build))
		return "\n" + buildText + help.View()
	}

	return "\n" + "Unable to get latest build!\n" + help.View()
}
