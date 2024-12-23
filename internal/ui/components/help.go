package components

import (
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// keyMap defines a set of keybindings. To work for help it must satisfy
// key.Map. It could also very easily be a map[string]key.Binding.
type KeyMap struct {
	Up      key.Binding
	Down    key.Binding
	Retry   key.Binding
	Accept  key.Binding
	Decline key.Binding
	Back    key.Binding
	Quit    key.Binding
}

// ShortHelp returns keybindings to be shown in the mini help view. It's part
// of the key.Map interface.
func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Retry, k.Back, k.Quit}
}

// FullHelp returns keybindings for the expanded help view. It's part of the
// key.Map interface.
func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Retry, k.Accept, k.Decline}, // first column
		{k.Back, k.Quit}, // second column
	}
}

type Help struct {
	help   help.Model
	keyMap KeyMap
}

func NewHelp(keyMap KeyMap) Help {
	return Help{
		help:   help.New(),
		keyMap: keyMap,
	}
}

func (h Help) View() string {
	return lipgloss.NewStyle().Padding(1, 0, 0, 2).Render(h.help.View(h.keyMap))
}

func (h Help) Update(msg tea.Msg) (Help, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		h.SetWidth(msg.Width)
	}

	h.help, cmd = h.help.Update(msg)
	return h, cmd
}

func (h *Help) SetWidth(width int) {
	h.help.Width = width
}
