package processor

import (
	"github.com/mrf/runbook-generator/internal/history"
)

// Redaction records what was redacted from a command.
type Redaction struct {
	EntryNumber int
	PatternName string
	Original    string // Only populated in strict mode
}

// Sanitizer removes or redacts sensitive information from commands.
type Sanitizer struct {
	patterns   []Pattern
	allowlist  map[string]bool
	strictMode bool
}

// NewSanitizer creates a new sanitizer with default patterns.
func NewSanitizer() *Sanitizer {
	return &Sanitizer{
		patterns:  DefaultPatterns(),
		allowlist: make(map[string]bool),
	}
}

// WithPatterns adds custom patterns to the sanitizer.
func (s *Sanitizer) WithPatterns(patterns []Pattern) *Sanitizer {
	s.patterns = append(s.patterns, patterns...)
	return s
}

// WithAllowlist sets values that should not be redacted.
func (s *Sanitizer) WithAllowlist(values []string) *Sanitizer {
	for _, v := range values {
		s.allowlist[v] = true
	}
	return s
}

// WithStrictMode enables strict mode which preserves originals for review.
func (s *Sanitizer) WithStrictMode(strict bool) *Sanitizer {
	s.strictMode = strict
	return s
}

// Process sanitizes entries and returns both sanitized entries and redaction records.
func (s *Sanitizer) Process(entries []history.Entry) ([]history.Entry, []Redaction) {
	var result []history.Entry
	var redactions []Redaction

	for _, entry := range entries {
		sanitized, entryRedactions := s.sanitizeEntry(entry)
		if sanitized != nil {
			result = append(result, *sanitized)
		}
		redactions = append(redactions, entryRedactions...)
	}

	return result, redactions
}

// sanitizeEntry processes a single entry and returns the sanitized version.
func (s *Sanitizer) sanitizeEntry(entry history.Entry) (*history.Entry, []Redaction) {
	var redactions []Redaction
	command := entry.Command
	original := command

	for _, pattern := range s.patterns {
		if !pattern.Regex.MatchString(command) {
			continue
		}

		// Check if this command should be fully removed
		if pattern.FullRemove {
			redaction := Redaction{
				EntryNumber: entry.Number,
				PatternName: pattern.Name,
			}
			if s.strictMode {
				redaction.Original = original
			}
			return nil, []Redaction{redaction}
		}

		// Apply the replacement
		newCommand := pattern.Regex.ReplaceAllString(command, pattern.Replacement)
		if newCommand != command {
			redactions = append(redactions, Redaction{
				EntryNumber: entry.Number,
				PatternName: pattern.Name,
				Original:    func() string { if s.strictMode { return original }; return "" }(),
			})
			command = newCommand
		}
	}

	// Return modified entry
	result := entry
	result.Command = command
	return &result, redactions
}

// SanitizeString sanitizes a single command string.
func (s *Sanitizer) SanitizeString(command string) string {
	for _, pattern := range s.patterns {
		if pattern.FullRemove && pattern.Regex.MatchString(command) {
			return "[REDACTED - contains sensitive data]"
		}
		command = pattern.Regex.ReplaceAllString(command, pattern.Replacement)
	}
	return command
}
