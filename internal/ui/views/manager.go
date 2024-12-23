package views

import (
	tea "github.com/charmbracelet/bubbletea"
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
}

func NewManager() *Manager {
	return &Manager{}
}

func (m *Manager) Init() tea.Cmd {
	// Start with home view
	homeView := NewHomeView()
	m.currentView = homeView
	return m.currentView.Init()
}

func (m *Manager) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
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
		view = NewHomeView()
	case VersionViewID:
		view = NewVersionView()
	case BuildViewID:
		view = NewBuildView()
	case CurrentBuildViewID:
		view = NewCurrentBuildView()
	case DownloadBuildID:
		view = NewDownloadView()
	default:
		view = NewHomeView()
	}

	m.currentView = view
	return m, view.Init()
}
