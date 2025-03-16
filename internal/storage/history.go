package storage

import (
	"database/sql"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/jonfk/tell/internal/model"
)

// AddHistoryEntry adds a new entry to the command history
func (db *DB) AddHistoryEntry(
	prompt string,
	response *model.CommandResponse,
	usage *model.LLMUsage,
	errorMsg string,
	parentID sql.NullInt64, // New parameter
) (int64, error) {
	slog.Debug("Adding history entry",
		"prompt", prompt,
		"usage", usage,
		"parentID", parentID)

	query := `
		INSERT INTO command_history (
			prompt, command, details, show_details, error_message, model, input_tokens, output_tokens, parent_id
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	var command, details, model string
	var inputTokens, outputTokens int
	var showDetails bool

	if response != nil {
		command = response.Command
		details = response.Details
		showDetails = response.ShowDetails
	}
	if usage != nil {
		model = usage.Model
		inputTokens = usage.InputTokens
		outputTokens = usage.OutputTokens
	}

	result, err := db.conn.Exec(
		query,
		prompt,
		command,
		details,
		showDetails,
		errorMsg,
		model,
		inputTokens, outputTokens,
		parentID,
	)
	if err != nil {
		return 0, fmt.Errorf("could not add history entry: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("could not get last insert ID: %w", err)
	}

	return id, nil
}

// GetHistoryEntries retrieves entries from the command history with optional filtering
func (db *DB) GetHistoryEntries(limit int, offset int, onlyFavorites bool, searchTerm string) ([]model.HistoryEntry, error) {
	var entries []model.HistoryEntry
	var params []any

	// Build the query
	query := `
		SELECT 
			id, timestamp, prompt, command, details, show_details, 
			error_message, model, input_tokens, output_tokens, favorite, parent_id
		FROM command_history
		WHERE 1=1
	`

	// Add filters
	if onlyFavorites {
		query += " AND favorite = 1"
	}

	if searchTerm != "" {
		query += " AND (prompt LIKE ? OR command LIKE ?)"
		searchParam := "%" + searchTerm + "%"
		params = append(params, searchParam, searchParam)
	}

	// Add order and limit
	query += " ORDER BY timestamp DESC LIMIT ? OFFSET ?"
	params = append(params, limit, offset)

	// Execute query
	rows, err := db.conn.Query(query, params...)
	if err != nil {
		return nil, fmt.Errorf("could not query history: %w", err)
	}
	defer rows.Close()

	// Process results
	for rows.Next() {
		var entry model.HistoryEntry
		var timestamp string

		err := rows.Scan(
			&entry.ID,
			&timestamp,
			&entry.Prompt,
			&entry.Command,
			&entry.Details,
			&entry.ShowDetails,
			&entry.ErrorMessage,
			&entry.Model,
			&entry.InputTokens,
			&entry.OutputTokens,
			&entry.Favorite,
			&entry.ParentID,
		)
		if err != nil {
			return nil, fmt.Errorf("could not scan row: %w", err)
		}

		// Parse timestamp
		entry.Timestamp, err = time.Parse("2006-01-02 15:04:05", timestamp)
		if err != nil {
			slog.Warn("Could not parse timestamp", "timestamp", timestamp, "error", err)
			// Use current time as fallback
			entry.Timestamp = time.Now()
		}

		entries = append(entries, entry)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return entries, nil
}

// GetHistoryEntry retrieves a single history entry by ID
func (db *DB) GetHistoryEntry(id int64) (*model.HistoryEntry, error) {
	query := `
		SELECT 
			id, timestamp, prompt, command, details, show_details, 
			error_message, model, input_tokens, output_tokens, favorite, parent_id
		FROM command_history
		WHERE id = ?
	`

	var entry model.HistoryEntry
	var timestamp string

	err := db.conn.QueryRow(query, id).Scan(
		&entry.ID,
		&timestamp,
		&entry.Prompt,
		&entry.Command,
		&entry.Details,
		&entry.ShowDetails,
		&entry.ErrorMessage,
		&entry.Model,
		&entry.InputTokens,
		&entry.OutputTokens,
		&entry.Favorite,
		&entry.ParentID,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no history entry found with ID %d", id)
		}
		return nil, fmt.Errorf("could not get history entry: %w", err)
	}

	// Parse timestamp
	entry.Timestamp, err = time.Parse("2006-01-02 15:04:05", timestamp)
	if err != nil {
		slog.Warn("Could not parse timestamp", "timestamp", timestamp, "error", err)
		// Use current time as fallback
		entry.Timestamp = time.Now()
	}

	return &entry, nil
}

// GetMostRecentSuccessfulCommand returns the last successful command
func (db *DB) GetMostRecentSuccessfulCommand() (*model.HistoryEntry, error) {
	query := `
		SELECT 
			id, timestamp, prompt, command, details, show_details, 
			error_message, model, input_tokens, output_tokens, favorite, parent_id
		FROM command_history
		WHERE command != '' AND error_message IS NULL OR error_message = ''
		ORDER BY timestamp DESC
		LIMIT 1
	`

	var entry model.HistoryEntry
	var timestamp string
	var parentID sql.NullInt64

	err := db.conn.QueryRow(query).Scan(
		&entry.ID,
		&timestamp,
		&entry.Prompt,
		&entry.Command,
		&entry.Details,
		&entry.ShowDetails,
		&entry.ErrorMessage,
		&entry.Model,
		&entry.InputTokens,
		&entry.OutputTokens,
		&entry.Favorite,
		&parentID,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no previous successful commands found")
		}
		return nil, fmt.Errorf("could not get most recent command: %w", err)
	}

	entry.ParentID = parentID

	// Parse timestamp
	entry.Timestamp, err = time.Parse("2006-01-02 15:04:05", timestamp)
	if err != nil {
		slog.Warn("Could not parse timestamp", "timestamp", timestamp, "error", err)
		// Use current time as fallback
		entry.Timestamp = time.Now()
	}

	return &entry, nil
}

// SetFavorite marks or unmarks a history entry as favorite
func (db *DB) SetFavorite(id int64, favorite bool) error {
	query := "UPDATE command_history SET favorite = ? WHERE id = ?"

	result, err := db.conn.Exec(query, favorite, id)
	if err != nil {
		return fmt.Errorf("could not update favorite status: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("could not get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("no history entry found with ID %d", id)
	}

	return nil
}

// DeleteHistoryEntry deletes a history entry by ID
func (db *DB) DeleteHistoryEntry(id int64) error {
	query := "DELETE FROM command_history WHERE id = ?"

	result, err := db.conn.Exec(query, id)
	if err != nil {
		return fmt.Errorf("could not delete history entry: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("could not get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("no history entry found with ID %d", id)
	}

	return nil
}

// SearchHistory searches through history entries
func (db *DB) SearchHistory(query string, limit int) ([]model.HistoryEntry, error) {
	if query == "" {
		return nil, fmt.Errorf("search query cannot be empty")
	}

	// Format search terms for LIKE queries
	searchParam := "%" + strings.Replace(query, "%", "\\%", -1) + "%"

	// Build query
	sqlQuery := `
		SELECT 
			id, timestamp, prompt, command, details, show_details, 
			error_message, model, input_tokens, output_tokens, favorite, parent_id
		FROM command_history
		WHERE prompt LIKE ? OR command LIKE ?
		ORDER BY timestamp DESC
		LIMIT ?
	`

	// Execute query
	rows, err := db.conn.Query(sqlQuery, searchParam, searchParam, limit)
	if err != nil {
		return nil, fmt.Errorf("could not search history: %w", err)
	}
	defer rows.Close()

	// Process results
	var entries []model.HistoryEntry
	for rows.Next() {
		var entry model.HistoryEntry
		var timestamp string

		err := rows.Scan(
			&entry.ID,
			&timestamp,
			&entry.Prompt,
			&entry.Command,
			&entry.Details,
			&entry.ShowDetails,
			&entry.ErrorMessage,
			&entry.Model,
			&entry.InputTokens,
			&entry.OutputTokens,
			&entry.Favorite,
			&entry.ParentID,
		)
		if err != nil {
			return nil, fmt.Errorf("could not scan row: %w", err)
		}

		// Parse timestamp
		entry.Timestamp, err = time.Parse("2006-01-02 15:04:05", timestamp)
		if err != nil {
			slog.Warn("Could not parse timestamp", "timestamp", timestamp, "error", err)
			// Use current time as fallback
			entry.Timestamp = time.Now()
		}

		entries = append(entries, entry)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return entries, nil
}
