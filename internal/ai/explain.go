package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mrf/runbook-generator/internal/history"
	"github.com/mrf/runbook-generator/internal/processor"
)

// ExplanationResult contains AI-generated explanations for command groups.
type ExplanationResult struct {
	Overview      string            `json:"overview"`
	Prerequisites []string          `json:"prerequisites"`
	Steps         []StepExplanation `json:"steps"`
}

// StepExplanation provides context for a command group.
type StepExplanation struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Why         string `json:"why"`
	Notes       string `json:"notes,omitempty"`
}

const explainSystemPrompt = `You are a technical writer creating runbook documentation from shell command sequences.
Your task is to analyze commands and generate clear, actionable explanations.

For each group of commands, provide:
1. A concise title (3-7 words)
2. A description of what the commands do
3. The "why" - explain the purpose and when someone would need to do this
4. Optional notes about prerequisites, gotchas, or alternatives

Also provide:
- An overview summarizing what this entire runbook accomplishes
- A list of prerequisites (tools, access, permissions needed)

Return a JSON object with this structure:
{
  "overview": "This runbook guides you through deploying a containerized application to Kubernetes.",
  "prerequisites": ["kubectl configured", "Docker installed", "Access to container registry"],
  "steps": [
    {
      "title": "Build the Docker Image",
      "description": "Compile the application and create a container image.",
      "why": "The container image packages your application with all dependencies for consistent deployment across environments.",
      "notes": "Ensure you're in the project root directory before building."
    }
  ]
}

Be practical and helpful. Focus on what engineers actually need to know.
Avoid generic filler text - every sentence should add value.`

// GenerateExplanations creates AI-powered explanations for command groups.
func (c *Client) GenerateExplanations(ctx context.Context, groups []processor.CommandGroup) (*ExplanationResult, error) {
	if len(groups) == 0 {
		return &ExplanationResult{}, nil
	}

	// Build description of command groups
	var sb strings.Builder
	sb.WriteString("Generate explanations for this command sequence:\n\n")

	for i, group := range groups {
		sb.WriteString(fmt.Sprintf("## Group %d", i+1))
		if group.Title != "" {
			sb.WriteString(fmt.Sprintf(": %s", group.Title))
		}
		sb.WriteString("\n")

		for _, entry := range group.Commands {
			sb.WriteString(fmt.Sprintf("  $ %s\n", entry.Command))
		}
		sb.WriteString("\n")
	}

	response, err := c.sendMessage(ctx, explainSystemPrompt, sb.String())
	if err != nil {
		return nil, fmt.Errorf("AI explanation generation failed: %w", err)
	}

	// Parse JSON response
	result := &ExplanationResult{}

	// Extract JSON from response
	jsonStart := strings.Index(response, "{")
	jsonEnd := strings.LastIndex(response, "}")
	if jsonStart >= 0 && jsonEnd > jsonStart {
		jsonStr := response[jsonStart : jsonEnd+1]
		if err := json.Unmarshal([]byte(jsonStr), result); err != nil {
			// If parsing fails, return empty result
			return &ExplanationResult{}, nil
		}
	}

	return result, nil
}

// EnhanceGroups merges AI explanations into command groups.
func EnhanceGroups(groups []processor.CommandGroup, explanations *ExplanationResult) []processor.CommandGroup {
	if explanations == nil || len(explanations.Steps) == 0 {
		return groups
	}

	enhanced := make([]processor.CommandGroup, len(groups))
	copy(enhanced, groups)

	for i := range enhanced {
		if i < len(explanations.Steps) {
			exp := explanations.Steps[i]
			if exp.Title != "" {
				enhanced[i].Title = exp.Title
			}
			if exp.Description != "" {
				enhanced[i].Description = exp.Description
			}
			if exp.Why != "" {
				enhanced[i].Intent = exp.Why
			}
		}
	}

	return enhanced
}

// ExplainCommands generates explanations for raw command entries.
// This is useful when you want explanations before grouping.
func (c *Client) ExplainCommands(ctx context.Context, entries []history.Entry) (*ExplanationResult, error) {
	if len(entries) == 0 {
		return &ExplanationResult{}, nil
	}

	// Convert entries to a single group for analysis
	group := processor.CommandGroup{
		Commands: entries,
	}

	return c.GenerateExplanations(ctx, []processor.CommandGroup{group})
}
