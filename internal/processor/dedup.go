package processor

import (
	"strings"
	"time"

	"github.com/mrf/runbook-generator/internal/history"
)

// Dedup removes redundant commands while preserving meaningful repetition.
type Dedup struct {
	TimeGap time.Duration // Gap to consider commands intentionally repeated
}

// NewDedup creates a new deduplicator with default settings.
func NewDedup() *Dedup {
	return &Dedup{
		TimeGap: 30 * time.Second,
	}
}

// Process removes duplicate and redundant commands from the entry list.
func (d *Dedup) Process(entries []history.Entry) []history.Entry {
	if len(entries) == 0 {
		return entries
	}

	var result []history.Entry

	for i, entry := range entries {
		// Skip empty commands
		if strings.TrimSpace(entry.Command) == "" {
			continue
		}

		// First entry always included
		if len(result) == 0 {
			result = append(result, entry)
			continue
		}

		prev := result[len(result)-1]

		// Check if this is an exact duplicate
		if d.isExactDuplicate(prev, entry) {
			// If there's a significant time gap, it's intentional repetition
			if d.hasSignificantGap(prev, entry) {
				result = append(result, entry)
			}
			// Otherwise skip the duplicate
			continue
		}

		// Check if this is a typo correction (minor edit of previous command)
		if d.isTypoCorrection(prev.Command, entry.Command) {
			// Replace previous with current (keep the corrected version)
			result[len(result)-1] = entry
			continue
		}

		// Check if this collapses with previous (e.g., multiple cd commands)
		if d.shouldCollapse(prev.Command, entry.Command) {
			// Replace previous with current
			result[len(result)-1] = entry
			continue
		}

		// Look ahead to see if this command is followed by a corrected version
		if i+1 < len(entries) && d.isTypoCorrection(entry.Command, entries[i+1].Command) {
			// Skip this one, we'll use the next one
			continue
		}

		result = append(result, entry)
	}

	return result
}

// isExactDuplicate checks if two entries have the exact same command.
func (d *Dedup) isExactDuplicate(a, b history.Entry) bool {
	return strings.TrimSpace(a.Command) == strings.TrimSpace(b.Command)
}

// hasSignificantGap checks if there's a meaningful time gap between entries.
func (d *Dedup) hasSignificantGap(a, b history.Entry) bool {
	if !a.HasTime || !b.HasTime {
		return false
	}
	return b.Timestamp.Sub(a.Timestamp) > d.TimeGap
}

// isTypoCorrection checks if command b is a minor edit of command a.
func (d *Dedup) isTypoCorrection(a, b string) bool {
	a = strings.TrimSpace(a)
	b = strings.TrimSpace(b)

	// Commands must be similar length
	lenDiff := len(a) - len(b)
	if lenDiff < 0 {
		lenDiff = -lenDiff
	}
	if lenDiff > 5 {
		return false
	}

	// Calculate Levenshtein distance
	distance := levenshteinDistance(a, b)

	// Consider it a typo correction if distance is small relative to length
	maxLen := len(a)
	if len(b) > maxLen {
		maxLen = len(b)
	}

	// For short commands, allow 1-2 edits; for longer ones, scale proportionally
	threshold := maxLen / 10
	if threshold < 2 {
		threshold = 2
	}
	if threshold > 5 {
		threshold = 5
	}

	return distance > 0 && distance <= threshold
}

// shouldCollapse checks if consecutive commands should be collapsed.
func (d *Dedup) shouldCollapse(a, b string) bool {
	a = strings.TrimSpace(a)
	b = strings.TrimSpace(b)

	// Collapse multiple cd commands
	if strings.HasPrefix(a, "cd ") && strings.HasPrefix(b, "cd ") {
		return true
	}

	// Collapse multiple export commands for the same variable
	if strings.HasPrefix(a, "export ") && strings.HasPrefix(b, "export ") {
		aVar := extractExportVar(a)
		bVar := extractExportVar(b)
		if aVar != "" && aVar == bVar {
			return true
		}
	}

	return false
}

// extractExportVar extracts the variable name from an export command.
func extractExportVar(cmd string) string {
	cmd = strings.TrimPrefix(cmd, "export ")
	parts := strings.SplitN(cmd, "=", 2)
	if len(parts) > 0 {
		return strings.TrimSpace(parts[0])
	}
	return ""
}

// levenshteinDistance calculates the edit distance between two strings.
func levenshteinDistance(a, b string) int {
	if len(a) == 0 {
		return len(b)
	}
	if len(b) == 0 {
		return len(a)
	}

	// Create matrix
	matrix := make([][]int, len(a)+1)
	for i := range matrix {
		matrix[i] = make([]int, len(b)+1)
		matrix[i][0] = i
	}
	for j := range matrix[0] {
		matrix[0][j] = j
	}

	// Fill matrix
	for i := 1; i <= len(a); i++ {
		for j := 1; j <= len(b); j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			matrix[i][j] = min(
				matrix[i-1][j]+1,      // deletion
				matrix[i][j-1]+1,      // insertion
				matrix[i-1][j-1]+cost, // substitution
			)
		}
	}

	return matrix[len(a)][len(b)]
}
