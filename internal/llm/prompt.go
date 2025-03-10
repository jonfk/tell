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
		fmt.Fprintf(&sb, "Current directory: %s\n", cwd)
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
				fmt.Fprintf(&sb, "- %s (%s, %d bytes)\n", file.Name(), fileType, info.Size())
			} else {
				fmt.Fprintf(&sb, "- %s\n", file.Name())
			}
		}
	}

	return sb.String()
}

// buildSystemPrompt builds the system prompt for the LLM
func buildSystemPrompt(cfg *config.Config, includeContext bool) string {
	var sb strings.Builder

	// Use raw string for the introduction
	sb.WriteString(`You are TELL (Terminal English Language Liaison), an expert in Unix/Linux command line tools. 
Your task is to convert natural language requests into shell commands.

`)

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

	// Use raw string for command formatting guidelines
	sb.WriteString(`Command formatting guidelines:
- Use backslashes (\) to break long commands into multiple lines for readability
- Include proper quoting for filenames and variables
- Prefer safe commands that won't accidentally destroy data
- Use modern alternatives to legacy commands when appropriate

`)

	// Output format - structured with XML tags
	sb.WriteString(`IMPORTANT: You must structure your response using the following XML format exactly, with no deviation:

<command>
The exact command to run, with proper formatting for multi-line commands if needed
</command>

<description>
A very brief one-line description (maximum 80 characters) of what the command does
</description>

<explanation>
A more detailed explanation (2-5 lines) of how the command works, what each part does, and any important notes
</explanation>

The <command> section must contain only the command to execute with no additional text.
The <description> must be extremely concise (under 80 characters) as it will be displayed in a terminal.
The <explanation> should be brief but informative, focusing on non-obvious aspects of the command.

`)

	// Add directory context if includeContext is true
	if includeContext {
		sb.WriteString("Current directory context:\n")
		sb.WriteString(gatherDirectoryContext())
		sb.WriteString("\n")
	}

	return sb.String()
}
