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

type MenuAction string

const (
	CheckVersion  MenuAction = "Check latest version"
	CheckBuild    MenuAction = "Check latest build"
	DownloadBuild MenuAction = "Download latest build"
	Quit          MenuAction = "Quit"
)

const (
	HomeViewID ViewID = iota
	VersionViewID
	BuildViewID
	DownloadBuildID
)

var menuActionViewMap = map[MenuAction]ViewID{
	CheckVersion:  VersionViewID,
	CheckBuild:    BuildViewID,
	DownloadBuild: DownloadBuildID,
	Quit:          HomeViewID,
}

func NewHomeView(s styles.DefaultStyles) *HomeView {
	items := []components.Item{
		components.NewItem(string(CheckVersion)),
		components.NewItem(string(CheckBuild)),
		components.NewItem(string(DownloadBuild)),
		components.NewItem(string(Quit)),
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
				case string(CheckVersion):
					return v, func() tea.Msg {
						return SwitchViewMsg{ViewID: VersionViewID}
					}
				case string(CheckBuild):
					return v, func() tea.Msg {
						return SwitchViewMsg{ViewID: BuildViewID}
					}
				case string(DownloadBuild):
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
