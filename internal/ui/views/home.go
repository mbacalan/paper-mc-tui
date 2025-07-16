package views

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mbacalan/paper-mc-tui/internal/ui/components"
)

type HomeView struct {
	list   components.List
	choice string
}

type MenuAction string

const (
	CheckLatestVersion  MenuAction = "Check latest version"
	CheckLatestBuild    MenuAction = "Check latest build"
	CheckInstalledBuild MenuAction = "Check installed build"
	DownloadLatestBuild MenuAction = "Download latest build"
	Quit                MenuAction = "Quit"
)

const (
	HomeViewID ViewID = iota
	VersionViewID
	BuildViewID
	CurrentBuildViewID
	DownloadBuildID
)

func NewHomeView() *HomeView {
	items := []components.Item{
		components.Item(CheckLatestVersion),
		components.Item(CheckLatestBuild),
		components.Item(CheckInstalledBuild),
		components.Item(DownloadLatestBuild),
		components.Item(Quit),
	}

	list := components.NewList(items, "PaperMC Management CLI")

	return &HomeView{
		list: list,
	}
}

func (v *HomeView) Init() tea.Cmd {
	return nil
}

func (v *HomeView) Update(msg tea.Msg) (View, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return v, tea.Quit

		case "enter":
			i, ok := v.list.SelectedItem()
			if ok {
				v.choice = string(i)

				switch v.choice {
				case string(CheckLatestVersion):
					return v, func() tea.Msg {
						return SwitchViewMsg{ViewID: VersionViewID}
					}
				case string(CheckLatestBuild):
					return v, func() tea.Msg {
						return SwitchViewMsg{ViewID: BuildViewID}
					}
				case string(CheckInstalledBuild):
					return v, func() tea.Msg {
						return SwitchViewMsg{ViewID: CurrentBuildViewID}
					}
				case string(DownloadLatestBuild):
					return v, func() tea.Msg {
						return SwitchViewMsg{ViewID: DownloadBuildID}
					}
				case string(Quit):
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
