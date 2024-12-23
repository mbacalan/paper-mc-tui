package views

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mbacalan/paper-mc-tui/internal/ui/components"
	"github.com/mbacalan/paper-mc-tui/internal/utils"
)

type VersionView struct {
	version string
}

const url = "https://api.papermc.io/v2"

func NewVersionView() *VersionView {
	return &VersionView{}
}

func (v *VersionView) Init() tea.Cmd {
	version, _ := utils.GetLatestStableVersion()
	v.version = string(version)
	return nil
}

func (v *VersionView) Update(msg tea.Msg) (View, tea.Cmd) {
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

func (v *VersionView) View() string {
	style := lipgloss.NewStyle().Margin(1, 2)

	var keys = components.KeyMap{
		Back: key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "back")),
		Quit: key.NewBinding(key.WithKeys("q"), key.WithHelp("q", "quit")),
	}
	help := components.NewHelp(keys)

	if v.version != "" {
		versionText := style.Render(fmt.Sprintf("Latest available version is %s\n\n", v.version))

		return versionText + help.View()
	}

	return "Unable to get latest version!\n" + help.View()
}
