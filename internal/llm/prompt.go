package llm

import (
	"strings"

	"github.com/jonfk/tell/internal/config"
)

// buildSystemPrompt builds the system prompt for the LLM
func buildSystemPrompt(cfg *config.Config) string {
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

	// Output Format
	sb.WriteString(`IMPORTANT: Return ONLY valid JSON with the following structure:

{
  "command": "The exact command to run, with proper formatting for multi-line commands if needed",
  "show_details": true,
  "details": "A more detailed explanation (2-5 lines) of how the command works, what each part does, and any important notes, pitfalls, subtleties"
}

Examples:

1. Simple command (listing files):
{
  "command": "ls -la",
  "show_details": false,
  "details": "Lists all files and directories in the current directory with detailed information."
}

2. Complex command (finding and processing files):
{
  "command": "find /path/to/search -type f -name \"*.log\" -mtime -7 | \\\n  xargs grep -l \"ERROR\" | \\\n  xargs wc -l | \\\n  sort -nr",
  "show_details": true,
  "details": "This command searches for .log files modified in the last 7 days, then filters for files containing 'ERROR', counts the lines in each file, and sorts the results by line count in descending order. The -l flag with grep only shows filenames instead of matching lines. Using xargs is more efficient than command substitution for large file sets. Be careful with file paths containing spaces."
}

Your response must contain ONLY the JSON object with no additional text, markdown, or commentary before or after it. Ensure all quotes are properly escaped and the JSON is valid and parseable.
`)

	return sb.String()
}
