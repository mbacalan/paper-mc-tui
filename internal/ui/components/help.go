package components

import (
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type KeyMap struct {
	Up      key.Binding
	Down    key.Binding
	Retry   key.Binding
	Accept  key.Binding
	Decline key.Binding
	Help    key.Binding
	Back    key.Binding
	Quit    key.Binding
	Custom  []key.Binding
}

func (k KeyMap) ShortHelp() []key.Binding {
	defaultKeys := []key.Binding{k.Accept, k.Decline, k.Retry, k.Back, k.Quit}
	return append(defaultKeys, k.Custom...)
}

func (k KeyMap) FullHelp() [][]key.Binding {
	defaultKeys := [][]key.Binding{
		{k.Up, k.Down, k.Retry, k.Accept, k.Decline},
		{k.Back, k.Quit},
	}
	if len(k.Custom) > 0 {
		defaultKeys = append(defaultKeys, k.Custom)
	}
	return defaultKeys
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
	case tea.WindowSizeMsg:
		h.help.Width = msg.Width
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
