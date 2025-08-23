package views

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mbacalan/paper-mc-tui/internal/ui/components"
	"strconv"
)

type HomeView struct {
	list   components.List
	choice string
	items  []components.Item
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
		list:  list,
		items: items,
	}
}

func (v *HomeView) Init() tea.Cmd {
	return nil
}

func (v *HomeView) handleMenuSelection(choice string) tea.Cmd {
	switch choice {
	case string(CheckLatestVersion):
		return func() tea.Msg {
			return SwitchViewMsg{ViewID: VersionViewID}
		}
	case string(CheckLatestBuild):
		return func() tea.Msg {
			return SwitchViewMsg{ViewID: BuildViewID}
		}
	case string(CheckInstalledBuild):
		return func() tea.Msg {
			return SwitchViewMsg{ViewID: CurrentBuildViewID}
		}
	case string(DownloadLatestBuild):
		return func() tea.Msg {
			return SwitchViewMsg{ViewID: DownloadBuildID}
		}
	case string(Quit):
		return tea.Quit
	}
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
				return v, v.handleMenuSelection(v.choice)
			}

		default:
			if key := msg.String(); len(key) == 1 && key >= "1" && key <= "9" {
				if num, err := strconv.Atoi(key); err == nil {
					index := num - 1 // Convert to 0-based index
					if index >= 0 && index < len(v.items) {
						v.choice = string(v.items[index])
						return v, v.handleMenuSelection(v.choice)
					}
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
