package screens

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/chrispugh/still/internal/config"
	"github.com/chrispugh/still/internal/journal"
	"github.com/chrispugh/still/internal/tui/styles"
)

type menuItem struct {
	label string
	key   string
	name  ScreenName
}

var homeMenu = []menuItem{
	{label: "New Entry", key: "n", name: ScreenNewEntry},
	{label: "Browse", key: "b", name: ScreenBrowse},
	{label: "Search", key: "s", name: ScreenSearch},
	{label: "Stats", key: "t", name: ScreenStats},
	{label: "Settings", key: ",", name: ScreenSettings},
}

// HomeModel is the main menu screen.
type HomeModel struct {
	config        *config.Config
	store         *journal.Store
	width         int
	height        int
	cursor        int
	onThisDay     []*journal.Entry
	hasEntryToday bool
	streak        int
	totalEntries  int
}

func NewHome(cfg *config.Config, store *journal.Store) *HomeModel {
	m := &HomeModel{
		config: cfg,
		store:  store,
	}
	m.refresh()
	return m
}

func (m *HomeModel) refresh() {
	m.onThisDay, _ = m.store.OnThisDay()
	m.hasEntryToday = m.store.HasEntryToday()
	m.streak = m.store.Streak()
	m.totalEntries = m.store.TotalEntries()
}

func (m *HomeModel) Init() tea.Cmd { return nil }

func (m *HomeModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(homeMenu)-1 {
				m.cursor++
			}
		case "enter", " ":
			return m, m.selectItem(m.cursor)
		case "n":
			return m, navigate(ScreenNewEntry)
		case "b":
			return m, navigate(ScreenBrowse)
		case "s":
			return m, navigate(ScreenSearch)
		case "t":
			return m, navigate(ScreenStats)
		case ",":
			return m, navigate(ScreenSettings)
		case "q":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m *HomeModel) selectItem(idx int) tea.Cmd {
	if idx >= len(homeMenu) {
		return tea.Quit
	}
	return navigate(homeMenu[idx].name)
}

func navigate(screen ScreenName) tea.Cmd {
	return func() tea.Msg { return ChangeScreenMsg{Screen: screen} }
}

func (m *HomeModel) View() string {
	if m.width == 0 {
		return ""
	}

	var sb strings.Builder

	// Top padding
	topPad := (m.height - 20) / 2
	if topPad < 0 {
		topPad = 0
	}
	sb.WriteString(strings.Repeat("\n", topPad))

	// App name + tagline
	name := styles.AppName.Render("still")
	tagline := styles.Muted.Render("for people who think more than they write")
	sb.WriteString(styles.Center(m.width, name+"  "+tagline))
	sb.WriteString("\n\n")

	// Streak / stats bar
	sb.WriteString(styles.Center(m.width, m.statsLine()))
	sb.WriteString("\n\n")

	// On This Day banner
	if len(m.onThisDay) > 0 {
		sb.WriteString(m.onThisDayBanner())
		sb.WriteString("\n\n")
	}

	// Nudge if no entry yet today
	if !m.hasEntryToday {
		nudge := styles.Subtitle.Render("No entry yet today — what's on your mind?")
		sb.WriteString(styles.Center(m.width, nudge))
		sb.WriteString("\n\n")
	}

	// Menu
	sb.WriteString(m.menuView())
	sb.WriteString("\n\n")

	// Help
	help := styles.Help.Render("↑↓ / jk  navigate   enter  select   q  quit")
	sb.WriteString(styles.Center(m.width, help))

	return sb.String()
}

func (m *HomeModel) statsLine() string {
	date := styles.Muted.Render(time.Now().Format("Monday, January 2"))

	var streakStr string
	if m.streak > 0 {
		streakStr = styles.KeyHint.Render(fmt.Sprintf("🔥 %d", m.streak)) +
			styles.Muted.Render(" day streak")
	} else {
		streakStr = styles.Muted.Render("no streak yet")
	}

	total := styles.Muted.Render(fmt.Sprintf("%d entries", m.totalEntries))
	return date + "   " + streakStr + "   " + total
}

func (m *HomeModel) onThisDayBanner() string {
	entry := m.onThisDay[0]
	year := entry.Date.Year()
	preview := entry.Raw
	if len(preview) > 100 {
		preview = preview[:100] + "…"
	}

	content := styles.Title.Render(fmt.Sprintf("On this day, %d", year)) +
		"\n\n" + styles.Muted.Render(preview)

	maxW := m.width - 8
	if maxW > 72 {
		maxW = 72
	}
	banner := styles.Banner.Width(maxW).Render(content)
	return styles.Center(m.width, banner)
}

func (m *HomeModel) menuView() string {
	menuW := 24
	var rows []string
	for i, item := range homeMenu {
		var row string
		if i == m.cursor {
			row = styles.Selected.Render("▶  " + item.label)
		} else {
			row = styles.Body.Render("   " + item.label)
		}
		rows = append(rows, styles.Center(m.width,
			lipgloss.NewStyle().Width(menuW).Render(row),
		))
	}
	// Quit option
	quitRow := styles.Muted.Render("   Quit  q")
	rows = append(rows, styles.Center(m.width,
		lipgloss.NewStyle().Width(menuW).Render(quitRow),
	))
	return strings.Join(rows, "\n")
}
