package components

import (
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type KeyMap struct {
	Help   key.Binding
	Back   key.Binding
	Quit   key.Binding
	Custom []key.Binding // per-view bindings (e.g. retry, y/n)
}

func (k KeyMap) ShortHelp() []key.Binding {
	return append([]key.Binding{k.Back, k.Quit}, k.Custom...)
}

func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{append([]key.Binding{k.Back, k.Quit}, k.Custom...)}
}

var DefaultKeyMap = KeyMap{
	Help: key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "toggle help")),
	Back: key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "back")),
	Quit: key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
}

type Help struct {
	help help.Model
	keys KeyMap
}

func NewHelp(additionalKeys ...key.Binding) Help {
	keys := DefaultKeyMap
	keys.Custom = additionalKeys
	return Help{
		help: help.New(),
		keys: keys,
	}
}

func (h Help) View() string {
	return lipgloss.NewStyle().Padding(1, 0, 0, 2).Render(h.help.View(h.keys))
}

func (h Help) Update(msg tea.Msg) (Help, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, h.keys.Help):
			h.help.ShowAll = !h.help.ShowAll
		case key.Matches(msg, h.keys.Quit):
			return h, tea.Quit
		}
	}
	return h, nil
}
