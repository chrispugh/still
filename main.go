package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/chrispugh/still/internal/config"
	"github.com/chrispugh/still/internal/tui"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "still: failed to load config: %v\n", err)
		os.Exit(1)
	}

	app := tui.New(cfg)
	p := tea.NewProgram(app, tea.WithAltScreen(), tea.WithMouseCellMotion())

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "still: %v\n", err)
		os.Exit(1)
	}
}
