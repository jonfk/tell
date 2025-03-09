package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var (
	// Flags
	contextFlag   bool
	debugFlag     bool
	formatFlag    string
	shellFlag     string
	executeFlag   bool
	noExplainFlag bool
	initFlag      bool
	versionFlag   bool
)

const version = "0.1.0"

func main() {
	rootCmd := &cobra.Command{
		Use:   "tell",
		Short: "Terminal English Language Liaison",
		Long:  "TELL: Convert English to shell commands",
		Run: func(cmd *cobra.Command, args []string) {
			if versionFlag {
				fmt.Printf("tell version %s\n", version)
				return
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

			// Stub implementation
			fmt.Fprintf(os.Stderr, "Not implemented yet: would process prompt: %s\n", prompt)
			fmt.Fprintf(os.Stderr, "Context: %v, Debug: %v, Format: %s, Shell: %s, Execute: %v, NoExplain: %v\n",
				contextFlag, debugFlag, formatFlag, shellFlag, executeFlag, noExplainFlag)
			os.Exit(1)
		},
	}
	
	// Add flags to prompt command
	promptCmd.Flags().BoolVarP(&contextFlag, "context", "c", false, "Include current directory contents in prompt")
	promptCmd.Flags().BoolVarP(&debugFlag, "debug", "d", false, "Show debug information (tokens used, cost)")
	promptCmd.Flags().StringVarP(&formatFlag, "format", "f", "text", "Output format: text|json")
	promptCmd.Flags().StringVarP(&shellFlag, "shell", "s", "auto", "Target shell: zsh|bash|fish")
	promptCmd.Flags().BoolVarP(&executeFlag, "execute", "e", false, "Execute command immediately")
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
			fmt.Fprintf(os.Stderr, "Not implemented yet: would generate integration for shell: %s\n", shell)
			os.Exit(1)
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
			fmt.Fprintln(os.Stderr, "Not implemented yet: would open config file in editor")
			os.Exit(1)
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
			fmt.Fprintf(os.Stderr, "Not implemented yet: would show history with query: %s\n", query)
			os.Exit(1)
		},
	}

	configCmd.AddCommand(configEditCmd)
	rootCmd.AddCommand(promptCmd, envCmd, configCmd, historyCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
