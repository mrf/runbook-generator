# Runbook Generator Architecture

A CLI tool that transforms zsh history into actionable runbook documentation.

## Overview

Reads `~/.zsh_history`, extracts commands by number, deduplicates, sanitizes secrets, analyzes intent, and generates markdown runbooks.

## Technology Stack

- **Language**: Go 1.21+
- **CLI**: [cobra](https://github.com/spf13/cobra)
- **AI** (optional): [anthropic-sdk-go](https://github.com/anthropics/anthropic-sdk-go)

## Project Structure

```
runbook-generator/
├── cmd/runbook-gen/main.go     # Entry point
├── internal/
│   ├── ai/                     # Optional Claude integration
│   │   ├── client.go
│   │   ├── dedup.go
│   │   └── explain.go
│   ├── cli/root.go             # CLI commands
│   ├── history/
│   │   ├── entry.go            # Entry type
│   │   ├── extractor.go        # Zsh history parsing
│   │   └── extractor_test.go
│   ├── processor/
│   │   ├── dedup.go            # Deduplication
│   │   ├── intent.go           # Intent grouping
│   │   ├── patterns.go         # Secret patterns
│   │   └── sanitizer.go        # Secret redaction
│   └── generator/markdown.go   # Markdown output
├── go.mod
├── Makefile
└── README.md
```

## Data Flow

```
~/.zsh_history
       │
       ▼
┌──────────────────┐
│ history.Extract  │ → []Entry (by command number)
└──────────────────┘
       │
       ▼
┌──────────────────┐
│ processor.Dedup  │ → []Entry (deduplicated)
│   (or AI dedup)  │
└──────────────────┘
       │
       ▼
┌──────────────────┐
│processor.Sanitize│ → []Entry (secrets redacted)
└──────────────────┘
       │
       ▼
┌──────────────────┐
│processor.Analyze │ → []CommandGroup
└──────────────────┘
       │
       ▼
┌──────────────────┐
│ AI Explanations  │ → Enhanced groups (optional)
└──────────────────┘
       │
       ▼
┌──────────────────┐
│generator.Generate│ → Markdown string
└──────────────────┘
       │
       ▼
   output.md
```

## Core Components

### History Extractor

Reads `~/.zsh_history` and extracts entries by command number.

```go
type Entry struct {
    Number    int       // Command number (matches history output)
    Timestamp time.Time
    Command   string
    HasTime   bool
}

func (e *Extractor) Extract(from, to int) ([]Entry, error)
```

Zsh history format: `: 1234567890:0;command`

### Processor

**Deduplicator**: Removes exact consecutive duplicates, typo corrections (Levenshtein distance < 3), and collapsed cd/export commands.

**Sanitizer**: 55+ regex patterns detecting passwords, API keys, tokens, private keys, webhook URLs, etc.

**Intent Analyzer**: Groups commands by tool (git, docker, kubectl) and workflow patterns.

### AI Module (Optional)

When `ANTHROPIC_API_KEY` is set:

- **Deduplication**: Semantic analysis finds duplicates regex misses
- **Explanations**: Generates "why" text, overview, prerequisites

Uses Claude 3.5 Haiku for cost-effective processing.

## Security

- Never transmits data externally
- Sanitization always runs (not optional)
- Output has 0600 permissions
- Only reads files, never executes commands
