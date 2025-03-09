package shell

import (
	"fmt"
	"log/slog"
)

// GenerateIntegrationScript generates a shell integration script for the specified shell
func GenerateIntegrationScript(shell string) (string, error) {
	// Auto-detect shell if not specified
	if shell == "auto" {
		detectedShell := DetectShell()
		slog.Info("Auto-detected shell", "shell", detectedShell)
		shell = detectedShell
	}
	
	slog.Debug("Generating integration script", "shell", shell)
	
	switch shell {
	case "zsh":
		return generateZshIntegration(), nil
	case "bash":
		return generateBashIntegration(), nil
	case "fish":
		return generateFishIntegration(), nil
	default:
		slog.Error("Unsupported shell", "shell", shell)
		return "", fmt.Errorf("unsupported shell: %s", shell)
	}
}

// generateZshIntegration generates a zsh integration script
func generateZshIntegration() string {
	return `# TELL - Terminal English Language Liaison
# ZSH integration script

tell_execute() {
  local result=$(tell prompt "$@")
  local cmd=$(echo "$result" | head -n 1)
  local explanation=$(echo "$result" | tail -n +3)

  echo "$explanation"
  print -z "$cmd"
}

# Alias for tell with command insertion
alias tellme='tell_execute'`
}

// generateBashIntegration generates a bash integration script
func generateBashIntegration() string {
	return `# TELL - Terminal English Language Liaison
# Bash integration script

tell_execute() {
  local result=$(tell prompt "$@")
  local cmd=$(echo "$result" | head -n 1)
  local explanation=$(echo "$result" | tail -n +3)

  echo "$explanation"
  history -s "$cmd"
  echo "$cmd"
}

# Alias for tell with command insertion
alias tellme='tell_execute'`
}

// generateFishIntegration generates a fish integration script
func generateFishIntegration() string {
	return `# TELL - Terminal English Language Liaison
# Fish integration script

function tell_execute
  set result (tell prompt $argv)
  set cmd (echo $result | head -n 1)
  set explanation (echo $result | tail -n +3)

  echo $explanation
  commandline $cmd
end

# Alias for tell with command insertion
alias tellme='tell_execute'`
}
