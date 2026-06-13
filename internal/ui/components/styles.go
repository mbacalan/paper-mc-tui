package components

import "github.com/charmbracelet/lipgloss"

// Accent is the app's highlight color, shared by the selected list item and the
// download progress bar so the UI has one accent.
const Accent = lipgloss.Color("170")

// Body is the standard margin applied to view content. lipgloss styles are immutable
// values, so this is safe to share read-only across views.
var Body = lipgloss.NewStyle().Margin(1, 2)
