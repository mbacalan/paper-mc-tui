package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mbacalan/paper-mc-tui/internal/buildinfo"
	"github.com/mbacalan/paper-mc-tui/internal/download"
	"github.com/mbacalan/paper-mc-tui/internal/paper"
	"github.com/mbacalan/paper-mc-tui/internal/papermc"
	"github.com/mbacalan/paper-mc-tui/internal/state"
	"github.com/mbacalan/paper-mc-tui/internal/ui/views"
)

func main() {
	showVersion := flag.Bool("version", false, "print version and exit")
	dir := flag.String("dir", envOr("PAPERMC_DIR", "."), "directory for paper.jar, backups, state and log")
	channel := flag.String("channel", envOr("PAPERMC_CHANNEL", "stable"), "release channel: stable|experimental")
	flag.Parse()

	if *showVersion {
		fmt.Printf("paper-mc-tui %s (commit %s, built %s)\n", buildinfo.Version, buildinfo.Commit, buildinfo.Date)
		return
	}

	channels, err := channelsFor(*channel)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(2)
	}

	userAgent := fmt.Sprintf("paper-mc-tui/%s (+https://github.com/mbacalan/paper-mc-tui)", buildinfo.Version)

	store, err := state.NewStore(*dir)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
	client := papermc.NewClient(papermc.WithUserAgent(userAgent))
	downloader := download.NewDownloader(download.WithUserAgent(userAgent))
	svc := paper.NewService(*dir, client, downloader, store, channels...)

	if _, err := tea.NewProgram(views.NewManager(svc)).Run(); err != nil {
		fmt.Printf("Uh oh, there was an error: %v\n", err)
		os.Exit(1)
	}
}

// channelsFor maps the --channel flag to the set of acceptable release channels.
func channelsFor(name string) ([]papermc.Channel, error) {
	switch name {
	case "stable":
		return []papermc.Channel{papermc.ChannelStable}, nil
	case "experimental":
		return []papermc.Channel{papermc.ChannelStable, papermc.ChannelBeta, papermc.ChannelAlpha}, nil
	default:
		return nil, fmt.Errorf("invalid channel %q (want stable or experimental)", name)
	}
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
