package generator

import (
	"fmt"
	"strings"
	"time"

	"github.com/mrf/runbook-generator/internal/processor"
)

// RunbookData contains all data needed to generate a runbook.
type RunbookData struct {
	Title            string
	Generated        time.Time
	TimeRange        string
	Groups           []processor.CommandGroup
	RedactedCount    int
	AIOverview       string   // AI-generated overview (optional)
	AIPrerequisites  []string // AI-generated prerequisites (optional)
}

// MarkdownGenerator generates markdown runbooks from command groups.
type MarkdownGenerator struct {
	includeTimestamps bool
	includeDirs       bool
}

// NewMarkdownGenerator creates a new markdown generator.
func NewMarkdownGenerator() *MarkdownGenerator {
	return &MarkdownGenerator{
		includeTimestamps: false,
		includeDirs:       true,
	}
}

// WithTimestamps enables timestamp inclusion in output.
func (g *MarkdownGenerator) WithTimestamps(include bool) *MarkdownGenerator {
	g.includeTimestamps = include
	return g
}

// WithDirectories enables directory inclusion in output.
func (g *MarkdownGenerator) WithDirectories(include bool) *MarkdownGenerator {
	g.includeDirs = include
	return g
}

// Generate creates a markdown runbook from the provided data.
func (g *MarkdownGenerator) Generate(data RunbookData) string {
	var sb strings.Builder

	// Title
	sb.WriteString(fmt.Sprintf("# %s\n\n", data.Title))

	// Overview (use AI if available, otherwise generate)
	sb.WriteString("## Overview\n\n")
	if data.AIOverview != "" {
		sb.WriteString(data.AIOverview)
	} else {
		sb.WriteString(g.generateOverview(data))
	}
	sb.WriteString("\n\n")

	// Prerequisites (use AI if available, otherwise infer)
	var prereqs []string
	if len(data.AIPrerequisites) > 0 {
		prereqs = data.AIPrerequisites
	} else {
		prereqs = g.inferPrerequisites(data.Groups)
	}
	if len(prereqs) > 0 {
		sb.WriteString("## Prerequisites\n\n")
		for _, prereq := range prereqs {
			sb.WriteString(fmt.Sprintf("- %s\n", prereq))
		}
		sb.WriteString("\n")
	}

	// Steps
	sb.WriteString("## Steps\n\n")
	for i, group := range data.Groups {
		sb.WriteString(g.generateStep(i+1, group))
		sb.WriteString("\n")
	}

	// Notes
	sb.WriteString("## Notes\n\n")
	sb.WriteString(fmt.Sprintf("- Generated from bash history on %s\n", data.Generated.Format("2006-01-02 15:04:05")))
	if data.TimeRange != "" {
		sb.WriteString(fmt.Sprintf("- Time range: %s\n", data.TimeRange))
	}
	if data.RedactedCount > 0 {
		sb.WriteString(fmt.Sprintf("- Commands sanitized: %d\n", data.RedactedCount))
	}

	return sb.String()
}

// generateOverview creates a summary of what the runbook accomplishes.
func (g *MarkdownGenerator) generateOverview(data RunbookData) string {
	if len(data.Groups) == 0 {
		return "This runbook contains no commands."
	}

	// Collect unique intents
	intents := make(map[string]bool)
	tools := make(map[string]bool)

	for _, group := range data.Groups {
		if group.Intent != "" {
			intents[group.Intent] = true
		}
		for _, cmd := range group.Commands {
			tool := extractTool(cmd.Command)
			if tool != "" {
				tools[tool] = true
			}
		}
	}

	var parts []string

	// Describe by intent if we have them
	if len(intents) > 0 {
		intentList := make([]string, 0, len(intents))
		for intent := range intents {
			intentList = append(intentList, formatIntent(intent))
		}
		parts = append(parts, fmt.Sprintf("This runbook covers: %s.", strings.Join(intentList, ", ")))
	}

	// Add command count
	totalCmds := 0
	for _, group := range data.Groups {
		totalCmds += len(group.Commands)
	}
	parts = append(parts, fmt.Sprintf("It contains %d steps with %d commands total.", len(data.Groups), totalCmds))

	return strings.Join(parts, " ")
}

// generateStep creates markdown for a single step.
func (g *MarkdownGenerator) generateStep(num int, group processor.CommandGroup) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("### Step %d: %s\n\n", num, group.Title))

	if group.Description != "" {
		sb.WriteString(group.Description)
		sb.WriteString("\n\n")
	}

	sb.WriteString("```bash\n")
	for _, cmd := range group.Commands {
		if g.includeTimestamps && cmd.HasTime {
			sb.WriteString(fmt.Sprintf("# %s\n", cmd.Timestamp.Format("15:04:05")))
		}
		sb.WriteString(cmd.Command)
		sb.WriteString("\n")
	}
	sb.WriteString("```\n")

	// Add "Why" section if we have an intent explanation
	if group.Intent != "" {
		sb.WriteString("\n**Why:** ")
		sb.WriteString(group.Intent)
		sb.WriteString("\n")
	}

	return sb.String()
}

// inferPrerequisites determines what tools/access are needed.
func (g *MarkdownGenerator) inferPrerequisites(groups []processor.CommandGroup) []string {
	tools := make(map[string]bool)

	for _, group := range groups {
		for _, cmd := range group.Commands {
			tool := extractTool(cmd.Command)
			if tool != "" {
				tools[tool] = true
			}
		}
	}

	var prereqs []string

	// Map tools to prerequisites
	prereqMap := map[string]string{
		"git":            "Git CLI installed",
		"docker":         "Docker installed and running",
		"docker-compose": "Docker Compose installed",
		"kubectl":        "kubectl installed with cluster access configured",
		"helm":           "Helm CLI installed",
		"terraform":      "Terraform CLI installed",
		"aws":            "AWS CLI installed and configured",
		"gcloud":         "Google Cloud SDK installed and configured",
		"az":             "Azure CLI installed and configured",
		"npm":            "Node.js and npm installed",
		"yarn":           "Yarn package manager installed",
		"go":             "Go toolchain installed",
		"python":         "Python installed",
		"python3":        "Python 3 installed",
		"pip":            "pip package manager installed",
		"pip3":           "pip3 package manager installed",
		"ssh":            "SSH client and appropriate key access",
		"scp":            "SSH/SCP access to remote hosts",
		"mysql":          "MySQL client installed with database access",
		"psql":           "PostgreSQL client installed with database access",
		"redis-cli":      "Redis CLI installed with server access",
		"mongosh":        "MongoDB shell installed with database access",
		"make":           "Make build tool installed",
		"cargo":          "Rust toolchain installed",
		"bundle":         "Ruby and Bundler installed",
		"rails":          "Ruby on Rails installed",
		"composer":       "PHP Composer installed",
	}

	seen := make(map[string]bool)
	for tool := range tools {
		if prereq, ok := prereqMap[tool]; ok {
			if !seen[prereq] {
				prereqs = append(prereqs, prereq)
				seen[prereq] = true
			}
		}
	}

	return prereqs
}

// extractTool returns the primary tool/command being used.
func extractTool(command string) string {
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return ""
	}

	tool := parts[0]

	// Handle sudo
	if tool == "sudo" && len(parts) > 1 {
		tool = parts[1]
	}

	return tool
}

// formatIntent converts an intent name to a readable string.
func formatIntent(intent string) string {
	replacements := map[string]string{
		"git-commit":     "Git version control",
		"git-branch":     "Git branching",
		"git-sync":       "Git synchronization",
		"docker-build":   "Docker image building",
		"docker-run":     "Docker container management",
		"docker-compose": "Docker Compose orchestration",
		"npm-build":      "Node.js build process",
		"npm-dev":        "Node.js development",
		"go-build":       "Go compilation",
		"go-mod":         "Go module management",
		"python-venv":    "Python environment setup",
		"kubectl-deploy": "Kubernetes deployment",
		"kubectl-debug":  "Kubernetes debugging",
		"terraform":      "Infrastructure provisioning",
		"ssh-scp":        "Remote operations",
	}

	if formatted, ok := replacements[intent]; ok {
		return formatted
	}

	// Default: replace hyphens with spaces and capitalize
	return strings.ReplaceAll(intent, "-", " ")
}
