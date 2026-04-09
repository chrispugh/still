package screens

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/chrispugh/still/internal/ai"
	"github.com/chrispugh/still/internal/config"
	"github.com/chrispugh/still/internal/tui/styles"
)

type settingItem int

const (
	sAIEnabled settingItem = iota
	sModel
	sVoiceProfile
	sJournalPath
	sNudge
	sNudgeTime
	sCount
)

// SettingsModel lets users configure the app.
type SettingsModel struct {
	config    *config.Config
	width     int
	height    int
	cursor    settingItem
	editing   bool
	input     textinput.Model
	models    []string
	modelIdx  int
	err       string
}

func NewSettings(cfg *config.Config) *SettingsModel {
	ti := textinput.New()
	ti.CharLimit = 256

	models, _ := ai.ListModels()
	modelIdx := 0
	for i, m := range models {
		if m == cfg.AI.Model {
			modelIdx = i
			break
		}
	}

	return &SettingsModel{
		config:   cfg,
		input:    ti,
		models:   models,
		modelIdx: modelIdx,
	}
}

func (m *SettingsModel) Init() tea.Cmd { return nil }

func (m *SettingsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		if m.editing {
			return m.handleEditing(msg)
		}
		return m.handleNav(msg)
	}

	if m.editing {
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m *SettingsModel) handleNav(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < sCount-1 {
			m.cursor++
		}
	case "enter", " ":
		return m.activate()
	case "esc", "q":
		return m, navigate(ScreenHome)
	}
	return m, nil
}

func (m *SettingsModel) handleEditing(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		m.applyEdit()
		m.editing = false
		_ = m.config.Save()
	case "esc":
		m.editing = false
	default:
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m *SettingsModel) activate() (tea.Model, tea.Cmd) {
	switch m.cursor {
	case sAIEnabled:
		m.config.AI.Enabled = !m.config.AI.Enabled
		_ = m.config.Save()
	case sModel:
		if len(m.models) > 0 {
			m.modelIdx = (m.modelIdx + 1) % len(m.models)
			m.config.AI.Model = m.models[m.modelIdx]
			_ = m.config.Save()
		}
	case sVoiceProfile:
		m.input.SetValue(m.config.AI.VoiceProfile)
		m.input.Focus()
		m.editing = true
		return m, textinput.Blink
	case sJournalPath:
		m.input.SetValue(m.config.JournalPath)
		m.input.Focus()
		m.editing = true
		return m, textinput.Blink
	case sNudge:
		m.config.Notifications.DailyNudge = !m.config.Notifications.DailyNudge
		_ = m.config.Save()
	case sNudgeTime:
		m.input.SetValue(m.config.Notifications.NudgeTime)
		m.input.Focus()
		m.editing = true
		return m, textinput.Blink
	}
	return m, nil
}

func (m *SettingsModel) applyEdit() {
	v := strings.TrimSpace(m.input.Value())
	switch m.cursor {
	case sVoiceProfile:
		if v != "" {
			m.config.AI.VoiceProfile = v
		}
	case sJournalPath:
		if v != "" {
			m.config.JournalPath = config.ExpandPath(v)
		}
	case sNudgeTime:
		if v != "" {
			m.config.Notifications.NudgeTime = v
		}
	}
}

func (m *SettingsModel) View() string {
	var sb strings.Builder

	sb.WriteString(styles.Center(m.width, styles.Title.Render("Settings")))
	sb.WriteString("\n\n")

	rows := []struct {
		label string
		value string
	}{
		{"AI polish", boolStr(m.config.AI.Enabled)},
		{"Model", m.config.AI.Model},
		{"Voice profile", truncate(m.config.AI.VoiceProfile, 40)},
		{"Journal path", m.config.JournalPath},
		{"Daily nudge", boolStr(m.config.Notifications.DailyNudge)},
		{"Nudge time", m.config.Notifications.NudgeTime},
	}

	labelW := 18
	valueW := 42

	for i, row := range rows {
		label := fmt.Sprintf("%-*s", labelW, row.label)
		value := row.value

		var line string
		if settingItem(i) == m.cursor {
			if m.editing {
				editBox := styles.FocusedInput.Width(valueW).Render(m.input.View())
				line = styles.Selected.Render("▶ "+label) + "  " + editBox
			} else {
				line = styles.Selected.Render("▶ "+label) + "  " + styles.KeyHint.Render(value)
			}
		} else {
			line = styles.Muted.Render("  "+label) + "  " + styles.Body.Render(value)
		}

		sb.WriteString(styles.Center(m.width, line))
		sb.WriteString("\n")
	}

	sb.WriteString("\n")
	sb.WriteString(styles.Center(m.width, styles.Help.Render("↑↓ navigate   enter toggle/edit   esc back")))

	return sb.String()
}

func boolStr(b bool) string {
	if b {
		return "on"
	}
	return "off"
}
