package storage

import (
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

// DB handles database operations
type DB struct {
	conn *sql.DB
}

// schema is the SQLite database schema
const schema = `
-- Schema for tell command history
CREATE TABLE IF NOT EXISTS command_history (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
    prompt TEXT NOT NULL,           -- User's natural language input
    command TEXT NOT NULL,          -- Generated shell command
    details TEXT,                   -- Command explanation
    show_details BOOLEAN DEFAULT 0, -- Whether details were shown
    -- LLM API information
    error_message TEXT,             -- Error message if failed
    model TEXT,                     -- LLM model used
    input_tokens INTEGER DEFAULT 0, -- Input token count
    output_tokens INTEGER DEFAULT 0, -- Output token count
    -- For filtering and searching
    favorite BOOLEAN DEFAULT 0      -- Allow users to mark favorite commands
);
-- Index for faster searches
CREATE INDEX IF NOT EXISTS idx_command_history_prompt ON command_history(prompt);
CREATE INDEX IF NOT EXISTS idx_command_history_command ON command_history(command);
CREATE INDEX IF NOT EXISTS idx_command_history_timestamp ON command_history(timestamp);
`

// GetDBPath returns the path to the SQLite database file
func GetDBPath() (string, error) {
	// Try XDG_DATA_HOME first
	dataDir := os.Getenv("XDG_DATA_HOME")
	if dataDir == "" {
		// Fall back to HOME/.local/share
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("could not determine home directory: %w", err)
		}
		dataDir = filepath.Join(home, ".local", "share")
	}

	// Ensure the directory exists
	tellDataDir := filepath.Join(dataDir, "tell-llm")
	if err := os.MkdirAll(tellDataDir, 0755); err != nil {
		return "", fmt.Errorf("could not create data directory: %w", err)
	}

	return filepath.Join(tellDataDir, "tell.db"), nil
}

// NewDB creates a new database connection
func NewDB() (*DB, error) {
	dbPath, err := GetDBPath()
	if err != nil {
		return nil, fmt.Errorf("could not get database path: %w", err)
	}

	slog.Debug("Opening database", "path", dbPath)
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("could not open database: %w", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("could not connect to database: %w", err)
	}

	return &DB{conn: db}, nil
}

// InitSchema initializes the database schema
func (db *DB) InitSchema() error {
	slog.Debug("Initializing database schema")
	_, err := db.conn.Exec(schema)
	if err != nil {
		return fmt.Errorf("could not initialize schema: %w", err)
	}
	return nil
}

// Close closes the database connection
func (db *DB) Close() error {
	if db.conn != nil {
		return db.conn.Close()
	}
	return nil
}
