package views

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mbacalan/paper-mc-tui/internal/ui/styles"
)

// View is the interface that all views must implement
type View interface {
	Init() tea.Cmd
	Update(msg tea.Msg) (View, tea.Cmd)
	View() string
}

// ViewID represents different views in the application
type ViewID int

// Manager handles view switching and rendering
type Manager struct {
	currentView View
	styles      styles.DefaultStyles
}

func NewManager(s styles.DefaultStyles) *Manager {
	return &Manager{
		styles: s,
	}
}

func (m *Manager) Init() tea.Cmd {
	// Start with home view
	homeView := NewHomeView(m.styles)
	m.currentView = homeView
	return m.currentView.Init()
}

func (m *Manager) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.styles.Width = msg.Width
		return m, nil

	case SwitchViewMsg:
		return m.switchView(msg.ViewID)
	}

	m.currentView, cmd = m.currentView.Update(msg)
	return m, cmd
}

func (m *Manager) View() string {
	return m.currentView.View()
}

// SwitchViewMsg is used to switch between views
type SwitchViewMsg struct {
	ViewID ViewID
}

func (m *Manager) switchView(id ViewID) (tea.Model, tea.Cmd) {
	var view View

	switch id {
	case HomeViewID:
		view = NewHomeView(m.styles)
	case VersionViewID:
		view = NewVersionView(m.styles)
	case BuildViewID:
		view = NewBuildView(m.styles)
	case DownloadBuildID:
		view = NewDownloadView(m.styles)
	default:
		view = NewHomeView(m.styles)
	}

	m.currentView = view
	return m, view.Init()
}
