package screens

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/chrispugh/still/internal/ai"
	"github.com/chrispugh/still/internal/config"
	"github.com/chrispugh/still/internal/tui/styles"
)

type obStep int

const (
	obWelcome obStep = iota
	obJournalPath
	obAIFeatures
	obOllamaCheck
	obVoiceCalibration
	obDone
)

// Voice calibration questions
var voiceQuestions = []struct {
	q    string
	hint string
}{
	{
		q:    "Are you more of a storyteller or straight-to-the-point?",
		hint: "e.g. I prefer to set the scene / Just the facts",
	},
	{
		q:    "Do you name emotions directly or describe them through actions?",
		hint: "e.g. I was furious / I slammed the door",
	},
	{
		q:    "How do you feel about humor in serious moments?",
		hint: "e.g. I lean on it / I keep things earnest",
	},
	{
		q:    "Do you notice small details or zoom to the big picture?",
		hint: "e.g. I notice what people wear / I track the arc of the day",
	},
	{
		q:    "Paste a few sentences you've written that sound like you. (optional)",
		hint: "From a text, email, or anything — just something in your voice",
	},
	{
		q:    "Raw and unpolished, or clean and considered?",
		hint: "e.g. I write like I talk / I like to edit as I go",
	},
}

// OnboardingModel runs the first-time setup wizard.
type OnboardingModel struct {
	config *config.Config
	width  int
	height int
	step   obStep

	input     textinput.Model
	aiEnabled bool

	voiceStep   int
	voiceInputs []string

	ollamaFound bool
	modelChosen string
	modelReason string

	err string
}

func NewOnboarding(cfg *config.Config) *OnboardingModel {
	ti := textinput.New()
	ti.CharLimit = 256

	model, reason := ai.RecommendModel()

	return &OnboardingModel{
		config:      cfg,
		step:        obWelcome,
		input:       ti,
		aiEnabled:   true,
		voiceInputs: make([]string, len(voiceQuestions)),
		ollamaFound: ai.IsAvailable(),
		modelChosen: model,
		modelReason: reason,
	}
}

func (m *OnboardingModel) Init() tea.Cmd { return nil }

func (m *OnboardingModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		return m.handleKey(msg)
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m *OnboardingModel) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	switch m.step {
	case obWelcome:
		if key == "enter" || key == " " {
			m.step = obJournalPath
			m.input.SetValue(m.config.JournalPath)
			m.input.Focus()
			return m, textinput.Blink
		}

	case obJournalPath:
		switch key {
		case "enter":
			v := strings.TrimSpace(m.input.Value())
			if v == "" {
				v = m.config.JournalPath
			}
			m.config.JournalPath = config.ExpandPath(v)
			m.step = obAIFeatures
			m.input.SetValue("")
			return m, nil
		default:
			var cmd tea.Cmd
			m.input, cmd = m.input.Update(msg)
			return m, cmd
		}

	case obAIFeatures:
		switch key {
		case "y", "Y", "enter":
			m.aiEnabled = true
			m.config.AI.Enabled = true
			m.step = obOllamaCheck
			return m, nil
		case "n", "N":
			m.aiEnabled = false
			m.config.AI.Enabled = false
			m.step = obVoiceCalibration
			return m, nil
		case "esc":
			m.step = obJournalPath
			return m, nil
		}

	case obOllamaCheck:
		if key == "enter" {
			m.config.AI.Model = m.modelChosen
			m.step = obVoiceCalibration
			m.input.SetValue("")
			m.input.Focus()
			return m, textinput.Blink
		}

	case obVoiceCalibration:
		switch key {
		case "ctrl+d", "enter":
			m.voiceInputs[m.voiceStep] = m.input.Value()
			if m.voiceStep < len(voiceQuestions)-1 {
				m.voiceStep++
				m.input.SetValue("")
				return m, textinput.Blink
			}
			m.config.AI.VoiceProfile = m.buildVoiceProfile()
			m.step = obDone
			return m, nil
		case "ctrl+s":
			// Skip voice calibration
			m.step = obDone
			return m, nil
		default:
			var cmd tea.Cmd
			m.input, cmd = m.input.Update(msg)
			return m, cmd
		}

	case obDone:
		if key == "enter" || key == " " {
			if err := os.MkdirAll(m.config.JournalPath, 0755); err == nil {
				_ = m.config.Save()
			}
			return m, func() tea.Msg {
				return ChangeScreenMsg{Screen: ScreenHome}
			}
		}
	}

	return m, nil
}

func (m *OnboardingModel) buildVoiceProfile() string {
	var parts []string
	for i, q := range voiceQuestions {
		ans := strings.TrimSpace(m.voiceInputs[i])
		if ans != "" {
			parts = append(parts, fmt.Sprintf("Q: %s\nA: %s", q.q, ans))
		}
	}
	if len(parts) == 0 {
		return m.config.AI.VoiceProfile
	}
	return strings.Join(parts, "\n\n")
}

// ─── View ─────────────────────────────────────────────────────────────────────

func (m *OnboardingModel) View() string {
	switch m.step {
	case obWelcome:
		return m.viewWelcome()
	case obJournalPath:
		return m.viewJournalPath()
	case obAIFeatures:
		return m.viewAIFeatures()
	case obOllamaCheck:
		return m.viewOllamaCheck()
	case obVoiceCalibration:
		return m.viewVoiceCalibration()
	case obDone:
		return m.viewDone()
	}
	return ""
}

func (m *OnboardingModel) frame(title, body, help string) string {
	maxW := 72
	if m.width > 0 && m.width < maxW+8 {
		maxW = m.width - 8
	}

	content := styles.Title.Render(title) + "\n\n" + body
	box := styles.Box.Width(maxW).Render(content)

	var sb strings.Builder
	boxH := strings.Count(box, "\n") + 1
	topPad := (m.height - boxH - 4) / 2
	if topPad < 2 {
		topPad = 2
	}
	sb.WriteString(strings.Repeat("\n", topPad))
	sb.WriteString(styles.Center(m.width, box))
	if help != "" {
		sb.WriteString("\n\n")
		sb.WriteString(styles.Center(m.width, styles.Help.Render(help)))
	}
	return sb.String()
}

func (m *OnboardingModel) viewWelcome() string {
	body := styles.Body.Render("A quiet place to think.\n\n") +
		styles.Muted.Render("still is a local journaling app.\n") +
		styles.Muted.Render("Plain Markdown files. No cloud. No account.\n") +
		styles.Muted.Render("A local AI model can rewrite your messy thoughts in your voice.\n\n") +
		styles.Subtitle.Render("Let's get you set up. This takes about two minutes.")
	return m.frame("still", body, "press enter to begin")
}

func (m *OnboardingModel) viewJournalPath() string {
	body := styles.Body.Render("Where should your journal live?\n\n") +
		styles.FocusedInput.Width(60).Render(m.input.View()) + "\n\n" +
		styles.Muted.Render("Entries are plain Markdown files — you can read them anywhere.")
	return m.frame("Journal location", body, "enter to confirm")
}

func (m *OnboardingModel) viewAIFeatures() string {
	body := styles.Body.Render("Would you like AI polish?\n\n") +
		styles.Muted.Render("A local model (Ollama) rewrites your raw thoughts in your voice.\n") +
		styles.Muted.Render("No data leaves your machine. Requires ~4 GB of disk space.\n\n") +
		styles.KeyHint.Render("y") + styles.Body.Render("  enable AI features\n") +
		styles.KeyHint.Render("n") + styles.Body.Render("  skip for now  (you can enable later in Settings)")
	return m.frame("AI features", body, "")
}

func (m *OnboardingModel) viewOllamaCheck() string {
	var statusLine string
	if m.ollamaFound {
		statusLine = styles.Success.Render("✓ Ollama is running")
	} else {
		statusLine = styles.ErrorStyle.Render("✗ Ollama not found\n\n") +
			styles.Body.Render("Install it from ollama.ai, then run:\n") +
			styles.KeyHint.Render("ollama pull "+m.modelChosen)
	}

	var ramNote string
	if runtime.GOARCH == "arm64" {
		ramNote = "Apple Silicon detected."
	} else {
		ramNote = "System detected."
	}

	body := statusLine + "\n\n" +
		styles.Body.Render("Model: ") + styles.KeyHint.Render(m.modelChosen) + "\n" +
		styles.Muted.Render(ramNote+" "+m.modelReason)

	return m.frame("Ollama setup", body, "enter to continue")
}

func (m *OnboardingModel) viewVoiceCalibration() string {
	if m.voiceStep >= len(voiceQuestions) {
		return ""
	}

	q := voiceQuestions[m.voiceStep]
	progress := styles.Muted.Render(fmt.Sprintf("Question %d of %d", m.voiceStep+1, len(voiceQuestions)))

	body := progress + "\n\n" +
		styles.Body.Render(q.q) + "\n" +
		styles.Muted.Render(q.hint) + "\n\n" +
		styles.FocusedInput.Width(60).Render(m.input.View())

	help := "enter  next question   ctrl+s  skip calibration"
	return m.frame("Voice calibration", body, help)
}

func (m *OnboardingModel) viewDone() string {
	body := styles.Success.Render("✓ All set!\n\n") +
		styles.Body.Render("Your journal lives at:\n") +
		styles.KeyHint.Render(m.config.JournalPath) + "\n\n" +
		styles.Muted.Render("Entries are stored as plain Markdown files.\n") +
		styles.Muted.Render("You can open and edit them in any editor.")
	return m.frame("You're ready", body, "enter to start writing")
}
