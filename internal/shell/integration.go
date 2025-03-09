package shell

import (
	"fmt"
	"strings"
)

// GenerateIntegrationScript generates a shell integration script for the specified shell
func GenerateIntegrationScript(shell string) (string, error) {
	// Auto-detect shell if not specified
	if shell == "auto" {
		shell = DetectShell()
	}
	
	switch shell {
	case "zsh":
		return generateZshIntegration(), nil
	case "bash":
		return generateBashIntegration(), nil
	case "fish":
		return generateFishIntegration(), nil
	default:
		return "", fmt.Errorf("unsupported shell: %s", shell)
	}
}

// generateZshIntegration generates a zsh integration script
func generateZshIntegration() string {
	var sb strings.Builder
	
	sb.WriteString("# TELL - Terminal English Language Liaison\n")
	sb.WriteString("# ZSH integration script\n\n")
	
	sb.WriteString("tell_execute() {\n")
	sb.WriteString("  local result=$(tell \"$@\")\n")
	sb.WriteString("  local cmd=$(echo \"$result\" | head -n 1)\n")
	sb.WriteString("  local explanation=$(echo \"$result\" | tail -n +3)\n\n")
	
	sb.WriteString("  echo \"$explanation\"\n")
	sb.WriteString("  print -z \"$cmd\"\n")
	sb.WriteString("}\n\n")
	
	sb.WriteString("# Alias for tell with command insertion\n")
	sb.WriteString("alias tellme='tell_execute'\n")
	
	return sb.String()
}

// generateBashIntegration generates a bash integration script
func generateBashIntegration() string {
	var sb strings.Builder
	
	sb.WriteString("# TELL - Terminal English Language Liaison\n")
	sb.WriteString("# Bash integration script\n\n")
	
	sb.WriteString("tell_execute() {\n")
	sb.WriteString("  local result=$(tell \"$@\")\n")
	sb.WriteString("  local cmd=$(echo \"$result\" | head -n 1)\n")
	sb.WriteString("  local explanation=$(echo \"$result\" | tail -n +3)\n\n")
	
	sb.WriteString("  echo \"$explanation\"\n")
	sb.WriteString("  history -s \"$cmd\"\n")
	sb.WriteString("  echo \"$cmd\"\n")
	sb.WriteString("}\n\n")
	
	sb.WriteString("# Alias for tell with command insertion\n")
	sb.WriteString("alias tellme='tell_execute'\n")
	
	return sb.String()
}

// generateFishIntegration generates a fish integration script
func generateFishIntegration() string {
	var sb strings.Builder
	
	sb.WriteString("# TELL - Terminal English Language Liaison\n")
	sb.WriteString("# Fish integration script\n\n")
	
	sb.WriteString("function tell_execute\n")
	sb.WriteString("  set result (tell $argv)\n")
	sb.WriteString("  set cmd (echo $result | head -n 1)\n")
	sb.WriteString("  set explanation (echo $result | tail -n +3)\n\n")
	
	sb.WriteString("  echo $explanation\n")
	sb.WriteString("  commandline $cmd\n")
	sb.WriteString("end\n\n")
	
	sb.WriteString("# Alias for tell with command insertion\n")
	sb.WriteString("alias tellme='tell_execute'\n")
	
	return sb.String()
}
