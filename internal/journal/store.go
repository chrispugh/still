package journal

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type Store struct {
	BasePath string
}

func NewStore(basePath string) *Store {
	return &Store{BasePath: basePath}
}

func (s *Store) EntryPath(date time.Time) string {
	return filepath.Join(
		s.BasePath,
		date.Format("2006"),
		date.Format("01"),
		date.Format("02")+".md",
	)
}

func (s *Store) HasEntryToday() bool {
	_, err := os.Stat(s.EntryPath(time.Now()))
	return err == nil
}

func (s *Store) SaveEntry(entry *Entry) error {
	path := s.EntryPath(entry.Date)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	entry.FilePath = path
	return os.WriteFile(path, []byte(entry.ToMarkdown()), 0644)
}

func (s *Store) LoadEntry(date time.Time) (*Entry, error) {
	path := s.EntryPath(date)
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	entry, err := ParseMarkdown(string(content))
	if err != nil {
		return nil, err
	}
	entry.FilePath = path
	return entry, nil
}

func (s *Store) AllEntries() ([]*Entry, error) {
	var entries []*Entry

	err := filepath.Walk(s.BasePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		entry, err := ParseMarkdown(string(content))
		if err != nil {
			return nil
		}
		entry.FilePath = path
		entries = append(entries, entry)
		return nil
	})

	if err != nil {
		return nil, err
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Date.After(entries[j].Date)
	})

	return entries, nil
}

func (s *Store) OnThisDay() ([]*Entry, error) {
	now := time.Now()
	allEntries, err := s.AllEntries()
	if err != nil {
		return nil, err
	}

	var entries []*Entry
	for _, entry := range allEntries {
		if entry.Date.Month() == now.Month() &&
			entry.Date.Day() == now.Day() &&
			entry.Date.Year() != now.Year() {
			entries = append(entries, entry)
		}
	}
	return entries, nil
}

func (s *Store) Streak() int {
	now := time.Now()
	streak := 0

	startOffset := 0
	if !s.HasEntryToday() {
		startOffset = 1
	}

	for i := startOffset; ; i++ {
		day := now.AddDate(0, 0, -i)
		if _, err := os.Stat(s.EntryPath(day)); err != nil {
			break
		}
		streak++
	}

	return streak
}

func (s *Store) LongestStreak() int {
	entries, err := s.AllEntries()
	if err != nil || len(entries) == 0 {
		return 0
	}

	// Sort ascending for streak calculation
	sorted := make([]*Entry, len(entries))
	copy(sorted, entries)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Date.Before(sorted[j].Date)
	})

	longest := 1
	current := 1

	for i := 1; i < len(sorted); i++ {
		prev := sorted[i-1].Date
		curr := sorted[i].Date
		diff := curr.Sub(prev)
		if diff.Hours() <= 48 && diff.Hours() > 0 { // consecutive days
			current++
			if current > longest {
				longest = current
			}
		} else if diff.Hours() > 48 {
			current = 1
		}
	}

	return longest
}

func (s *Store) TotalEntries() int {
	entries, _ := s.AllEntries()
	return len(entries)
}

func (s *Store) AvgWordCount() int {
	entries, err := s.AllEntries()
	if err != nil || len(entries) == 0 {
		return 0
	}
	total := 0
	for _, e := range entries {
		total += WordCount(e.Raw)
	}
	return total / len(entries)
}

// MoodHistory returns (date, mood) pairs for the last N days
func (s *Store) MoodHistory(days int) []struct {
	Date time.Time
	Mood int
} {
	var result []struct {
		Date time.Time
		Mood int
	}

	now := time.Now()
	for i := days - 1; i >= 0; i-- {
		day := now.AddDate(0, 0, -i)
		entry, err := s.LoadEntry(day)
		if err != nil {
			result = append(result, struct {
				Date time.Time
				Mood int
			}{day, 0})
		} else {
			result = append(result, struct {
				Date time.Time
				Mood int
			}{day, entry.Mood})
		}
	}
	return result
}
