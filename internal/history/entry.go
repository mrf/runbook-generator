package history

import "time"

// Entry represents a single command from zsh history.
type Entry struct {
	Number    int       // Command number (matches `history` output)
	Timestamp time.Time // When the command was executed
	Command   string    // The command itself
	HasTime   bool      // Whether timestamp was parsed successfully
}
