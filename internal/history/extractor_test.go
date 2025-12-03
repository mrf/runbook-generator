package history

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExtractor_Extract_CommandNumbers(t *testing.T) {
	// Create temp zsh history file
	content := `: 1699000000:0;first
: 1699000010:0;second
: 1699000020:0;third
: 1699000030:0;fourth
: 1699000040:0;fifth
`
	extractor := createTestExtractor(t, content)

	entries, err := extractor.Extract(2, 4)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(entries) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(entries))
	}

	expected := []struct {
		num int
		cmd string
	}{
		{2, "second"},
		{3, "third"},
		{4, "fourth"},
	}

	for i, exp := range expected {
		if entries[i].Number != exp.num {
			t.Errorf("entry %d: expected number %d, got %d", i, exp.num, entries[i].Number)
		}
		if entries[i].Command != exp.cmd {
			t.Errorf("entry %d: expected command %q, got %q", i, exp.cmd, entries[i].Command)
		}
	}
}

func TestExtractor_Extract_SkipsNonMatchingLines(t *testing.T) {
	// Zsh history with continuation lines and empty lines
	content := `: 1699000000:0;first
: 1699000010:0;multi-line\
continuation here
: 1699000020:0;third

: 1699000030:0;fourth
`
	extractor := createTestExtractor(t, content)

	entries, err := extractor.Extract(1, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should get 4 commands (continuation line and empty line skipped)
	if len(entries) != 4 {
		t.Fatalf("expected 4 entries, got %d", len(entries))
	}

	expected := []string{"first", "multi-line\\", "third", "fourth"}
	for i, exp := range expected {
		if entries[i].Command != exp {
			t.Errorf("entry %d: expected %q, got %q", i, exp, entries[i].Command)
		}
		if entries[i].Number != i+1 {
			t.Errorf("entry %d: expected number %d, got %d", i, i+1, entries[i].Number)
		}
	}
}

func TestExtractor_Extract_InvalidRange(t *testing.T) {
	content := `: 1699000000:0;first
`
	extractor := createTestExtractor(t, content)

	_, err := extractor.Extract(10, 5)
	if err != ErrInvalidRange {
		t.Errorf("expected ErrInvalidRange, got %v", err)
	}
}

func TestExtractor_Extract_EmptyResult(t *testing.T) {
	content := `: 1699000000:0;first
: 1699000010:0;second
`
	extractor := createTestExtractor(t, content)

	_, err := extractor.Extract(100, 200)
	if err != ErrEmptyResult {
		t.Errorf("expected ErrEmptyResult, got %v", err)
	}
}

func TestExtractor_Extract_HasTimestamp(t *testing.T) {
	content := `: 1699000000:0;test command
`
	extractor := createTestExtractor(t, content)

	entries, err := extractor.Extract(1, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}

	if !entries[0].HasTime {
		t.Error("expected HasTime=true")
	}

	if entries[0].Timestamp.Unix() != 1699000000 {
		t.Errorf("expected timestamp 1699000000, got %d", entries[0].Timestamp.Unix())
	}
}

// createTestExtractor creates an extractor with a temp history file.
func createTestExtractor(t *testing.T, content string) *Extractor {
	t.Helper()

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, ".zsh_history")

	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	return &Extractor{filePath: tmpFile}
}
