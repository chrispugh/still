package journal

import (
	"fmt"
	"strings"
	"time"
)

type Entry struct {
	Date     time.Time
	Mood     int
	Tags     []string
	Raw      string
	Polished string
	FilePath string
}

func (e *Entry) ToMarkdown() string {
	var sb strings.Builder

	sb.WriteString("---\n")
	sb.WriteString(fmt.Sprintf("date: %s\n", e.Date.Format("2006-01-02")))
	sb.WriteString(fmt.Sprintf("mood: %d\n", e.Mood))
	if len(e.Tags) > 0 {
		sb.WriteString(fmt.Sprintf("tags: [%s]\n", strings.Join(e.Tags, ", ")))
	} else {
		sb.WriteString("tags: []\n")
	}
	sb.WriteString("---\n\n")

	sb.WriteString("## Raw\n\n")
	sb.WriteString(strings.TrimSpace(e.Raw))
	sb.WriteString("\n")

	if e.Polished != "" {
		sb.WriteString("\n## Polished\n\n")
		sb.WriteString(strings.TrimSpace(e.Polished))
		sb.WriteString("\n")
	}

	return sb.String()
}

func ParseMarkdown(content string) (*Entry, error) {
	entry := &Entry{}

	if !strings.HasPrefix(content, "---") {
		return nil, fmt.Errorf("no frontmatter found")
	}

	rest := content[3:]
	parts := strings.SplitN(rest, "---", 2)
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid frontmatter")
	}

	frontmatter := parts[0]
	body := parts[1]

	for _, line := range strings.Split(frontmatter, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		kv := strings.SplitN(line, ": ", 2)
		if len(kv) != 2 {
			continue
		}
		key := strings.TrimSpace(kv[0])
		value := strings.TrimSpace(kv[1])

		switch key {
		case "date":
			t, err := time.Parse("2006-01-02", value)
			if err == nil {
				entry.Date = t
			}
		case "mood":
			fmt.Sscanf(value, "%d", &entry.Mood)
		case "tags":
			value = strings.Trim(value, "[]")
			if value != "" {
				for _, tag := range strings.Split(value, ",") {
					t := strings.TrimSpace(tag)
					if t != "" {
						entry.Tags = append(entry.Tags, t)
					}
				}
			}
		}
	}

	// Parse body sections — split on "\n## " but handle leading "## " too
	body = strings.TrimPrefix(body, "\n")
	sections := strings.Split(body, "\n## ")
	for i, section := range sections {
		if i == 0 {
			section = strings.TrimPrefix(section, "## ")
		}
		section = strings.TrimSpace(section)
		if strings.HasPrefix(section, "Raw") {
			entry.Raw = strings.TrimSpace(strings.TrimPrefix(section, "Raw"))
		} else if strings.HasPrefix(section, "Polished") {
			entry.Polished = strings.TrimSpace(strings.TrimPrefix(section, "Polished"))
		}
	}

	return entry, nil
}

func MoodEmoji(mood int) string {
	switch mood {
	case 1:
		return "😔"
	case 2:
		return "😕"
	case 3:
		return "😐"
	case 4:
		return "🙂"
	case 5:
		return "😄"
	default:
		return "•"
	}
}

func WordCount(text string) int {
	return len(strings.Fields(text))
}
