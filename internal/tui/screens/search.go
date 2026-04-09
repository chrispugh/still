package screens

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/chrispugh/still/internal/config"
	"github.com/chrispugh/still/internal/journal"
	"github.com/chrispugh/still/internal/tui/styles"
)

// SearchModel provides full-text search across all entries.
type SearchModel struct {
	config  *config.Config
	store   *journal.Store
	width   int
	height  int
	input   textinput.Model
	entries []*journal.Entry
	results []*journal.Entry
	cursor  int
}

func NewSearch(cfg *config.Config, store *journal.Store) *SearchModel {
	ti := textinput.New()
	ti.Placeholder = "Search entries…"
	ti.CharLimit = 128
	ti.Focus()

	entries, _ := store.AllEntries()

	return &SearchModel{
		config:  cfg,
		store:   store,
		input:   ti,
		entries: entries,
	}
}

func (m *SearchModel) Init() tea.Cmd { return textinput.Blink }

func (m *SearchModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "ctrl+c":
			return m, navigate(ScreenHome)
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.results)-1 {
				m.cursor++
			}
		default:
			var cmd tea.Cmd
			m.input, cmd = m.input.Update(msg)
			m.search(m.input.Value())
			m.cursor = 0
			return m, cmd
		}
	}
	return m, nil
}

func (m *SearchModel) search(query string) {
	query = strings.ToLower(strings.TrimSpace(query))
	if query == "" {
		m.results = nil
		return
	}

	var results []*journal.Entry
	for _, e := range m.entries {
		if strings.Contains(strings.ToLower(e.Raw), query) ||
			strings.Contains(strings.ToLower(e.Polished), query) ||
			strings.Contains(strings.ToLower(strings.Join(e.Tags, " ")), query) {
			results = append(results, e)
		}
	}
	m.results = results
}

func (m *SearchModel) View() string {
	var sb strings.Builder

	sb.WriteString(styles.Center(m.width, styles.Title.Render("Search")))
	sb.WriteString("\n\n")

	inputW := m.width - 16
	if inputW > 72 {
		inputW = 72
	}
	sb.WriteString(styles.Center(m.width, styles.FocusedInput.Width(inputW).Render(m.input.View())))
	sb.WriteString("\n\n")

	if len(m.results) == 0 && m.input.Value() != "" {
		sb.WriteString(styles.Center(m.width, styles.Muted.Render("No entries found.")))
	}

	for i, e := range m.results {
		preview := e.Raw
		if len(preview) > 80 {
			preview = preview[:80] + "…"
		}
		date := e.Date.Format("Jan 2, 2006")
		line := fmt.Sprintf("%s — %s", date, preview)
		if i == m.cursor {
			line = styles.Selected.Render("▶ " + line)
		} else {
			line = styles.Muted.Render("  " + line)
		}
		sb.WriteString(styles.Center(m.width, line))
		sb.WriteString("\n")
	}

	sb.WriteString("\n")
	sb.WriteString(styles.Center(m.width, styles.Help.Render("type to search   ↑↓ navigate   esc back")))

	return sb.String()
}
