package views

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mbacalan/paper-mc-tui/internal/ui/components"
	"github.com/mbacalan/paper-mc-tui/internal/utils"
)

type CurrentBuildView struct {
	logger *utils.Logger
	build  string
}

func NewCurrentBuildView() *CurrentBuildView {
	logger, err := utils.NewLogger()

	if err != nil {
		fmt.Printf("Error creating logger: %v\n", err)
	}

	return &CurrentBuildView{
		logger: logger,
	}
}

func (v *CurrentBuildView) Init() tea.Cmd {
	build, _ := v.logger.GetLastDownloadedVersion()
	v.build = build
	return nil
}

func (v *CurrentBuildView) Update(msg tea.Msg) (View, tea.Cmd) {
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

func (v *CurrentBuildView) View() string {
	style := lipgloss.NewStyle().Margin(1, 2)
	help := components.NewHelp()

	if v.build != "" {
		buildText := style.Render(fmt.Sprintf("Current build is %s", v.build))
		noteText := style.Render("Note: this is according to logs!\nIt might be incorrect if you've updated manually.\n\n")

		return lipgloss.JoinVertical(lipgloss.Left, buildText, noteText) + help.View()
	}

	return "Unable to get current build!\n" + help.View()
}
