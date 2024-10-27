package views

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mbacalan/paper-mc-tui/internal/ui/components"
	"github.com/mbacalan/paper-mc-tui/internal/ui/styles"
	"github.com/mbacalan/paper-mc-tui/internal/utils"
)

type VersionView struct {
	list    components.List
	styles  styles.DefaultStyles
	version string
}

type statusMsg string

type errMsg struct{ err error }

// For messages that contain errors it's often handy to also implement the
// error interface on the message.
func (e errMsg) Error() string { return e.err.Error() }

type actionKeyMap struct {
	Esc  key.Binding
	Quit key.Binding
}

func (k actionKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Esc, k.Quit}
}

func (k actionKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{{k.Esc, k.Quit}}
}

const url = "https://api.papermc.io/v2"

func (v *VersionView) getLatestVersion() tea.Msg {
	versions, err := utils.FetchAPIData(url + "/projects/paper/")

	if err != nil {
		return errMsg{err}
	}

	latestVersion := versions.Versions[len(versions.Versions)-1]
	return statusMsg(latestVersion)
}

func NewVersionView(s styles.DefaultStyles) *VersionView {
	return &VersionView{
		list:   components.New([]components.Item{}, s),
		styles: s,
	}
}

func (v *VersionView) Init() tea.Cmd {
	version := v.getLatestVersion().(statusMsg)
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

	v.list, cmd = v.list.Update(msg)
	return v, cmd
}

func (v *VersionView) View() string {
	if v.version != "" {
		versionText := v.styles.List.Title.Render(fmt.Sprintf("Latest available version is %s\n\n", v.version))

		var keys = actionKeyMap{
			Esc:  key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "back")),
			Quit: key.NewBinding(key.WithKeys("q"), key.WithHelp("q", "quit")),
		}

		return versionText + v.list.Help().View(keys)
	}

	return "Unable to get latest version!\n"
}