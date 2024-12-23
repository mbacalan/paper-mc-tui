package components

import (
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type KeyMap interface {
	ShortHelp() []key.Binding
	FullHelp() [][]key.Binding
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
