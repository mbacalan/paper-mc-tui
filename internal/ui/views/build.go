package views

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mbacalan/paper-mc-tui/internal/ui/components"
	"github.com/mbacalan/paper-mc-tui/internal/ui/styles"
	"github.com/mbacalan/paper-mc-tui/internal/utils"
)

type BuildView struct {
	list    components.List
	styles  styles.DefaultStyles
	version string
	build   string
}

func NewBuildView(s styles.DefaultStyles) *BuildView {
	return &BuildView{
		list:   components.New([]components.Item{}, s),
		styles: s,
	}
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

	v.list, cmd = v.list.Update(msg)
	return v, cmd
}

func (v *BuildView) View() string {
	// return fmt.Sprintf(
	// 	"\nDebug: %s",
	// 	v.build,
	// )

	if v.build != "" {
		buildText := v.styles.List.Title.Render(fmt.Sprintf("Latest available build is %s\n\n", v.build))

		var keys = actionKeyMap{
			Esc:  key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "back")),
			Quit: key.NewBinding(key.WithKeys("q"), key.WithHelp("q", "quit")),
		}

		return buildText + v.list.Help().View(keys)
	}

	return "Unable to get latest build!\n"
}
