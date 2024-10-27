package components

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mbacalan/paper-mc-tui/internal/ui/styles"
)

type Item string

func (i Item) FilterValue() string { return "" }

type itemDelegate struct{}

func (d itemDelegate) Height() int                             { return 1 }
func (d itemDelegate) Spacing() int                            { return 0 }
func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	styles := styles.New().List
	i, ok := listItem.(Item)
	if !ok {
		return
	}

	str := fmt.Sprintf("%d. %s", index+1, i)

	fn := styles.Item.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return styles.SelectedItem.Render("> " + strings.Join(s, " "))
		}
	}

	fmt.Fprint(w, fn(str))
}

func NewItem(i string) Item {
	return Item(i)
}

// List is a reusable list component
type List struct {
	list   list.Model
	styles styles.DefaultStyles
}

func New(items []Item, styles styles.DefaultStyles) List {
	// Convert our items to list.Items
	listItems := make([]list.Item, len(items))
	for i, item := range items {
		listItems[i] = item
	}

	// Create new list
	l := list.New(listItems, itemDelegate{}, 20, 14) // width and height will be set in Update
	l.Styles.Title = styles.List.Title
	l.Styles.PaginationStyle = styles.List.Pagination
	l.Styles.HelpStyle = styles.List.Help
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)

	return List{
		list:   l,
		styles: styles,
	}
}

func (l *List) SetWidth(width int) {
	l.list.SetWidth(width)
}

// SetItems updates the list items
func (l *List) SetItems(items []Item) {
	listItems := make([]list.Item, len(items))
	for i, item := range items {
		listItems[i] = item
	}
	l.list.SetItems(listItems)
}

// SelectedItem returns the currently selected item
func (l *List) SelectedItem() (Item, bool) {
	selectedItem := l.list.SelectedItem()
	if selectedItem == nil {
		return "", false
	}
	return selectedItem.(Item), true
}

// Update handles list updates
func (l *List) Update(msg tea.Msg) (List, tea.Cmd) {
	var cmd tea.Cmd
	l.list, cmd = l.list.Update(msg)
	return *l, cmd
}

// View renders the list
func (l List) View() string {
	return l.list.View()
}

func (l List) Help() help.Model {
	return l.list.Help
}

// SetTitle sets the list title
func (l *List) SetTitle(title string) {
	l.list.Title = title
}

// SetShowTitle controls whether the title is displayed
func (l *List) SetShowTitle(show bool) {
	l.list.SetShowTitle(show)
}

// SetShowFilter controls whether the filter input is displayed
func (l *List) SetShowFilter(show bool) {
	l.list.SetShowFilter(show)
}

// SetFilteringEnabled controls whether filtering is enabled
func (l *List) SetFilteringEnabled(enabled bool) {
	l.list.SetFilteringEnabled(enabled)
}
