package shellenv

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

	// TODO: Add support for more shells (e.g., fish, PowerShell, nushell)
	switch shell {
	case "zsh":
		return generateZshIntegration(), nil
	case "bash":
		return generateBashIntegration(), nil
	default:
		slog.Error("Unsupported shell", "shell", shell)
		return "", fmt.Errorf("unsupported shell: %s", shell)
	}
}

// generateZshIntegration generates a zsh integration script
func generateZshIntegration() string {
	return `# tell-zsh-integration.zsh
# ZSH integration for tell command
function tellme() {
  if ! command -v jq &> /dev/null; then
    echo "Error: jq command not found. Please install jq to use this function."
    return 1
  fi
  local result=$(tell -f json prompt "$@")
  
  if [[ $? -ne 0 ]]; then
    return $?
  fi
  
  local command=$(echo "$result" | jq -r '.command')
  local show_details=$(echo "$result" | jq -r '.show_details')
  
  if [[ "$show_details" == "true" ]]; then
    echo -e "$(echo "$result" | jq -r '.details')\n"
  fi
  
  print -z "$command"
}`
}

// generateBashIntegration generates a bash integration script
func generateBashIntegration() string {
	return `# tell-bash-integration.sh
# Bash integration for tell command
function tellme() {
  local result=$(tell -f json prompt "$@")
  
  if [[ $? -ne 0 ]]; then
    echo "Tell command failed"
    return $?
  fi
  
  # Make sure jq is available
  if ! command -v jq &> /dev/null; then
    echo "Error: jq is required but not installed"
    return 1
  fi
  
  local command=$(echo "$result" | jq -r '.command // empty')
  local show_details=$(echo "$result" | jq -r '.show_details // "false"')
  
  # Validate we got a command back
  if [[ -z "$command" ]]; then
    return 1
  fi
  
  if [[ "$show_details" == "true" ]]; then
    echo -e "$(echo "$result" | jq -r '.details // empty')\n"
  fi
  
  # Add command to history
  history -s "$command"
  
  # Directly modify the readline buffer (no temp file needed)
  READLINE_LINE="$command"
  READLINE_POINT=${#READLINE_LINE}
}`
}
