package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/chrispugh/still/internal/config"
	"github.com/chrispugh/still/internal/journal"
	"github.com/chrispugh/still/internal/tui/screens"
)

// App is the root BubbleTea model. It owns the current screen and routes
// ChangeScreenMsg to swap screens.
type App struct {
	config *config.Config
	store  *journal.Store
	width  int
	height int
	model  tea.Model // current screen
}

// New creates the App. The first screen is chosen based on whether this is a
// first run.
func New(cfg *config.Config) *App {
	store := journal.NewStore(cfg.JournalPath)
	return &App{config: cfg, store: store}
}

func (a *App) Init() tea.Cmd {
	if a.config.IsFirstRun {
		a.model = screens.NewOnboarding(a.config)
	} else {
		a.model = screens.NewHome(a.config, a.store)
	}
	return a.model.Init()
}

func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height

	case tea.KeyMsg:
		// ctrl+c always exits regardless of screen
		if msg.String() == "ctrl+c" {
			return a, tea.Quit
		}

	case screens.ChangeScreenMsg:
		return a.switchTo(msg.Screen)
	}

	// Delegate all other messages to the current screen.
	newModel, cmd := a.model.Update(msg)
	a.model = newModel
	return a, cmd
}

func (a *App) View() string {
	if a.model == nil {
		return ""
	}
	return a.model.View()
}

func (a *App) switchTo(name screens.ScreenName) (tea.Model, tea.Cmd) {
	var next tea.Model

	switch name {
	case screens.ScreenOnboarding:
		next = screens.NewOnboarding(a.config)
	case screens.ScreenHome:
		// Reload store in case config changed during onboarding
		a.store = journal.NewStore(a.config.JournalPath)
		next = screens.NewHome(a.config, a.store)
	case screens.ScreenNewEntry:
		next = screens.NewNewEntry(a.config, a.store)
	case screens.ScreenBrowse:
		next = screens.NewBrowse(a.config, a.store)
	case screens.ScreenSearch:
		next = screens.NewSearch(a.config, a.store)
	case screens.ScreenStats:
		next = screens.NewStats(a.config, a.store)
	case screens.ScreenSettings:
		next = screens.NewSettings(a.config)
	default:
		return a, nil
	}

	a.model = next
	initCmd := next.Init()

	// Immediately deliver the current window size to the new screen.
	sized, sizeCmd := a.model.Update(tea.WindowSizeMsg{Width: a.width, Height: a.height})
	a.model = sized

	return a, tea.Batch(initCmd, sizeCmd)
}
