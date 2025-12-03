package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mrf/runbook-generator/internal/history"
)

// DedupResult represents commands grouped by semantic similarity.
type DedupResult struct {
	Groups []CommandGroup `json:"groups"`
}

// CommandGroup represents semantically related commands.
type CommandGroup struct {
	Representative int    `json:"representative"` // Index of the command to keep
	Indices        []int  `json:"indices"`        // All command indices in this group
	Reason         string `json:"reason"`         // Why these are grouped
}

const dedupSystemPrompt = `You are a command-line expert analyzing shell command sequences.
Your task is to identify semantically similar or duplicate commands that should be deduplicated.

Group commands that are:
- Exact duplicates
- Typo corrections (e.g., "git stauts" followed by "git status")
- Same command with minor flag variations that don't change intent
- Failed attempts followed by successful versions (keep the successful one)
- Repeated status checks (e.g., multiple "kubectl get pods" - keep the last one)

Do NOT group commands that are:
- Intentionally repeated for different purposes
- Similar but operating on different targets
- Part of a deliberate retry pattern with meaningful changes

Return a JSON object with this structure:
{
  "groups": [
    {
      "representative": 2,
      "indices": [0, 1, 2],
      "reason": "Typo correction: 'git stauts' corrected to 'git status'"
    }
  ]
}

Only include groups with more than one command. Commands not in any group will be kept as-is.
Use 0-based indices matching the input order.`

// DeduplicateCommands uses Claude to identify semantically similar commands.
func (c *Client) DeduplicateCommands(ctx context.Context, entries []history.Entry) (*DedupResult, error) {
	if len(entries) == 0 {
		return &DedupResult{}, nil
	}

	// Build command list for analysis
	var sb strings.Builder
	sb.WriteString("Analyze these commands for semantic duplicates:\n\n")
	for i, entry := range entries {
		sb.WriteString(fmt.Sprintf("%d: %s\n", i, entry.Command))
	}

	response, err := c.sendMessage(ctx, dedupSystemPrompt, sb.String())
	if err != nil {
		return nil, fmt.Errorf("AI deduplication failed: %w", err)
	}

	// Parse JSON response
	result := &DedupResult{}

	// Extract JSON from response (Claude may add explanation text)
	jsonStart := strings.Index(response, "{")
	jsonEnd := strings.LastIndex(response, "}")
	if jsonStart >= 0 && jsonEnd > jsonStart {
		jsonStr := response[jsonStart : jsonEnd+1]
		if err := json.Unmarshal([]byte(jsonStr), result); err != nil {
			// If parsing fails, return empty result (fall back to regular dedup)
			return &DedupResult{}, nil
		}
	}

	return result, nil
}

// ApplyDedup applies AI deduplication results to entries.
// Returns the deduplicated entries and a summary of what was removed.
func ApplyDedup(entries []history.Entry, result *DedupResult) ([]history.Entry, []string) {
	if result == nil || len(result.Groups) == 0 {
		return entries, nil
	}

	// Build set of indices to remove (all except representatives)
	removeSet := make(map[int]bool)
	var summaries []string

	for _, group := range result.Groups {
		for _, idx := range group.Indices {
			if idx != group.Representative && idx >= 0 && idx < len(entries) {
				removeSet[idx] = true
			}
		}
		if group.Reason != "" {
			summaries = append(summaries, group.Reason)
		}
	}

	// Filter entries
	var filtered []history.Entry
	for i, entry := range entries {
		if !removeSet[i] {
			filtered = append(filtered, entry)
		}
	}

	return filtered, summaries
}
