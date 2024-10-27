package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mbacalan/paper-mc-tui/utils"
)

const url = "https://api.papermc.io/v2"

type model struct {
	list    list.Model
	choice  string
	version string
}

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

type statusMsg string

type errMsg struct{ err error }

const listHeight = 14

var (
	titleStyle        = lipgloss.NewStyle().MarginLeft(2)
	itemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
	paginationStyle   = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	helpStyle         = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)
	quitTextStyle     = lipgloss.NewStyle().Margin(1, 0, 2, 4)
)

type item string

func (i item) FilterValue() string { return "" }

type itemDelegate struct{}

func (d itemDelegate) Height() int                             { return 1 }
func (d itemDelegate) Spacing() int                            { return 0 }
func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(item)
	if !ok {
		return
	}

	str := fmt.Sprintf("%d. %s", index+1, i)

	fn := itemStyle.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return selectedItemStyle.Render("> " + strings.Join(s, " "))
		}
	}

	fmt.Fprint(w, fn(str))
}

func getLatestVersion() tea.Msg {
	versions, err := utils.FetchAPIData(url + "/projects/paper/")

	if err != nil {
		return errMsg{err}
	}

	latestVersion := versions.Versions[len(versions.Versions)-1]
	return statusMsg(latestVersion)
}

// For messages that contain errors it's often handy to also implement the
// error interface on the message.
func (e errMsg) Error() string { return e.err.Error() }

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		return m, nil

	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "q", "ctrl+c":
			return m, tea.Quit

		case "enter":
			i, ok := m.list.SelectedItem().(item)
			if ok {
				m.choice = string(i)

				if m.choice == "Check latest version" {
					// handle error?
					version, _ := getLatestVersion().(statusMsg)
					m.version = string(version)
				}
			}

			return m, nil

		case "esc":
			m.choice = ""
			m.version = ""
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m model) View() string {
	if m.version != "" {
		versionText := titleStyle.Render(fmt.Sprintf("Latest available version is %s\n\n", m.version))

		var keys = actionKeyMap{
			Esc:  key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "back")),
			Quit: key.NewBinding(key.WithKeys("q"), key.WithHelp("q", "quit")),
		}

		return versionText + m.list.Help.View(keys)
	}

	if m.choice == "Quit" {
		return quitTextStyle.Render("Bye!")
	}

	return "\n" + m.list.View()
}

func main() {
	items := []list.Item{
		item("Check latest version"),
		item("Quit"),
	}

	const defaultWidth = 20

	l := list.New(items, itemDelegate{}, defaultWidth, listHeight)
	l.Title = "PaperMC Management CLI"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.HelpStyle = helpStyle

	m := model{list: l}

	if _, err := tea.NewProgram(m).Run(); err != nil {
		fmt.Printf("Uh oh, there was an error: %v\n", err)
		os.Exit(1)
	}
}
