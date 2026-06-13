package views

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mbacalan/paper-mc-tui/internal/paper"
	"github.com/mbacalan/paper-mc-tui/internal/state"
	"github.com/mbacalan/paper-mc-tui/internal/ui/components"
)

type installedMsg struct {
	state state.State
	err   error
}

type CurrentBuildView struct {
	svc       *paper.Service
	installed state.State
	loading   bool
	err       error
}

func NewCurrentBuildView(svc *paper.Service) *CurrentBuildView {
	return &CurrentBuildView{svc: svc, loading: true}
}

func (v *CurrentBuildView) Init() tea.Cmd {
	return func() tea.Msg {
		st, err := v.svc.Installed()
		return installedMsg{state: st, err: err}
	}
}

func (v *CurrentBuildView) Update(msg tea.Msg) (View, tea.Cmd) {
	switch msg := msg.(type) {
	case installedMsg:
		v.loading = false
		v.installed = msg.state
		v.err = msg.err

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return v, tea.Quit
		case "esc":
			return v, backToHome
		}
	}

	return v, nil
}

func (v *CurrentBuildView) View() string {
	style := components.Body
	help := components.NewHelp()

	switch {
	case v.loading:
		return style.Render("Reading installed build…") + help.View()
	case v.err != nil:
		return style.Render(fmt.Sprintf("Unable to read installed build:\n%v", v.err)) + help.View()
	case v.installed.Build == 0:
		return style.Render("No build has been installed by this tool yet.") + help.View()
	default:
		buildText := style.Render(fmt.Sprintf("Installed build is %d (%s)",
			v.installed.Build, v.installed.JarName))
		noteText := style.Render("Note: this is according to this tool's records.\nIt may be incorrect if you've updated manually.")
		return lipgloss.JoinVertical(lipgloss.Left, buildText, noteText) + help.View()
	}
}
