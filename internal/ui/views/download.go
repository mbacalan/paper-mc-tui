package views

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mbacalan/paper-mc-tui/internal/paper"
	"github.com/mbacalan/paper-mc-tui/internal/ui/components"
)

// downloadTimeout bounds the jar transfer (separate from the short API timeout).
const downloadTimeout = 15 * time.Minute

type downloadState int

const (
	stateLoading downloadState = iota
	stateUpToDate
	stateBackupPrompt
	stateBackupInput
	stateDownloading
	stateDone
	stateError
)

// prepareMsg is the result of the initial "what's latest / do we need it" check.
type prepareMsg struct {
	info      paper.LatestInfo
	jarExists bool
	err       error
}

type progressMsg float64

type doneMsg struct{ err error }

type DownloadView struct {
	svc         *paper.Service
	state       downloadState
	info        paper.LatestInfo
	err         error
	backupInput textinput.Model
	progress    progress.Model

	// progress plumbing: the download runs in a goroutine that reports on these.
	progressCh chan float64
	doneCh     chan error
}

func NewDownloadView(svc *paper.Service) *DownloadView {
	ti := textinput.New()
	ti.Placeholder = "paper.backup.jar"
	ti.Focus()
	ti.CharLimit = 150
	ti.Width = 30

	return &DownloadView{
		svc:         svc,
		state:       stateLoading,
		backupInput: ti,
		progress:    progress.New(progress.WithSolidFill(string(components.Accent)), progress.WithWidth(40)),
	}
}

func (v *DownloadView) Init() tea.Cmd {
	v.state = stateLoading
	v.err = nil
	svc := v.svc
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), checkTimeout)
		defer cancel()
		info, err := svc.CheckLatest(ctx)
		if err != nil {
			return prepareMsg{err: err}
		}
		return prepareMsg{info: info, jarExists: svc.JarExists()}
	}
}

// startDownload launches the transfer in a goroutine and begins listening for progress.
func (v *DownloadView) startDownload() tea.Cmd {
	v.state = stateDownloading
	v.progressCh = make(chan float64)
	v.doneCh = make(chan error, 1)

	svc := v.svc
	progressCh := v.progressCh
	doneCh := v.doneCh
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), downloadTimeout)
		defer cancel()
		err := svc.Install(ctx, func(done, total int64) {
			if total <= 0 {
				return
			}
			select {
			case progressCh <- float64(done) / float64(total):
			default: // UI busy; drop this tick
			}
		})
		doneCh <- err
	}()

	return tea.Batch(v.progress.SetPercent(0), v.waitForActivity())
}

// waitForActivity blocks (off the UI thread) for the next progress tick or completion.
func (v *DownloadView) waitForActivity() tea.Cmd {
	progressCh := v.progressCh
	doneCh := v.doneCh
	return func() tea.Msg {
		select {
		case p := <-progressCh:
			return progressMsg(p)
		case err := <-doneCh:
			return doneMsg{err: err}
		}
	}
}

func (v *DownloadView) Update(msg tea.Msg) (View, tea.Cmd) {
	switch msg := msg.(type) {
	case prepareMsg:
		if msg.err != nil {
			v.state = stateError
			v.err = msg.err
			return v, nil
		}
		v.info = msg.info
		switch {
		case msg.info.UpToDate:
			v.state = stateUpToDate
			return v, nil
		case msg.jarExists:
			v.state = stateBackupPrompt
			return v, nil
		default:
			return v, v.startDownload()
		}

	case progressMsg:
		cmd := v.progress.SetPercent(float64(msg))
		return v, tea.Batch(cmd, v.waitForActivity())

	case doneMsg:
		if msg.err != nil {
			v.state = stateError
			v.err = msg.err
			return v, nil
		}
		v.state = stateDone
		return v, v.progress.SetPercent(1.0)

	case progress.FrameMsg:
		m, cmd := v.progress.Update(msg)
		v.progress = m.(progress.Model)
		return v, cmd

	case tea.KeyMsg:
		return v.handleKey(msg)
	}

	return v, nil
}

func (v *DownloadView) handleKey(msg tea.KeyMsg) (View, tea.Cmd) {
	// Quit works from any state.
	if msg.String() == "ctrl+c" {
		return v, tea.Quit
	}

	switch v.state {
	case stateBackupInput:
		switch msg.String() {
		case "enter":
			filename := strings.TrimSpace(v.backupInput.Value())
			if _, err := v.svc.Backup(filename); err != nil {
				v.state = stateError
				v.err = fmt.Errorf("failed to create backup: %w", err)
				return v, nil
			}
			return v, v.startDownload()
		case "esc":
			v.state = stateBackupPrompt
			return v, nil
		default:
			var cmd tea.Cmd
			v.backupInput, cmd = v.backupInput.Update(msg)
			return v, cmd
		}

	case stateBackupPrompt:
		switch msg.String() {
		case "y":
			v.state = stateBackupInput
			v.backupInput.Focus()
			return v, nil
		case "n", "esc":
			return v, backToHome
		}

	case stateError:
		switch msg.String() {
		case "q":
			return v, tea.Quit
		case "r":
			return v, v.Init()
		case "esc":
			return v, backToHome
		}

	default: // stateLoading, stateUpToDate, stateDownloading, stateDone
		switch msg.String() {
		case "q":
			return v, tea.Quit
		case "esc":
			if v.state != stateDownloading {
				return v, backToHome
			}
		}
	}

	return v, nil
}

func (v *DownloadView) View() string {
	style := components.Body

	switch v.state {
	case stateLoading:
		return style.Render("Checking for the latest build…") + components.NewHelp().View()

	case stateUpToDate:
		text := fmt.Sprintf("You already have the latest build %d (%s). Nothing to do.",
			v.info.Build, v.info.JarName)
		return style.Render(text) + components.NewHelp().View()

	case stateBackupPrompt:
		help := components.NewHelp(
			key.NewBinding(key.WithKeys("y"), key.WithHelp("y", "yes")),
			key.NewBinding(key.WithKeys("n"), key.WithHelp("n", "no")),
		)
		return style.Render("A paper.jar already exists. Back it up first? (y/n)") + help.View()

	case stateBackupInput:
		text := style.Render("Enter backup filename (default: paper.backup.jar):")
		return text + "\n" + v.backupInput.View() + "\n\n(press Enter to confirm, Esc to go back)"

	case stateDownloading:
		header := style.Render(fmt.Sprintf("Downloading %s (%s)…", v.info.JarName, humanMB(v.info.Download.Size)))
		return header + "\n" + lipgloss.NewStyle().Margin(0, 2).Render(v.progress.View()) + "\n"

	case stateDone:
		return style.Render(fmt.Sprintf("Downloaded and verified %s!", v.info.JarName)) + components.NewHelp().View()

	case stateError:
		help := components.NewHelp(key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "retry")))
		return style.Render(fmt.Sprintf("Download failed:\n%v", v.err)) + help.View()

	default:
		return style.Render("Unexpected state.") + components.NewHelp().View()
	}
}

// humanMB renders a byte count as megabytes.
func humanMB(b int64) string {
	return fmt.Sprintf("%.1f MB", float64(b)/(1024*1024))
}
