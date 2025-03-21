package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/jonfk/tell/internal/config"
	"github.com/jonfk/tell/internal/llm"
	"github.com/jonfk/tell/internal/model"
	"github.com/jonfk/tell/internal/shellenv"
	"github.com/jonfk/tell/internal/storage"
	"github.com/spf13/cobra"
)

var (
	// Flags
	verboseFlag   bool
	formatFlag    string
	shellFlag     string
	noExplainFlag bool
	initFlag      bool
	versionFlag   bool
	limitFlag     int
	favoriteFlag  bool
	continueFlag  bool
)

const version = "0.1.0"

func main() {
	// Initially disable logging completely by using a no-op handler
	// Logging is only enabled if debugFlag is set
	slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, nil)))

	rootCmd := &cobra.Command{
		Use:   "tell",
		Short: "Terminal English Language Liaison",
		Long:  "TELL: Convert English to shell commands",
		// This PersistentPreRun sets up logging for all commands
		// Child commands with their own PersistentPreRun MUST call setupLogging()
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			setupLogging(verboseFlag)
		},
		Run: func(cmd *cobra.Command, args []string) {
			if versionFlag {
				fmt.Printf("tell version %s\n", version)
				return
			}

			if initFlag {
				config.InitConfig()
				return
			}

			cmd.Help()
		},
	}

	// Add global flags
	rootCmd.PersistentFlags().BoolVarP(&verboseFlag, "verbose", "v", false, "Enable verbose logging to stderr")
	rootCmd.Flags().BoolVarP(&initFlag, "init", "i", false, "Create default configuration file")
	rootCmd.Flags().BoolVarP(&versionFlag, "version", "", false, "Show version information")

	promptCmd := &cobra.Command{
		Use:   "prompt [text]",
		Short: "Convert natural language to shell commands",
		Long:  "Convert a natural language description into appropriate shell commands",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			// Join all args to form the prompt
			prompt := strings.Join(args, " ")

			// Load configuration
			cfg, err := config.Load()
			if err != nil {
				slog.Error("Failed to load configuration", "error", err)
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			// Check if API key is set
			if cfg.AnthropicAPIKey == "" {
				slog.Error("Anthropic API key not set")
				fmt.Fprintf(os.Stderr, "Error: Anthropic API key not set. Run 'tell config edit' to set it.\n")
				os.Exit(1)
			}

			// Initialize database
			db, err := initializeDatabase()
			if err != nil {
				slog.Error("Failed to initialize database", "error", err)
				// Don't exit if just the database fails; we can still generate the command
			}

			// Create LLM client
			client := llm.NewClient(cfg)

			// Variables for parent tracking
			var parentID sql.NullInt64
			parentID.Valid = false

			// Generate command
			var response *model.CommandResponse
			var usage *model.LLMUsage
			var genErr error

			// Handle continue flag
			if continueFlag && db != nil {
				// Get most recent successful command
				previousEntry, prevErr := db.GetMostRecentSuccessfulCommand()
				if prevErr != nil {
					slog.Error("Failed to get previous command", "error", prevErr)
					fmt.Fprintf(os.Stderr, "Error: Failed to get previous command: %v\n", prevErr)
					os.Exit(1)
				}

				slog.Debug("Continuing from previous command", "id", previousEntry.ID)
				fmt.Fprintf(os.Stderr, "Continuing from previous command: %s\n", previousEntry.Command)

				// Generate command as continuation
				response, usage, genErr = client.GenerateCommandContinuation(prompt, previousEntry)

				// Set parent ID
				parentID.Valid = true
				parentID.Int64 = previousEntry.ID
			} else {
				// Normal command generation
				response, usage, genErr = client.GenerateCommand(prompt)
			}

			// Log to database if available
			if db != nil {
				var errorMsg string
				if genErr != nil {
					errorMsg = genErr.Error()
				}

				_, dbErr := db.AddHistoryEntry(
					prompt,
					response,
					usage,
					errorMsg,
					parentID, // Include parent ID
				)

				if dbErr != nil {
					slog.Error("Failed to save to history", "error", dbErr)
				}

				// Close database connection after use
				db.Close()
			}

			// Handle command generation error after attempting to log it
			if genErr != nil {
				slog.Error("Failed to generate command", "error", genErr)
				fmt.Fprintf(os.Stderr, "Error: %v\n", genErr)
				os.Exit(1)
			}

			// Display debug info if requested
			if verboseFlag && usage != nil {
				fmt.Fprintf(os.Stderr, "Model: %s\n", usage.Model)
				fmt.Fprintf(os.Stderr, "Tokens used: input=%d, output=%d\n", usage.InputTokens, usage.OutputTokens)
			}

			// Handle output based on format
			if formatFlag == "json" {
				// Output JSON
				jsonData, err := json.Marshal(response)
				if err != nil {
					slog.Error("Failed to marshal response to JSON", "error", err)
					fmt.Fprintf(os.Stderr, "Error: %v\n", err)
					os.Exit(1)
				}
				fmt.Println(string(jsonData))
			} else {
				// Output text format
				if noExplainFlag {
					// Just print the command
					fmt.Println(response.Command)
				} else {
					// Print command and explanation
					fmt.Println(response.Command)
					fmt.Println()
					if response.ShowDetails {
						fmt.Println(response.Details)
					}
				}
			}
		},
	}

	// Add flags to prompt command
	promptCmd.Flags().StringVarP(&formatFlag, "format", "f", "text", "Output format: text|json")
	promptCmd.Flags().StringVarP(&shellFlag, "shell", "s", "auto", "Target shell: zsh|bash|fish")
	promptCmd.Flags().BoolVarP(&noExplainFlag, "no-explain", "n", false, "Skip command explanation")
	promptCmd.Flags().BoolVarP(&continueFlag, "continue", "c", false, "Continue from the most recent successful command")

	// History command
	historyCmd := &cobra.Command{
		Use:   "history [query]",
		Short: "Show command history",
		Long:  "Show command history with optional search query",
		Run: func(cmd *cobra.Command, args []string) {
			query := ""
			if len(args) > 0 {
				query = args[0]
			}

			db, err := initializeDatabase()
			if err != nil {
				slog.Error("Failed to initialize database", "error", err)
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			defer db.Close()

			var entries []model.HistoryEntry

			if query != "" {
				// Search by query
				entries, err = db.SearchHistory(query, limitFlag)
			} else {
				// List all entries (or favorites)
				entries, err = db.GetHistoryEntries(limitFlag, 0, favoriteFlag, "")
			}

			if err != nil {
				slog.Error("Failed to retrieve history", "error", err)
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			if len(entries) == 0 {
				fmt.Println("No history entries found.")
				return
			}

			// Print entries
			for _, entry := range entries {
				// Format timestamp
				timestamp := entry.Timestamp.Format("2006-01-02 15:04:05")

				// Print entry ID and timestamp
				fmt.Printf("[%d] %s", entry.ID, timestamp)

				// Add favorite indicator
				if entry.Favorite {
					fmt.Print(" ⭐")
				}
				// Add continuation indicator
				if entry.ParentID.Valid {
					fmt.Printf(" (continues from %d)", entry.ParentID.Int64)
				}
				fmt.Println()

				// Print prompt
				fmt.Printf("Prompt: %s\n", entry.Prompt)

				// Print command
				fmt.Printf("Command: %s\n", entry.Command)

				// Print separator
				fmt.Println(strings.Repeat("-", 80))
			}
		},
	}

	// Add flags to history command
	historyCmd.Flags().IntVarP(&limitFlag, "limit", "l", 10, "Maximum number of entries to show")
	historyCmd.Flags().BoolVarP(&favoriteFlag, "favorites", "f", false, "Show only favorite entries")

	// History show command
	historyShowCmd := &cobra.Command{
		Use:   "show [id]",
		Short: "Show details of a specific history entry",
		Long:  "Show complete details of a specific history entry by ID",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			// Parse ID
			id, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				slog.Error("Invalid history ID", "input", args[0], "error", err)
				fmt.Fprintf(os.Stderr, "Error: Invalid history ID: %s\n", args[0])
				os.Exit(1)
			}

			db, err := initializeDatabase()
			if err != nil {
				slog.Error("Failed to initialize database", "error", err)
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			defer db.Close()

			// Get entry by ID
			entry, err := db.GetHistoryEntry(id)
			if err != nil {
				slog.Error("Failed to retrieve history entry", "id", id, "error", err)
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			// Format output
			fmt.Printf("ID: %d\n", entry.ID)
			fmt.Printf("Time: %s\n", entry.Timestamp.Format(time.RFC1123))
			fmt.Printf("Favorite: %v\n", entry.Favorite)

			// Display parent ID if present
			if entry.ParentID.Valid {
				fmt.Printf("Continues from: %d\n", entry.ParentID.Int64)
			}

			fmt.Printf("Model: %s\n", entry.Model)
			fmt.Printf("Input Tokens: %d\n", entry.InputTokens)
			fmt.Printf("Output Tokens: %d\n", entry.OutputTokens)
			fmt.Println()
			fmt.Printf("Prompt: %s\n", entry.Prompt)
			fmt.Println()
			fmt.Printf("Command: %s\n", entry.Command)
			fmt.Println()

			if entry.Details != "" {
				fmt.Printf("Details: %s\n", entry.Details)
				fmt.Println()
			}

			if entry.ErrorMessage != "" {
				fmt.Printf("Error: %s\n", entry.ErrorMessage)
			}
		},
	}

	// History favorite command
	historyFavoriteCmd := &cobra.Command{
		Use:   "favorite [id]",
		Short: "Toggle favorite status of a history entry",
		Long:  "Mark or unmark a history entry as favorite by ID",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			// Parse ID
			id, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				slog.Error("Invalid history ID", "input", args[0], "error", err)
				fmt.Fprintf(os.Stderr, "Error: Invalid history ID: %s\n", args[0])
				os.Exit(1)
			}

			db, err := initializeDatabase()
			if err != nil {
				slog.Error("Failed to initialize database", "error", err)
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			defer db.Close()

			// Get current favorite status
			entry, err := db.GetHistoryEntry(id)
			if err != nil {
				slog.Error("Failed to retrieve history entry", "id", id, "error", err)
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			// Toggle favorite status
			newStatus := !entry.Favorite
			if err := db.SetFavorite(id, newStatus); err != nil {
				slog.Error("Failed to update favorite status", "id", id, "error", err)
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			if newStatus {
				fmt.Printf("Entry %d marked as favorite.\n", id)
			} else {
				fmt.Printf("Entry %d unmarked as favorite.\n", id)
			}
		},
	}

	// History delete command
	historyDeleteCmd := &cobra.Command{
		Use:   "delete [id]",
		Short: "Delete a history entry",
		Long:  "Delete a specific history entry by ID",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			// Parse ID
			id, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				slog.Error("Invalid history ID", "input", args[0], "error", err)
				fmt.Fprintf(os.Stderr, "Error: Invalid history ID: %s\n", args[0])
				os.Exit(1)
			}

			db, err := initializeDatabase()
			if err != nil {
				slog.Error("Failed to initialize database", "error", err)
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			defer db.Close()

			// Delete the entry
			if err := db.DeleteHistoryEntry(id); err != nil {
				slog.Error("Failed to delete history entry", "id", id, "error", err)
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			fmt.Printf("Entry %d deleted.\n", id)
		},
	}

	// Add subcommands to historyCmd
	historyCmd.AddCommand(historyShowCmd, historyFavoriteCmd, historyDeleteCmd)

	// Add subcommands
	envCmd := &cobra.Command{
		Use:   "env [shell]",
		Short: "Print shell integration script",
		Long:  "Print shell integration script for specified shell",
		Run: func(cmd *cobra.Command, args []string) {
			shell := "auto"
			if len(args) > 0 {
				shell = args[0]
			}

			script, err := shellenv.GenerateIntegrationScript(shell)
			if err != nil {
				slog.Error("Failed to generate shell integration", "error", err)
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			fmt.Println(script)
		},
	}

	configCmd := &cobra.Command{
		Use:   "config",
		Short: "Configuration management",
		Long:  "Manage tell configuration",
	}

	configEditCmd := &cobra.Command{
		Use:   "edit",
		Short: "Edit configuration file",
		Run: func(cmd *cobra.Command, args []string) {
			config.EditConfig()
		},
	}

	configShowCmd := &cobra.Command{
		Use:   "show",
		Short: "Show current configuration",
		Run: func(cmd *cobra.Command, args []string) {
			slog.Info("Showing configuration")

			cfg, err := config.Load()
			if err != nil {
				slog.Error("Failed to load configuration", "error", err)
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			// Print config with sensitive information truncated
			fmt.Println(cfg.String())
		},
	}

	configInitCmd := &cobra.Command{
		Use:   "init",
		Short: "Create default configuration file",
		Run: func(cmd *cobra.Command, args []string) {
			config.InitConfig()
		},
	}

	configCmd.AddCommand(configEditCmd, configShowCmd, configInitCmd)
	rootCmd.AddCommand(promptCmd, envCmd, configCmd, historyCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// initializeDatabase creates and initializes the SQLite database
func initializeDatabase() (*storage.DB, error) {
	db, err := storage.NewDB()
	if err != nil {
		return nil, fmt.Errorf("could not create database connection: %w", err)
	}

	if err := db.InitSchema(); err != nil {
		db.Close()
		return nil, fmt.Errorf("could not initialize database schema: %w", err)
	}

	return db, nil
}

// setupLogging configures the application logging based on verbose flag
// IMPORTANT: All commands with custom PersistentPreRun MUST call this function
// to maintain consistent logging behavior
func setupLogging(verbose bool) {
	if verbose {
		handler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})
		slog.SetDefault(slog.New(handler))
	}
}
