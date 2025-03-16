package model

import (
	"database/sql"
	"time"
)

// HistoryEntry represents a single entry in the command history
type HistoryEntry struct {
	ID           int64
	Timestamp    time.Time
	Prompt       string
	Command      string
	Details      string
	ShowDetails  bool
	ErrorMessage string
	Model        string
	InputTokens  int
	OutputTokens int
	Favorite     bool
	ParentID     sql.NullInt64
}
