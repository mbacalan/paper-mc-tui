package components

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Item string
type itemDelegate struct{}
type List struct {
	list list.Model
}

func (i Item) FilterValue() string { return "" }

func (d itemDelegate) Height() int                             { return 1 }
func (d itemDelegate) Spacing() int                            { return 0 }
func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(Item)
	if !ok {
		return
	}

	str := fmt.Sprintf("%d. %s", index+1, i)

	fn := lipgloss.NewStyle().PaddingLeft(4).Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170")).Render("> " + strings.Join(s, " "))
		}
	}

	fmt.Fprint(w, fn(str))
}

func NewList(items []Item, title string) List {
	listItems := make([]list.Item, len(items))
	for i, item := range items {
		listItems[i] = item
	}

	l := list.New(listItems, itemDelegate{}, 28, 12)
	l.Title = title
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)

	return List{list: l}
}

func (l List) SelectedItem() (Item, bool) {
	selectedItem := l.list.SelectedItem()
	if selectedItem == nil {
		return "", false
	}
	return selectedItem.(Item), true
}

func (l List) Update(msg tea.Msg) (List, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		l.list.SetWidth(msg.Width)
	}

	l.list, cmd = l.list.Update(msg)
	return l, cmd
}

func (l List) View() string {
	return l.list.View()
}

func (l List) SetTitle(title string) {
	l.list.Title = title
}

func (l List) SetWidth(width int) {
	l.list.SetWidth(width)
}
