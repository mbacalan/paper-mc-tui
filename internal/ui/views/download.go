package views

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mbacalan/paper-mc-tui/internal/ui/components"
	"github.com/mbacalan/paper-mc-tui/internal/utils"
)

type backupState int

const (
	stateNormal backupState = iota
	stateBackupPrompt
	stateBackupInput
)

type DownloadView struct {
	success      bool
	error        error
	retryCount   int
	state        backupState
	backupInput  textinput.Model
	backupExists bool
	logger       *utils.Logger
	version      string
	build        string
}

func NewDownloadView() *DownloadView {
	ti := textinput.New()
	ti.Placeholder = "paper.backup.jar"
	ti.Focus()
	ti.CharLimit = 150
	ti.Width = 30

	logger, err := utils.NewLogger("paper.log", "version.txt")
	if err != nil {
		// Handle logger creation error
		fmt.Printf("Error creating logger: %v\n", err)
	}

	return &DownloadView{
		backupInput: ti,
		logger:      logger,
	}
}

func (v *DownloadView) backup(filename string) error {
	if filename == "" {
		filename = "paper.backup.jar"
	}

	// Actually move the file
	if err := os.Rename("paper.jar", filename); err != nil {
		return fmt.Errorf("failed to backup file: %w", err)
	}

	v.logger.Log(fmt.Sprintf("Backing up existing paper.jar as %s", filename))
	return nil
}

func (v *DownloadView) checkForExisting() bool {
	_, err := os.Stat("paper.jar")
	return err == nil
}

func (v *DownloadView) Download() error {
	version, err := utils.GetLatestStableVersion()
	if err != nil {
		return err
	}

	// Get build information
	build, err := utils.GetLatestBuild(version)
	if err != nil {
		return err
	}

	v.version = version
	v.build = build

	v.logger.Log(fmt.Sprintf("Latest version is %s", build))

	// Check if we already have the latest version
	lastVersion, err := v.logger.GetLastDownloadedVersion()
	if err == nil && lastVersion == build {
		return fmt.Errorf("already have the latest version (%s)", build)
	}

	v.logger.Log(fmt.Sprintf("Attempting to download %s...", build))

	err = utils.DownloadLatestBuild(version)
	if err != nil {
		v.logger.Log(fmt.Sprintf("Error downloading %s: %v", build, err))
		return err
	}

	// Save the new version information
	if err := v.logger.SaveDownloadedVersion(build); err != nil {
		return fmt.Errorf("failed to save version info: %w", err)
	}

	v.logger.Log(fmt.Sprintf("Download of %s successful!", build))
	return nil
}

func (v *DownloadView) Init() tea.Cmd {
	v.error = nil
	v.success = false
	v.state = stateNormal

	// First check if we need a new version at all
	version, err := utils.GetLatestStableVersion()
	if err != nil {
		v.error = err
		return nil
	}

	build, err := utils.GetLatestBuild(version)
	if err != nil {
		v.error = err
		return nil
	}

	lastVersion, err := v.logger.GetLastDownloadedVersion()
	if err == nil && lastVersion == build {
		v.error = fmt.Errorf("already have the latest version (%s)", build)
		return nil
	}

	// Then check if we need to backup
	v.backupExists = v.checkForExisting()
	if v.backupExists {
		v.state = stateBackupPrompt
		v.logger.Log("paper.jar already exists, prompting for backup...")
		return nil
	}

	// Finally proceed with download
	err = v.Download()
	if err != nil {
		v.error = err
		return nil
	}

	v.success = true
	return nil
}

func (v *DownloadView) Update(msg tea.Msg) (View, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case v.state == stateBackupInput:
			switch msg.String() {
			case "enter":
				filename := strings.TrimSpace(v.backupInput.Value())
				if err := v.backup(filename); err != nil {
					v.error = fmt.Errorf("failed to create backup: %w", err)
					v.state = stateNormal
					return v, nil
				}
				v.state = stateNormal
				return v, v.Init()
			case "esc":
				v.state = stateBackupPrompt
				return v, nil
			default:
				v.backupInput, cmd = v.backupInput.Update(msg)
				return v, cmd
			}

		case v.state == stateBackupPrompt:
			switch msg.String() {
			case "y":
				v.state = stateBackupInput
				v.backupInput.Focus()
				return v, nil
			case "n", "esc":
				v.logger.Log("Operation cancelled by user")
				return v, func() tea.Msg {
					return SwitchViewMsg{ViewID: HomeViewID}
				}
			case "q", "ctrl+c":
				return v, tea.Quit
			}

		default:
			switch msg.String() {
			case "q", "ctrl+c":
				return v, tea.Quit
			case "r":
				v.retryCount++
				return v, v.Init()
			case "esc":
				return v, func() tea.Msg {
					return SwitchViewMsg{ViewID: HomeViewID}
				}
			}
		}
	}

	return v, cmd
}

func (v *DownloadView) View() string {
	style := lipgloss.NewStyle().Margin(1, 2)
	var keys = components.KeyMap{
		Back: key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "back")),
		Quit: key.NewBinding(key.WithKeys("q"), key.WithHelp("q", "quit")),
	}
	help := components.NewHelp(keys)

	switch v.state {
	case stateBackupPrompt:
		promptText := style.Render("A paper.jar file already exists. Would you like to back it up? (y/n)\n\n")
		return promptText + help.View()

	case stateBackupInput:
		inputText := style.Render("Enter backup filename (default: paper.backup.jar):\n\n")
		return inputText + v.backupInput.View() + "\n\n(press Enter to confirm, Esc to go back)"

	default:
		if v.success {
			buildText := style.Render(fmt.Sprint("Downloaded latest build!\n\n"))
			return buildText + help.View()
		}

		if v.error != nil {
			retries := fmt.Sprintf("Retries: %d\n\n", v.retryCount)
			text := fmt.Sprint(v.error.Error() + "\n\n")

			if v.retryCount > 0 {
				text = fmt.Sprint(text + "Retrying...\n\n" + retries)
			}

			errorText := style.Render(text)

			var keys = components.KeyMap{
				Back:  key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "back")),
				Quit:  key.NewBinding(key.WithKeys("q"), key.WithHelp("q", "quit")),
				Retry: key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "retry")),
			}
			help := components.NewHelp(keys)

			return errorText + help.View()
		}

		return "Unable to download latest build!\n" + help.View()
	}
}
