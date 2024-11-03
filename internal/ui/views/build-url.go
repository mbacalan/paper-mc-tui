package views

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mbacalan/paper-mc-tui/internal/ui/components"
	"github.com/mbacalan/paper-mc-tui/internal/ui/styles"
	"github.com/mbacalan/paper-mc-tui/internal/utils"
)

type BuildURLView struct {
	list       components.List
	styles     styles.DefaultStyles
	success    bool
	error      error
	retryCount int
}

func NewBuildURLView(s styles.DefaultStyles) *BuildURLView {
	return &BuildURLView{
		list:   components.New([]components.Item{}, s),
		styles: s,
	}
}

func (v *BuildURLView) Download() error {
	version, _ := utils.GetLatestStableVersion()
	err := utils.DownloadLatestBuild(version)

	if err != nil {
		return err
	}

	return nil
}

func (v *BuildURLView) Init() tea.Cmd {
	v.error = nil
	v.success = false

	err := v.Download()

	if err != nil {
		v.error = err
		return nil
	}

	v.success = true
	return nil
}

func (v *BuildURLView) Update(msg tea.Msg) (View, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return v, tea.Quit

		case "r":
			v.retryCount++
			return v, v.Init()

		case "esc":
			return v, func() tea.Msg {
				return SwitchViewMsg{ViewID: HomeViewID}
			}
		}
	}

	v.list, cmd = v.list.Update(msg)
	return v, cmd
}

func (v *BuildURLView) View() string {
	// return fmt.Sprintf(
	// 	"\nDebug: %s",
	// 	v.build,
	// )

	if v.success {
		buildText := v.styles.List.Title.Render(fmt.Sprint("Downloaded latest build!\n\n"))

		var keys = actionKeyMap{
			Esc:  key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "back")),
			Quit: key.NewBinding(key.WithKeys("q"), key.WithHelp("q", "quit")),
		}

		return buildText + v.list.Help().View(keys)
	}

	if v.error != nil {
		retries := fmt.Sprintf("Retries: %d\n\n", v.retryCount)
		text := fmt.Sprint(v.error.Error() + "\n\n")

		if v.retryCount > 0 {
			text = fmt.Sprint(text + "Retrying...\n\n" + retries)
		}

		errorText := v.styles.List.Title.Render(text)

		var keys = actionKeyMap{
			Esc:   key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "back")),
			Quit:  key.NewBinding(key.WithKeys("q"), key.WithHelp("q", "quit")),
			Retry: key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "retry")),
		}

		return errorText + v.list.Help().View(keys)
	}

	return "Unable to download latest build!\n"
}
