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

// StatsModel shows writing statistics and a mood heatmap.
type StatsModel struct {
	config        *config.Config
	store         *journal.Store
	width         int
	height        int
	streak        int
	longestStreak int
	total         int
	avgWords      int
	moodHistory   []struct {
		Date time.Time
		Mood int
	}
}

func NewStats(cfg *config.Config, store *journal.Store) *StatsModel {
	m := &StatsModel{
		config: cfg,
		store:  store,
	}
	m.streak = store.Streak()
	m.longestStreak = store.LongestStreak()
	m.total = store.TotalEntries()
	m.avgWords = store.AvgWordCount()
	m.moodHistory = store.MoodHistory(91) // ~3 months
	return m
}

func (m *StatsModel) Init() tea.Cmd { return nil }

func (m *StatsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		if msg.String() == "esc" || msg.String() == "q" {
			return m, navigate(ScreenHome)
		}
	}
	return m, nil
}

func (m *StatsModel) View() string {
	var sb strings.Builder

	sb.WriteString(styles.Center(m.width, styles.Title.Render("Stats")))
	sb.WriteString("\n\n")

	// Quick stats grid
	stats := []struct{ label, val string }{
		{"Current streak", fmt.Sprintf("%d days", m.streak)},
		{"Longest streak", fmt.Sprintf("%d days", m.longestStreak)},
		{"Total entries", fmt.Sprintf("%d", m.total)},
		{"Avg. word count", fmt.Sprintf("%d words", m.avgWords)},
	}

	maxW := 64
	colW := maxW / 2

	var rows []string
	for i := 0; i < len(stats); i += 2 {
		left := m.statCard(stats[i].label, stats[i].val, colW-2)
		right := ""
		if i+1 < len(stats) {
			right = m.statCard(stats[i+1].label, stats[i+1].val, colW-2)
		}
		row := lipgloss.JoinHorizontal(lipgloss.Top, left, "  ", right)
		rows = append(rows, styles.Center(m.width, row))
	}
	sb.WriteString(strings.Join(rows, "\n"))
	sb.WriteString("\n\n")

	// Mood heatmap
	sb.WriteString(styles.Center(m.width, styles.Subtitle.Render("Mood — last 13 weeks")))
	sb.WriteString("\n\n")
	sb.WriteString(styles.Center(m.width, m.moodHeatmap()))
	sb.WriteString("\n\n")

	sb.WriteString(styles.Center(m.width, styles.Help.Render("esc  back")))

	return sb.String()
}

func (m *StatsModel) statCard(label, val string, w int) string {
	content := styles.Muted.Render(label) + "\n" +
		styles.Title.Render(val)
	return styles.Box.Width(w).Render(content)
}

// moodHeatmap renders a GitHub-style grid. Each cell = one day.
// Colors: no entry = ·, mood 1–2 = dim, 3 = mid, 4–5 = bright
func (m *StatsModel) moodHeatmap() string {
	moodColor := func(mood int) lipgloss.Style {
		switch {
		case mood == 0:
			return lipgloss.NewStyle().Foreground(lipgloss.Color("#2A2A2A"))
		case mood <= 2:
			return lipgloss.NewStyle().Foreground(lipgloss.Color("#5A4A3A"))
		case mood == 3:
			return lipgloss.NewStyle().Foreground(lipgloss.Color("#9B7A4A"))
		default:
			return lipgloss.NewStyle().Foreground(lipgloss.Color("#D4A574"))
		}
	}

	cell := func(mood int) string {
		if mood == 0 {
			return moodColor(0).Render("·")
		}
		return moodColor(mood).Render("█")
	}

	// Arrange into weeks (columns), 7 rows (days)
	// moodHistory is ordered oldest-first by MoodHistory
	days := m.moodHistory
	weeks := len(days) / 7
	if weeks > 13 {
		weeks = 13
		days = days[len(days)-91:]
	}

	var lines [7]strings.Builder
	for i, d := range days {
		row := i % 7
		if i > 0 && i%7 == 0 {
			for r := range lines {
				lines[r].WriteString(" ")
			}
		}
		lines[row].WriteString(cell(d.Mood))
	}

	var result strings.Builder
	dayLabels := []string{"Mon", "   ", "Wed", "   ", "Fri", "   ", "Sun"}
	for r, line := range lines {
		result.WriteString(styles.Muted.Render(dayLabels[r]) + "  " + line.String())
		if r < 6 {
			result.WriteString("\n")
		}
	}
	return result.String()
}
