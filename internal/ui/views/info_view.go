package views

import (
	"context"
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mbacalan/paper-mc-tui/internal/paper"
	"github.com/mbacalan/paper-mc-tui/internal/ui/components"
)

// checkTimeout bounds a "check latest" API round-trip.
const checkTimeout = 30 * time.Second

// latestMsg carries the result of an async CheckLatest call.
type latestMsg struct {
	info paper.LatestInfo
	err  error
}

// checkLatestCmd returns a command that resolves the latest release off the UI thread.
func checkLatestCmd(svc *paper.Service) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), checkTimeout)
		defer cancel()
		info, err := svc.CheckLatest(ctx)
		return latestMsg{info: info, err: err}
	}
}

// infoView fetches the latest release once and renders a single line from it. The
// "latest version" and "latest build" menu items differ only in their labels and how
// they render the result, so they share this implementation.
type infoView struct {
	svc        *paper.Service
	info       paper.LatestInfo
	loading    bool
	err        error
	loadingMsg string
	errLabel   string
	render     func(paper.LatestInfo) string
}

func NewVersionView(svc *paper.Service) *infoView {
	return &infoView{
		svc:        svc,
		loading:    true,
		loadingMsg: "Checking latest version…",
		errLabel:   "Unable to get latest version",
		render:     func(i paper.LatestInfo) string { return fmt.Sprintf("Latest available version is %s", i.Version) },
	}
}

func NewBuildView(svc *paper.Service) *infoView {
	return &infoView{
		svc:        svc,
		loading:    true,
		loadingMsg: "Checking latest build…",
		errLabel:   "Unable to get latest build",
		render: func(i paper.LatestInfo) string {
			return fmt.Sprintf("Latest available build is %d (%s)", i.Build, i.JarName)
		},
	}
}

func (v *infoView) Init() tea.Cmd {
	return checkLatestCmd(v.svc)
}

func (v *infoView) Update(msg tea.Msg) (View, tea.Cmd) {
	switch msg := msg.(type) {
	case latestMsg:
		v.loading = false
		v.info = msg.info
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

func (v *infoView) View() string {
	style := components.Body
	help := components.NewHelp()

	switch {
	case v.loading:
		return style.Render(v.loadingMsg) + help.View()
	case v.err != nil:
		return style.Render(fmt.Sprintf("%s:\n%v", v.errLabel, v.err)) + help.View()
	default:
		return style.Render(v.render(v.info)) + help.View()
	}
}
