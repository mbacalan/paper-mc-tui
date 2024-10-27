package styles

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"
)

type List struct {
	Title        lipgloss.Style
	Item         lipgloss.Style
	SelectedItem lipgloss.Style
	Pagination   lipgloss.Style
	Help         lipgloss.Style
}

// General contains styles used across multiple components
type General struct {
	QuitText lipgloss.Style
}

// DefaultStyles returns all default styles used in the application
type DefaultStyles struct {
	List    List
	General General
	Width   int
}

func New() DefaultStyles {
	return DefaultStyles{
		List: List{
			Title:        lipgloss.NewStyle().MarginLeft(2),
			Item:         lipgloss.NewStyle().PaddingLeft(4),
			SelectedItem: lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170")),
			Help:         list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1),
		},
		General: General{
			QuitText: lipgloss.NewStyle().Margin(1, 0, 2, 4),
		},
	}
}
