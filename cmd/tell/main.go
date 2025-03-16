package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/jonfk/tell/internal/config"
	"github.com/jonfk/tell/internal/llm"
	"github.com/jonfk/tell/internal/shellenv"
	"github.com/spf13/cobra"
)

var (
	// Flags
	debugFlag     bool
	formatFlag    string
	shellFlag     string
	noExplainFlag bool
	initFlag      bool
	versionFlag   bool
)

const version = "0.1.0"

func main() {
	// Configure slog
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})))

	rootCmd := &cobra.Command{
		Use:   "tell",
		Short: "Terminal English Language Liaison",
		Long:  "TELL: Convert English to shell commands",
		Run: func(cmd *cobra.Command, args []string) {
			if versionFlag {
				fmt.Printf("tell version %s\n", version)
				return
			}

			if initFlag {
				config.InitConfig()

			}

			cmd.Help()
		},
	}

	// Add global flags
	rootCmd.Flags().BoolVarP(&initFlag, "init", "i", false, "Create default configuration file")
	rootCmd.Flags().BoolVarP(&versionFlag, "version", "v", false, "Show version information")

	// Create prompt command
	promptCmd := &cobra.Command{
		Use:   "prompt [text]",
		Short: "Convert natural language to shell commands",
		Long:  "Convert a natural language description into appropriate shell commands",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			// Join all args to form the prompt
			prompt := strings.Join(args, " ")

			// Set debug level if debug flag is enabled
			if debugFlag {
				handler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
					Level: slog.LevelDebug,
				})
				slog.SetDefault(slog.New(handler))
			}

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

			// Create LLM client
			client := llm.NewClient(cfg)

			// Generate command
			response, err := client.GenerateCommand(prompt)
			if err != nil {
				slog.Error("Failed to generate command", "error", err)
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
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
					// Print command and explanation (now using ShortDesc and LongDesc)
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
	promptCmd.Flags().BoolVarP(&debugFlag, "debug", "d", false, "Show debug information (tokens used, cost)")
	promptCmd.Flags().StringVarP(&formatFlag, "format", "f", "text", "Output format: text|json")
	promptCmd.Flags().StringVarP(&shellFlag, "shell", "s", "auto", "Target shell: zsh|bash|fish")
	promptCmd.Flags().BoolVarP(&noExplainFlag, "no-explain", "n", false, "Skip command explanation")

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

	historyCmd := &cobra.Command{
		Use:   "history [query]",
		Short: "Show command history",
		Long:  "Show command history with optional search query",
		Run: func(cmd *cobra.Command, args []string) {
			query := ""
			if len(args) > 0 {
				query = args[0]
			}
			// TODO: Implement command history
			slog.Info("Showing command history", "query", query)
			fmt.Println("Unimplemented: command history")
			os.Exit(1)
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
