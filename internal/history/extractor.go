package history

import (
	"bufio"
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"time"
)

var (
	ErrHistoryNotFound = errors.New("zsh history file not found at ~/.zsh_history")
	ErrInvalidRange    = errors.New("invalid range: 'from' must be less than or equal to 'to'")
	ErrEmptyResult     = errors.New("no commands found in specified range")
	ErrUnreadableFile  = errors.New("cannot read history file")
)

// zshPattern matches zsh extended history format: : 1234567890:0;command
// Matches same lines as: grep '^: [0-9]*:[0-9]*;'
var zshPattern = regexp.MustCompile(`^: ([0-9]+):[0-9]+;(.*)$`)

// Extractor reads and parses zsh history.
type Extractor struct {
	filePath string
}

// NewExtractor creates a new history extractor using ~/.zsh_history.
func NewExtractor() (*Extractor, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	filePath := filepath.Join(home, ".zsh_history")
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, ErrHistoryNotFound
	}

	return &Extractor{filePath: filePath}, nil
}

// Extract reads history entries within the specified command number range.
// Command numbers match what you see in `history` output.
func (e *Extractor) Extract(from, to int) ([]Entry, error) {
	if from > to {
		return nil, ErrInvalidRange
	}

	file, err := os.Open(e.filePath)
	if err != nil {
		return nil, ErrUnreadableFile
	}
	defer file.Close()

	var entries []Entry
	scanner := bufio.NewScanner(file)
	commandNumber := 0

	for scanner.Scan() {
		line := scanner.Text()

		// Only process lines matching zsh history format
		matches := zshPattern.FindStringSubmatch(line)
		if matches == nil {
			continue
		}

		commandNumber++

		// Skip if outside range
		if commandNumber < from {
			continue
		}
		if commandNumber > to {
			break
		}

		ts, _ := strconv.ParseInt(matches[1], 10, 64)
		entry := Entry{
			Number:    commandNumber,
			Timestamp: time.Unix(ts, 0),
			Command:   matches[2],
			HasTime:   true,
		}

		entries = append(entries, entry)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	if len(entries) == 0 {
		return nil, ErrEmptyResult
	}

	return entries, nil
}
