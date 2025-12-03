package processor

import (
	"strings"
	"time"

	"github.com/mrf/runbook-generator/internal/history"
)

// Workflow defines a recognized command sequence pattern.
type Workflow struct {
	Name        string
	Patterns    []string // Command prefixes to match
	Description string
}

// CommandGroup represents a logical grouping of related commands.
type CommandGroup struct {
	Title       string
	Description string
	Commands    []history.Entry
	Intent      string
}

// IntentAnalyzer groups commands into logical steps and infers purpose.
type IntentAnalyzer struct {
	workflows []Workflow
	threshold time.Duration
}

// DefaultWorkflows returns the built-in workflow patterns.
func DefaultWorkflows() []Workflow {
	return []Workflow{
		{
			Name:        "git-commit",
			Patterns:    []string{"git add", "git commit", "git push"},
			Description: "Commit and push changes",
		},
		{
			Name:        "git-branch",
			Patterns:    []string{"git checkout", "git branch", "git switch"},
			Description: "Branch management",
		},
		{
			Name:        "git-sync",
			Patterns:    []string{"git fetch", "git pull", "git merge", "git rebase"},
			Description: "Sync with remote",
		},
		{
			Name:        "docker-build",
			Patterns:    []string{"docker build", "docker tag", "docker push"},
			Description: "Build and publish container image",
		},
		{
			Name:        "docker-run",
			Patterns:    []string{"docker run", "docker exec", "docker logs"},
			Description: "Run and manage containers",
		},
		{
			Name:        "docker-compose",
			Patterns:    []string{"docker-compose", "docker compose"},
			Description: "Manage multi-container application",
		},
		{
			Name:        "npm-build",
			Patterns:    []string{"npm install", "npm run build", "npm test"},
			Description: "Install dependencies and build",
		},
		{
			Name:        "npm-dev",
			Patterns:    []string{"npm install", "npm run dev", "npm start"},
			Description: "Set up development environment",
		},
		{
			Name:        "go-build",
			Patterns:    []string{"go build", "go test", "go run"},
			Description: "Build and test Go application",
		},
		{
			Name:        "go-mod",
			Patterns:    []string{"go mod init", "go mod tidy", "go get"},
			Description: "Manage Go modules",
		},
		{
			Name:        "python-venv",
			Patterns:    []string{"python -m venv", "source", "pip install"},
			Description: "Set up Python virtual environment",
		},
		{
			Name:        "kubectl-deploy",
			Patterns:    []string{"kubectl apply", "kubectl rollout", "kubectl get"},
			Description: "Deploy to Kubernetes",
		},
		{
			Name:        "kubectl-debug",
			Patterns:    []string{"kubectl describe", "kubectl logs", "kubectl exec"},
			Description: "Debug Kubernetes resources",
		},
		{
			Name:        "terraform",
			Patterns:    []string{"terraform init", "terraform plan", "terraform apply"},
			Description: "Provision infrastructure",
		},
		{
			Name:        "ssh-scp",
			Patterns:    []string{"ssh", "scp", "rsync"},
			Description: "Remote file operations",
		},
	}
}

// NewIntentAnalyzer creates a new intent analyzer with default settings.
func NewIntentAnalyzer() *IntentAnalyzer {
	return &IntentAnalyzer{
		workflows: DefaultWorkflows(),
		threshold: 60 * time.Second,
	}
}

// WithWorkflows adds custom workflow patterns.
func (a *IntentAnalyzer) WithWorkflows(workflows []Workflow) *IntentAnalyzer {
	a.workflows = append(a.workflows, workflows...)
	return a
}

// WithThreshold sets the time gap threshold for grouping commands.
func (a *IntentAnalyzer) WithThreshold(d time.Duration) *IntentAnalyzer {
	a.threshold = d
	return a
}

// Analyze groups commands into logical steps with inferred intent.
func (a *IntentAnalyzer) Analyze(entries []history.Entry) []CommandGroup {
	if len(entries) == 0 {
		return nil
	}

	var groups []CommandGroup
	var currentGroup *CommandGroup

	for i, entry := range entries {
		tool := extractTool(entry.Command)
		intent := a.inferIntent(entry.Command)

		// Determine if we should start a new group
		startNewGroup := false

		if currentGroup == nil {
			startNewGroup = true
		} else if i > 0 {
			prevEntry := entries[i-1]
			prevTool := extractTool(prevEntry.Command)

			// Start new group if tool changes significantly
			if !areRelatedTools(prevTool, tool) {
				startNewGroup = true
			}

			// Start new group if there's a significant time gap
			if a.hasTimeGap(prevEntry, entry) {
				startNewGroup = true
			}

			// Start new group if we've matched a complete workflow
			if currentGroup.Intent != "" && currentGroup.Intent != intent {
				startNewGroup = true
			}
		}

		if startNewGroup {
			if currentGroup != nil {
				a.finalizeGroup(currentGroup)
				groups = append(groups, *currentGroup)
			}
			currentGroup = &CommandGroup{
				Commands: []history.Entry{entry},
				Intent:   intent,
			}
		} else {
			currentGroup.Commands = append(currentGroup.Commands, entry)
			// Update intent if we find a more specific one
			if intent != "" && currentGroup.Intent == "" {
				currentGroup.Intent = intent
			}
		}
	}

	// Don't forget the last group
	if currentGroup != nil {
		a.finalizeGroup(currentGroup)
		groups = append(groups, *currentGroup)
	}

	return groups
}

// inferIntent tries to match the command against known workflows.
func (a *IntentAnalyzer) inferIntent(command string) string {
	for _, workflow := range a.workflows {
		for _, pattern := range workflow.Patterns {
			if strings.HasPrefix(command, pattern) {
				return workflow.Name
			}
		}
	}
	return ""
}

// hasTimeGap checks if there's a significant time gap between entries.
func (a *IntentAnalyzer) hasTimeGap(prev, curr history.Entry) bool {
	if !prev.HasTime || !curr.HasTime {
		return false
	}
	return curr.Timestamp.Sub(prev.Timestamp) > a.threshold
}

// finalizeGroup sets the title and description for a group.
func (a *IntentAnalyzer) finalizeGroup(group *CommandGroup) {
	// Try to find a matching workflow for the title
	if group.Intent != "" {
		for _, workflow := range a.workflows {
			if workflow.Name == group.Intent {
				group.Title = workflow.Description
				break
			}
		}
	}

	// Fall back to tool-based title
	if group.Title == "" {
		group.Title = generateTitle(group.Commands)
	}

	// Generate description based on commands
	group.Description = generateDescription(group.Commands)
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

	// Handle common command wrappers
	if tool == "time" || tool == "nice" || tool == "nohup" {
		if len(parts) > 1 {
			tool = parts[1]
		}
	}

	return tool
}

// areRelatedTools checks if two tools are related.
func areRelatedTools(a, b string) bool {
	if a == b {
		return true
	}

	// Define tool families
	families := [][]string{
		{"git", "gh"},
		{"docker", "docker-compose"},
		{"kubectl", "helm", "k9s"},
		{"npm", "npx", "yarn", "pnpm"},
		{"go", "gofmt", "golangci-lint"},
		{"python", "pip", "python3", "pip3"},
		{"terraform", "tf"},
		{"aws", "awscli"},
		{"gcloud", "gsutil"},
		{"az", "azure"},
	}

	for _, family := range families {
		aInFamily := false
		bInFamily := false
		for _, tool := range family {
			if a == tool {
				aInFamily = true
			}
			if b == tool {
				bInFamily = true
			}
		}
		if aInFamily && bInFamily {
			return true
		}
	}

	return false
}

// generateTitle creates a title based on the commands in the group.
func generateTitle(commands []history.Entry) string {
	if len(commands) == 0 {
		return "Commands"
	}

	tools := make(map[string]int)
	for _, cmd := range commands {
		tool := extractTool(cmd.Command)
		if tool != "" {
			tools[tool]++
		}
	}

	// Find the most common tool
	var primaryTool string
	maxCount := 0
	for tool, count := range tools {
		if count > maxCount {
			maxCount = count
			primaryTool = tool
		}
	}

	// Generate title based on primary tool
	switch primaryTool {
	case "git":
		return "Git operations"
	case "docker", "docker-compose":
		return "Docker operations"
	case "kubectl", "helm":
		return "Kubernetes operations"
	case "npm", "yarn", "pnpm":
		return "Node.js package operations"
	case "go":
		return "Go operations"
	case "python", "pip", "python3":
		return "Python operations"
	case "terraform", "tf":
		return "Terraform operations"
	case "ssh", "scp", "rsync":
		return "Remote operations"
	case "curl", "wget":
		return "HTTP requests"
	case "cd", "ls", "mkdir", "rm", "cp", "mv":
		return "File system operations"
	default:
		if primaryTool != "" {
			return primaryTool + " operations"
		}
		return "Shell commands"
	}
}

// generateDescription creates a brief description of what the group does.
func generateDescription(commands []history.Entry) string {
	if len(commands) == 0 {
		return ""
	}

	if len(commands) == 1 {
		return ""
	}

	return ""
}
