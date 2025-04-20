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

// generateZshIntegration generates an improved zsh integration script using printf.
func generateZshIntegration() string {
	// Using standard spaces for indentation now.
	// Using printf '%s' "$result" | jq ... for robustness.
	return `# tell-zsh-integration.zsh
# ZSH integration for tell command
function tellme() {
  # Check if jq command is available
  if ! command -v jq &> /dev/null; then
    echo "Error: jq command not found. Please install jq to use this function." >&2 # Write errors to stderr
    return 1
  fi

  # Execute the tell command and capture the JSON output
  local result
  result=$(tell -f json prompt "$@")
  local tell_exit_code=$? # Capture exit code immediately

  # Check if the tell command executed successfully
  if [[ $tell_exit_code -ne 0 ]]; then
    echo "Tell command failed with exit code $tell_exit_code:" >&2
    echo "$result" >&2
    return $tell_exit_code
  fi

  # Use printf to pass the JSON to jq, which is more robust than echo
  local command
  command=$(printf '%s' "$result" | jq -r '.command // empty') # Added fallback
  local jq_command_exit_code=$?

  local show_details
  show_details=$(printf '%s' "$result" | jq -r '.show_details // "false"') # Added fallback
  local jq_details_exit_code=$?

  # Check if jq failed to parse the command or details
  if [[ $jq_command_exit_code -ne 0 || $jq_details_exit_code -ne 0 ]]; then
      echo "Error: Failed to parse JSON output from tell command using jq." >&2
      echo "Raw output:" >&2
      printf '%s\n' "$result" >&2 # Print raw output for debugging
      return 1 # Indicate failure
  fi

  # Check if the command extracted is empty (could be valid JSON but missing the field)
   if [[ -z "$command" && $jq_command_exit_code -eq 0 ]]; then
       echo "Error: Tell command returned empty command." >&2
       # Optionally print details if they exist, even if command is empty
       if [[ "$show_details" == "true" ]]; then
           local details
           details=$(printf '%s' "$result" | jq -r '.details // empty')
           if [[ $? -eq 0 && -n "$details" ]]; then
               # Use printf for potentially multi-line details
               printf '%s\n\n' "$details"
           fi
       fi
       return 1 # Indicate failure as no command was provided
   fi

  # Show details if requested
  if [[ "$show_details" == "true" ]]; then
    local details
    details=$(printf '%s' "$result" | jq -r '.details // empty')
    if [[ $? -eq 0 && -n "$details" ]]; then
        # Use printf for potentially multi-line details
        printf '%s\n\n' "$details"
    fi
  fi

  # Add the command to the Zsh command line buffer
  print -z "$command"
}`
}

// generateBashIntegration generates a bash integration script
// (Added jq check and improved READLINE handling)
func generateBashIntegration() string {
	// Using standard spaces for indentation.
	// Using printf for jq and added fallbacks similar to zsh.
	return `# tell-bash-integration.sh
# Bash integration for tell command
function tellme() {
  # Check if jq command is available
  if ! command -v jq &> /dev/null; then
    echo "Error: jq is required but not installed." >&2 # Write errors to stderr
    return 1
  fi

  # Execute the tell command and capture the JSON output
  local result
  result=$(tell -f json prompt "$@")
  local tell_exit_code=$? # Capture exit code immediately

  # Check if the tell command executed successfully
  if [[ $tell_exit_code -ne 0 ]]; then
    # echo "Tell command failed with exit code $tell_exit_code:" >&2
    # echo "$result" >&2
    return $tell_exit_code
  fi

  # Use printf to pass the JSON to jq
  local command
  command=$(printf '%s' "$result" | jq -r '.command // empty')
  local jq_command_exit_code=$?

  local show_details
  show_details=$(printf '%s' "$result" | jq -r '.show_details // "false"')
  local jq_details_exit_code=$?

   # Check if jq failed to parse the command or details
  if [[ $jq_command_exit_code -ne 0 || $jq_details_exit_code -ne 0 ]]; then
      echo "Error: Failed to parse JSON output from tell command using jq." >&2
      echo "Raw output:" >&2
      printf '%s\n' "$result" >&2 # Print raw output for debugging
      return 1 # Indicate failure
  fi

  # Check if the command extracted is empty
   if [[ -z "$command" && $jq_command_exit_code -eq 0 ]]; then
       echo "Error: Tell command returned empty command." >&2
       if [[ "$show_details" == "true" ]]; then
           local details
           details=$(printf '%s' "$result" | jq -r '.details // empty')
           if [[ $? -eq 0 && -n "$details" ]]; then
               printf '%s\n\n' "$details" # Use printf
           fi
       fi
       return 1 # Indicate failure
   fi

  # Show details if requested
  if [[ "$show_details" == "true" ]]; then
    local details
    details=$(printf '%s' "$result" | jq -r '.details // empty')
    if [[ $? -eq 0 && -n "$details" ]]; then
        printf '%s\n\n' "$details" # Use printf
    fi
  fi

  # Add command to history (Bash specific)
  history -s "$command"

  # Add command to the Readline buffer (Bash specific)
  # This makes the command appear on the prompt, ready to be edited or executed
  READLINE_LINE="$command"
  READLINE_POINT=${#READLINE_LINE} # Set cursor position to the end
}`
}
