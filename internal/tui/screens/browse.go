package screens

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/chrispugh/still/internal/config"
	"github.com/chrispugh/still/internal/journal"
	"github.com/chrispugh/still/internal/tui/styles"
)

type browseState int

const (
	bsList browseState = iota
	bsRead
)

// entryItem implements list.Item
type entryItem struct {
	entry *journal.Entry
}

func (i entryItem) Title() string {
	mood := journal.MoodEmoji(i.entry.Mood)
	words := fmt.Sprintf("%d words", journal.WordCount(i.entry.Raw))
	tags := ""
	if len(i.entry.Tags) > 0 {
		tags = "  " + strings.Join(i.entry.Tags, ", ")
	}
	return fmt.Sprintf("%s  %s%s  %s", i.entry.Date.Format("Mon Jan 2, 2006"), mood, tags, words)
}

func (i entryItem) Description() string {
	preview := i.entry.Raw
	if len(preview) > 80 {
		preview = preview[:80] + "…"
	}
	return preview
}

func (i entryItem) FilterValue() string {
	return i.entry.Date.Format("2006-01-02") + " " + strings.Join(i.entry.Tags, " ") + " " + i.entry.Raw
}

// BrowseModel lets users browse past entries.
type BrowseModel struct {
	config  *config.Config
	store   *journal.Store
	width   int
	height  int
	state   browseState
	list    list.Model
	entries []*journal.Entry
	viewing *journal.Entry
	rendered string
}

func NewBrowse(cfg *config.Config, store *journal.Store) *BrowseModel {
	l := list.New(nil, list.NewDefaultDelegate(), 80, 20)
	l.Title = "Past Entries"
	l.Styles.Title = styles.Title
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(true)

	m := &BrowseModel{
		config: cfg,
		store:  store,
		list:   l,
	}
	m.loadEntries()
	return m
}

func (m *BrowseModel) loadEntries() {
	entries, err := m.store.AllEntries()
	if err != nil {
		return
	}
	m.entries = entries

	items := make([]list.Item, len(entries))
	for i, e := range entries {
		items[i] = entryItem{entry: e}
	}
	m.list.SetItems(items)
}

func (m *BrowseModel) Init() tea.Cmd { return nil }

func (m *BrowseModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.SetSize(msg.Width, msg.Height-4)

	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "q":
			if m.state == bsRead {
				m.state = bsList
				return m, nil
			}
			return m, navigate(ScreenHome)
		case "enter":
			if m.state == bsList {
				if item, ok := m.list.SelectedItem().(entryItem); ok {
					m.viewing = item.entry
					m.rendered = m.renderEntry(item.entry)
					m.state = bsRead
					return m, nil
				}
			}
		}
	}

	if m.state == bsList {
		var cmd tea.Cmd
		m.list, cmd = m.list.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m *BrowseModel) renderEntry(e *journal.Entry) string {
	renderer, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(m.width-8),
	)
	if err != nil {
		return e.Raw
	}
	out, err := renderer.Render(e.ToMarkdown())
	if err != nil {
		return e.Raw
	}
	return out
}

func (m *BrowseModel) View() string {
	if m.state == bsList {
		header := styles.Center(m.width, styles.Title.Render("Browse Entries"))
		return header + "\n\n" + m.list.View() + "\n\n" +
			styles.Center(m.width, styles.Help.Render("↑↓ navigate   enter read   / filter   esc back"))
	}

	// Read mode
	if m.viewing == nil {
		return ""
	}

	var sb strings.Builder
	sb.WriteString(styles.Center(m.width, styles.Title.Render(m.viewing.Date.Format("Monday, January 2, 2006"))))
	sb.WriteString("\n\n")

	// Rendered markdown
	rendered := m.rendered
	if rendered == "" {
		rendered = m.viewing.Raw
	}

	contentBox := lipgloss.NewStyle().
		Width(m.width - 4).
		Padding(0, 2).
		Render(rendered)
	sb.WriteString(contentBox)
	sb.WriteString("\n")
	sb.WriteString(styles.Center(m.width, styles.Help.Render("esc  back to list")))

	return sb.String()
}
