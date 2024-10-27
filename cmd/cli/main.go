package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mbacalan/paper-mc-tui/internal/ui/components"
	"github.com/mbacalan/paper-mc-tui/internal/ui/styles"
	"github.com/mbacalan/paper-mc-tui/internal/utils"
)

const url = "https://api.papermc.io/v2"

type model struct {
	list    components.List
	styles  styles.DefaultStyles
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
			i, ok := m.list.SelectedItem()
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
		versionText := m.styles.List.Title.Render(fmt.Sprintf("Latest available version is %s\n\n", m.version))

		var keys = actionKeyMap{
			Esc:  key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "back")),
			Quit: key.NewBinding(key.WithKeys("q"), key.WithHelp("q", "quit")),
		}

		return versionText + m.list.Help().View(keys)
	}

	if m.choice == "Quit" {
		return m.styles.General.QuitText.Render("Bye!")
	}

	return "\n" + m.list.View()
}

func main() {
	s := styles.New()
	items := []components.Item{
		"Check latest version",
		"Quit",
	}

	l := components.New(items, s)
	l.SetTitle("PaperMC Management CLI")

	m := model{list: l, styles: s}

	if _, err := tea.NewProgram(m).Run(); err != nil {
		fmt.Printf("Uh oh, there was an error: %v\n", err)
		os.Exit(1)
	}
}
