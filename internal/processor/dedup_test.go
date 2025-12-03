package processor

import (
	"testing"
	"time"

	"github.com/mrf/runbook-generator/internal/history"
)

func TestDedup_ExactDuplicates(t *testing.T) {
	dedup := NewDedup()

	entries := []history.Entry{
		{Number: 1, Command: "ls -la"},
		{Number: 2, Command: "ls -la"},
		{Number: 3, Command: "ls -la"},
		{Number: 4, Command: "pwd"},
	}

	result := dedup.Process(entries)

	if len(result) != 2 {
		t.Errorf("expected 2 entries, got %d", len(result))
	}

	if result[0].Command != "ls -la" {
		t.Errorf("expected first command 'ls -la', got %q", result[0].Command)
	}

	if result[1].Command != "pwd" {
		t.Errorf("expected second command 'pwd', got %q", result[1].Command)
	}
}

func TestDedup_PreservesIntentionalRepetition(t *testing.T) {
	dedup := NewDedup()
	dedup.TimeGap = 30 * time.Second

	now := time.Now()

	entries := []history.Entry{
		{Number: 1, Command: "kubectl get pods", Timestamp: now, HasTime: true},
		{Number: 2, Command: "kubectl get pods", Timestamp: now.Add(45 * time.Second), HasTime: true},
		{Number: 3, Command: "kubectl get pods", Timestamp: now.Add(90 * time.Second), HasTime: true},
	}

	result := dedup.Process(entries)

	// All three should be preserved because they have significant time gaps
	if len(result) != 3 {
		t.Errorf("expected 3 entries (intentional repetition), got %d", len(result))
	}
}

func TestDedup_CollapsesQuickDuplicates(t *testing.T) {
	dedup := NewDedup()
	dedup.TimeGap = 30 * time.Second

	now := time.Now()

	entries := []history.Entry{
		{Number: 1, Command: "ls", Timestamp: now, HasTime: true},
		{Number: 2, Command: "ls", Timestamp: now.Add(2 * time.Second), HasTime: true},
		{Number: 3, Command: "ls", Timestamp: now.Add(4 * time.Second), HasTime: true},
	}

	result := dedup.Process(entries)

	// Should collapse to 1 because no significant time gap
	if len(result) != 1 {
		t.Errorf("expected 1 entry (quick duplicates collapsed), got %d", len(result))
	}
}

func TestDedup_TypoCorrection(t *testing.T) {
	dedup := NewDedup()

	entries := []history.Entry{
		{Number: 1, Command: "git stauts"},
		{Number: 2, Command: "git status"},
	}

	result := dedup.Process(entries)

	if len(result) != 1 {
		t.Errorf("expected 1 entry (typo corrected), got %d", len(result))
	}

	if result[0].Command != "git status" {
		t.Errorf("expected corrected command 'git status', got %q", result[0].Command)
	}
}

func TestDedup_CollapsesCdCommands(t *testing.T) {
	dedup := NewDedup()

	entries := []history.Entry{
		{Number: 1, Command: "cd /home/user"},
		{Number: 2, Command: "cd /home/user/projects"},
		{Number: 3, Command: "cd /home/user/projects/myapp"},
		{Number: 4, Command: "ls"},
	}

	result := dedup.Process(entries)

	if len(result) != 2 {
		t.Errorf("expected 2 entries (cd collapsed), got %d", len(result))
	}

	if result[0].Command != "cd /home/user/projects/myapp" {
		t.Errorf("expected final cd command, got %q", result[0].Command)
	}
}

func TestDedup_CollapsesExportSameVar(t *testing.T) {
	dedup := NewDedup()

	entries := []history.Entry{
		{Number: 1, Command: "export PATH=/usr/bin"},
		{Number: 2, Command: "export PATH=/usr/bin:/usr/local/bin"},
		{Number: 3, Command: "export OTHER=value"},
	}

	result := dedup.Process(entries)

	if len(result) != 2 {
		t.Errorf("expected 2 entries (same export var collapsed), got %d", len(result))
	}

	if result[0].Command != "export PATH=/usr/bin:/usr/local/bin" {
		t.Errorf("expected final PATH export, got %q", result[0].Command)
	}

	if result[1].Command != "export OTHER=value" {
		t.Errorf("expected OTHER export, got %q", result[1].Command)
	}
}

func TestDedup_PreservesDifferentCommands(t *testing.T) {
	dedup := NewDedup()

	entries := []history.Entry{
		{Number: 1, Command: "git status"},
		{Number: 2, Command: "git add ."},
		{Number: 3, Command: "git commit -m 'test'"},
		{Number: 4, Command: "git push"},
	}

	result := dedup.Process(entries)

	if len(result) != 4 {
		t.Errorf("expected 4 entries (all different), got %d", len(result))
	}
}

func TestDedup_EmptyInput(t *testing.T) {
	dedup := NewDedup()

	result := dedup.Process([]history.Entry{})

	if len(result) != 0 {
		t.Errorf("expected 0 entries for empty input, got %d", len(result))
	}
}

func TestDedup_SkipsEmptyCommands(t *testing.T) {
	dedup := NewDedup()

	entries := []history.Entry{
		{Number: 1, Command: "ls"},
		{Number: 2, Command: ""},
		{Number: 3, Command: "   "},
		{Number: 4, Command: "pwd"},
	}

	result := dedup.Process(entries)

	if len(result) != 2 {
		t.Errorf("expected 2 entries (empty skipped), got %d", len(result))
	}
}

func TestLevenshteinDistance(t *testing.T) {
	tests := []struct {
		a        string
		b        string
		expected int
	}{
		{"", "", 0},
		{"a", "", 1},
		{"", "a", 1},
		{"abc", "abc", 0},
		{"abc", "abd", 1},
		{"abc", "ab", 1},
		{"abc", "abcd", 1},
		{"kitten", "sitting", 3},
		{"git stauts", "git status", 2},
	}

	for _, tt := range tests {
		t.Run(tt.a+"->"+tt.b, func(t *testing.T) {
			result := levenshteinDistance(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("levenshteinDistance(%q, %q) = %d, want %d", tt.a, tt.b, result, tt.expected)
			}
		})
	}
}
