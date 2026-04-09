# still

> For people who think more than they write.

A terminal journaling app with a TUI. Write messy thoughts; a local AI model rewrites them in your voice. Entries are plain Markdown files. No cloud, no subscription, no account.

## Install

```sh
go install github.com/chrispugh/still@latest
```

Or build from source:

```sh
git clone https://github.com/chrispugh/still
cd still
go build -o still .
```

## Usage

```sh
still
```

On first run, a two-minute setup wizard walks you through:

1. Journal location (default: `~/.journal`)
2. AI features opt-in (requires [Ollama](https://ollama.ai))
3. Voice calibration — 6 questions that shape how the AI rewrites your entries

## Features

- **New Entry** — write freely; AI polish is optional and shown side-by-side before you choose
- **Browse** — calendar/list view of past entries with rendered Markdown
- **Search** — full-text search across all entries
- **Stats** — streak, total entries, avg word count, mood heatmap
- **Settings** — toggle AI, change model, edit voice profile, adjust nudge time

## Storage

Entries are stored as dated Markdown files:

```
~/.journal/
  2026/
    04/
      09.md
  config.toml
```

Each entry:

```markdown
---
date: 2026-04-09
mood: 4
tags: [work, travel]
---

## Raw

Your unedited words.

## Polished

AI-rewritten version (only present if you chose polish).
```

## AI

AI features use [Ollama](https://ollama.ai) — a local model runner. Nothing leaves your machine.

- Install Ollama from [ollama.ai](https://ollama.ai)
- still will auto-select a model based on your RAM (llama3 for 8GB+, gemma2:2b for less)

## Keyboard shortcuts

| Key | Action |
|-----|--------|
| `↑↓` / `jk` | Navigate |
| `enter` | Select / confirm |
| `esc` | Back |
| `ctrl+s` | Save entry |
| `ctrl+p` | Generate writing prompt (New Entry) |
| `q` | Quit |

## Tech stack

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) — TUI framework
- [Lipgloss](https://github.com/charmbracelet/lipgloss) — styling
- [Glamour](https://github.com/charmbracelet/glamour) — Markdown rendering
- [Ollama](https://ollama.ai) — local AI
