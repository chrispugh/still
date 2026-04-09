package styles

import "github.com/charmbracelet/lipgloss"

// Color palette — warm journal aesthetic
var (
	ColorPrimary   = lipgloss.Color("#D4A574")
	ColorSecondary = lipgloss.Color("#9B8A76")
	ColorText      = lipgloss.Color("#E8E0D5")
	ColorMuted     = lipgloss.Color("#6B6666")
	ColorError     = lipgloss.Color("#E07070")
	ColorSuccess   = lipgloss.Color("#7EAA7E")
	ColorHighlight = lipgloss.Color("#F0C89A")
)

var (
	AppName = lipgloss.NewStyle().
		Foreground(ColorPrimary).
		Bold(true)

	Title = lipgloss.NewStyle().
		Foreground(ColorPrimary).
		Bold(true)

	Subtitle = lipgloss.NewStyle().
		Foreground(ColorSecondary)

	Body = lipgloss.NewStyle().
		Foreground(ColorText)

	Muted = lipgloss.NewStyle().
		Foreground(ColorMuted)

	Success = lipgloss.NewStyle().
		Foreground(ColorSuccess)

	ErrorStyle = lipgloss.NewStyle().
		Foreground(ColorError)

	Selected = lipgloss.NewStyle().
		Foreground(ColorHighlight).
		Bold(true)

	Help = lipgloss.NewStyle().
		Foreground(ColorMuted).
		Italic(true)

	KeyHint = lipgloss.NewStyle().
		Foreground(ColorPrimary).
		Bold(true)

	Banner = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(ColorPrimary).
		Padding(1, 2).
		Foreground(ColorText)

	Box = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(ColorSecondary).
		Padding(1, 2)

	FocusedInput = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(ColorPrimary).
		Padding(0, 1)

	UnfocusedInput = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(ColorMuted).
		Padding(0, 1)

	MoodActive = lipgloss.NewStyle().
		Foreground(ColorHighlight).
		Bold(true)

	MoodInactive = lipgloss.NewStyle().
		Foreground(ColorMuted)
)

// Center returns content centered within the given width.
func Center(width int, content string) string {
	return lipgloss.NewStyle().Width(width).Align(lipgloss.Center).Render(content)
}

// Pad wraps content with horizontal margin.
func Pad(h int, content string) string {
	return lipgloss.NewStyle().Padding(0, h).Render(content)
}
