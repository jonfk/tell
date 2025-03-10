package llm

import (
	"fmt"
	"os"
	"strings"

	"github.com/jonfk/tell/internal/config"
)

// gatherDirectoryContext gathers information about the current directory
func gatherDirectoryContext() string {
	var sb strings.Builder
	
	// Get current directory
	cwd, err := os.Getwd()
	if err == nil {
		sb.WriteString(fmt.Sprintf("Current directory: %s\n", cwd))
	}
	
	// List files in current directory
	files, err := os.ReadDir(".")
	if err == nil {
		sb.WriteString("Directory contents:\n")
		for _, file := range files {
			info, err := file.Info()
			if err == nil {
				// Add file info: name, size, and whether it's a directory
				fileType := "file"
				if file.IsDir() {
					fileType = "dir"
				}
				sb.WriteString(fmt.Sprintf("- %s (%s, %d bytes)\n", file.Name(), fileType, info.Size()))
			} else {
				sb.WriteString(fmt.Sprintf("- %s\n", file.Name()))
			}
		}
	}
	
	return sb.String()
}

// buildSystemPrompt builds the system prompt for the LLM
func buildSystemPrompt(cfg *config.Config, includeContext bool) string {
	var sb strings.Builder

	sb.WriteString("You are TELL (Terminal English Language Liaison), an expert in Unix/Linux command line tools. ")
	sb.WriteString("Your task is to convert natural language requests into shell commands.\n\n")

	// Add preferred commands
	if len(cfg.PreferredCommands) > 0 {
		sb.WriteString("Preferred commands: ")
		sb.WriteString(strings.Join(cfg.PreferredCommands, ", "))
		sb.WriteString("\n\n")
	}

	// Add extra instructions
	if len(cfg.ExtraInstructions) > 0 {
		sb.WriteString("Additional guidelines:\n")
		for _, instruction := range cfg.ExtraInstructions {
			sb.WriteString("- ")
			sb.WriteString(instruction)
			sb.WriteString("\n")
		}
		sb.WriteString("\n")
	}

	// Command formatting guidelines
	sb.WriteString("Command formatting guidelines:\n")
	sb.WriteString("- Use backslashes (\\) to break long commands into multiple lines for readability\n")
	sb.WriteString("- Include proper quoting for filenames and variables\n")
	sb.WriteString("- Prefer safe commands that won't accidentally destroy data\n")
	sb.WriteString("- Use modern alternatives to legacy commands when appropriate\n\n")

	// Output format
	sb.WriteString("Your response should be structured as follows:\n")
	sb.WriteString("1. First line: The exact command to run, with no additional text\n")
	sb.WriteString("2. After a blank line, provide a brief explanation of what the command does\n\n")

	// Add directory context if includeContext is true
	if includeContext {
		sb.WriteString("Current directory context:\n")
		sb.WriteString(gatherDirectoryContext())
		sb.WriteString("\n")
	}

	return sb.String()
}
