package views

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mbacalan/paper-mc-tui/internal/paper"
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
	svc         *paper.Service
	currentView View
}

func NewManager(svc *paper.Service) *Manager {
	return &Manager{svc: svc}
}

func (m *Manager) Init() tea.Cmd {
	m.currentView = NewHomeView()
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
		view = NewVersionView(m.svc)
	case BuildViewID:
		view = NewBuildView(m.svc)
	case CurrentBuildViewID:
		view = NewCurrentBuildView(m.svc)
	case DownloadBuildID:
		view = NewDownloadView(m.svc)
	default:
		view = NewHomeView()
	}

	m.currentView = view
	return m, view.Init()
}
