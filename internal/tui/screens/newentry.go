package screens

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/chrispugh/still/internal/ai"
	"github.com/chrispugh/still/internal/config"
	"github.com/chrispugh/still/internal/journal"
	"github.com/chrispugh/still/internal/tui/styles"
)

type newEntryState int

const (
	neWriting newEntryState = iota
	neMood
	neTags
	neAskPolish
	nePolishing
	nePolishReview
	neSaved
)

// Polish result messages
type polishDoneMsg struct{ text string }
type polishErrMsg struct{ err error }
type promptDoneMsg struct{ text string }

// NewEntryModel handles the full new-entry flow.
type NewEntryModel struct {
	config  *config.Config
	store   *journal.Store
	width   int
	height  int
	state   newEntryState
	err     string
	writing bool

	textarea  textarea.Model
	tagInput  textinput.Model
	spinner   spinner.Model
	aiClient  *ai.Client

	mood        int // 1–5; 0 = not set
	rawText     string
	polishedText string
	usePolished bool
	tags        []string
	prompt      string // writing prompt from AI
	showPrompt  bool
}

func NewNewEntry(cfg *config.Config, store *journal.Store) *NewEntryModel {
	ta := textarea.New()
	ta.Placeholder = "What happened today? Write freely…"
	ta.ShowLineNumbers = false
	ta.SetWidth(80)
	ta.SetHeight(20)
	ta.Focus()

	ti := textinput.New()
	ti.Placeholder = "work, travel, health  (comma-separated, or leave blank)"
	ti.CharLimit = 200

	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(styles.ColorPrimary)

	var aiClient *ai.Client
	if cfg.AI.Enabled {
		aiClient = ai.NewClient(cfg.AI.Model, cfg.AI.VoiceProfile)
	}

	return &NewEntryModel{
		config:   cfg,
		store:    store,
		state:    neWriting,
		textarea: ta,
		tagInput: ti,
		spinner:  sp,
		aiClient: aiClient,
		mood:     3,
	}
}

func (m *NewEntryModel) Init() tea.Cmd {
	return textarea.Blink
}

func (m *NewEntryModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.textarea.SetWidth(msg.Width - 8)
		m.textarea.SetHeight(msg.Height - 12)

	case polishDoneMsg:
		m.polishedText = msg.text
		m.state = nePolishReview

	case polishErrMsg:
		m.err = fmt.Sprintf("AI polish failed: %v", msg.err)
		m.state = neWriting // fall back

	case promptDoneMsg:
		m.prompt = msg.text
		m.showPrompt = true

	case tea.KeyMsg:
		return m.handleKey(msg)

	case spinner.TickMsg:
		if m.state == nePolishing {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
	}

	// Delegate to active input
	var cmd tea.Cmd
	switch m.state {
	case neWriting:
		m.textarea, cmd = m.textarea.Update(msg)
	case neTags:
		m.tagInput, cmd = m.tagInput.Update(msg)
	}
	return m, cmd
}

func (m *NewEntryModel) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	switch m.state {
	case neWriting:
		switch key {
		case "ctrl+s", "ctrl+d":
			raw := strings.TrimSpace(m.textarea.Value())
			if raw == "" {
				m.err = "Nothing to save — write something first."
				return m, nil
			}
			m.rawText = raw
			m.err = ""
			m.state = neMood
			return m, nil
		case "ctrl+p":
			// Generate prompt
			if m.aiClient != nil && !m.showPrompt {
				return m, m.fetchPromptCmd()
			}
		case "esc":
			return m, navigate(ScreenHome)
		}
		var cmd tea.Cmd
		m.textarea, cmd = m.textarea.Update(msg)
		return m, cmd

	case neMood:
		switch key {
		case "1", "2", "3", "4", "5":
			m.mood = int(key[0] - '0')
		case "left", "h":
			if m.mood > 1 {
				m.mood--
			}
		case "right", "l":
			if m.mood < 5 {
				m.mood++
			}
		case "enter":
			m.state = neTags
			m.tagInput.Focus()
			return m, textinput.Blink
		case "esc":
			m.state = neWriting
			return m, nil
		}

	case neTags:
		switch key {
		case "enter":
			m.parseTags()
			if m.config.AI.Enabled && m.aiClient != nil && ai.IsAvailable() {
				m.state = neAskPolish
			} else {
				return m, m.saveAndReturn()
			}
			return m, nil
		case "esc":
			m.state = neMood
			return m, nil
		default:
			var cmd tea.Cmd
			m.tagInput, cmd = m.tagInput.Update(msg)
			return m, cmd
		}

	case neAskPolish:
		switch key {
		case "y", "Y", "enter":
			m.state = nePolishing
			return m, tea.Batch(m.spinner.Tick, m.polishCmd())
		case "n", "N", "esc":
			return m, m.saveAndReturn()
		}

	case nePolishReview:
		switch key {
		case "1", "r":
			m.usePolished = false
			return m, m.saveAndReturn()
		case "2", "p":
			m.usePolished = true
			return m, m.saveAndReturn()
		case "esc":
			m.usePolished = false
			return m, m.saveAndReturn()
		}
	}

	return m, nil
}

func (m *NewEntryModel) parseTags() {
	raw := m.tagInput.Value()
	m.tags = nil
	for _, t := range strings.Split(raw, ",") {
		t = strings.TrimSpace(t)
		if t != "" {
			m.tags = append(m.tags, t)
		}
	}
}

func (m *NewEntryModel) polishCmd() tea.Cmd {
	rawText := m.rawText
	client := m.aiClient
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
		defer cancel()
		polished, err := client.Polish(ctx, rawText)
		if err != nil {
			return polishErrMsg{err}
		}
		return polishDoneMsg{polished}
	}
}

func (m *NewEntryModel) fetchPromptCmd() tea.Cmd {
	client := m.aiClient
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		prompt, err := client.GeneratePrompt(ctx, nil)
		if err != nil {
			return promptDoneMsg{"What's been on your mind lately that you haven't said out loud?"}
		}
		return promptDoneMsg{prompt}
	}
}

func (m *NewEntryModel) saveAndReturn() tea.Cmd {
	polished := ""
	if m.usePolished {
		polished = m.polishedText
	}

	entry := &journal.Entry{
		Date:     time.Now(),
		Mood:     m.mood,
		Tags:     m.tags,
		Raw:      m.rawText,
		Polished: polished,
	}

	_ = m.store.SaveEntry(entry) // TODO: surface save error

	return navigate(ScreenHome)
}

// ─── View ─────────────────────────────────────────────────────────────────────

func (m *NewEntryModel) View() string {
	switch m.state {
	case neWriting:
		return m.viewWriting()
	case neMood:
		return m.viewMood()
	case neTags:
		return m.viewTags()
	case neAskPolish:
		return m.viewAskPolish()
	case nePolishing:
		return m.viewPolishing()
	case nePolishReview:
		return m.viewPolishReview()
	default:
		return ""
	}
}

func (m *NewEntryModel) viewWriting() string {
	var sb strings.Builder

	date := time.Now().Format("Monday, January 2, 2006")
	sb.WriteString(styles.Pad(4, styles.Title.Render(date)))
	sb.WriteString("\n")

	if m.showPrompt && m.prompt != "" {
		prompt := styles.Muted.Render("💬 " + m.prompt)
		sb.WriteString(styles.Pad(4, prompt))
		sb.WriteString("\n")
	}

	sb.WriteString("\n")

	// Textarea
	taView := styles.FocusedInput.
		Width(m.width - 8).
		Render(m.textarea.View())
	sb.WriteString(styles.Pad(2, taView))

	sb.WriteString("\n\n")

	if m.err != "" {
		sb.WriteString(styles.Pad(4, styles.ErrorStyle.Render(m.err)))
		sb.WriteString("\n\n")
	}

	hints := styles.Help.Render("ctrl+s  save   ctrl+p  get a prompt   esc  cancel")
	sb.WriteString(styles.Pad(4, hints))

	return sb.String()
}

func (m *NewEntryModel) viewMood() string {
	var sb strings.Builder

	sb.WriteString(styles.Center(m.width, styles.Title.Render("How are you feeling?")))
	sb.WriteString("\n\n")

	emojis := []struct {
		n int
		e string
		l string
	}{
		{1, "😔", "rough"},
		{2, "😕", "meh"},
		{3, "😐", "okay"},
		{4, "🙂", "good"},
		{5, "😄", "great"},
	}

	var cells []string
	for _, e := range emojis {
		var cell string
		if e.n == m.mood {
			cell = styles.MoodActive.Render(fmt.Sprintf("%s  %d %s", e.e, e.n, e.l))
		} else {
			cell = styles.MoodInactive.Render(fmt.Sprintf("%s  %d %s", e.e, e.n, e.l))
		}
		cells = append(cells, cell)
	}
	sb.WriteString(styles.Center(m.width, strings.Join(cells, "   ")))
	sb.WriteString("\n\n")

	sb.WriteString(styles.Center(m.width, styles.Help.Render("1–5 or ←→ to choose   enter to continue   esc to go back")))

	return sb.String()
}

func (m *NewEntryModel) viewTags() string {
	var sb strings.Builder

	sb.WriteString(styles.Center(m.width, styles.Title.Render("Any tags?")))
	sb.WriteString("\n\n")

	w := m.width - 16
	if w < 40 {
		w = 40
	}
	if w > 72 {
		w = 72
	}
	inputView := styles.FocusedInput.Width(w).Render(m.tagInput.View())
	sb.WriteString(styles.Center(m.width, inputView))
	sb.WriteString("\n\n")

	sb.WriteString(styles.Center(m.width, styles.Help.Render("comma-separated   enter to continue   esc to go back")))

	return sb.String()
}

func (m *NewEntryModel) viewAskPolish() string {
	var sb strings.Builder

	sb.WriteString(styles.Center(m.width, styles.Title.Render("Polish with AI?")))
	sb.WriteString("\n\n")

	desc := styles.Muted.Render("A local AI will rewrite your entry in your voice.\nYour words stay on your device.")
	sb.WriteString(styles.Center(m.width, desc))
	sb.WriteString("\n\n")

	opts := styles.KeyHint.Render("y") + styles.Body.Render(" yes   ") +
		styles.KeyHint.Render("n") + styles.Body.Render(" skip")
	sb.WriteString(styles.Center(m.width, opts))

	return sb.String()
}

func (m *NewEntryModel) viewPolishing() string {
	line := m.spinner.View() + "  " + styles.Body.Render("Polishing your entry…")
	return lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Align(lipgloss.Center, lipgloss.Center).
		Render(line)
}

func (m *NewEntryModel) viewPolishReview() string {
	var sb strings.Builder

	sb.WriteString(styles.Center(m.width, styles.Title.Render("Compare versions")))
	sb.WriteString("\n\n")

	halfW := (m.width / 2) - 4

	rawBox := styles.Box.Width(halfW).Render(
		styles.Subtitle.Render("① Raw — your words") + "\n\n" +
			styles.Muted.Render(truncate(m.rawText, halfW*6)),
	)
	polishedBox := styles.Box.Width(halfW).Render(
		styles.Subtitle.Render("② Polished — AI rewrite") + "\n\n" +
			styles.Body.Render(truncate(m.polishedText, halfW*6)),
	)

	side := lipgloss.JoinHorizontal(lipgloss.Top, rawBox, "  ", polishedBox)
	sb.WriteString(styles.Center(m.width, side))
	sb.WriteString("\n\n")

	help := styles.Help.Render("1 / r  keep raw   2 / p  use polished")
	sb.WriteString(styles.Center(m.width, help))

	return sb.String()
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "…"
}
