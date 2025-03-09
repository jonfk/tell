package shell

import (
	"os"
	"path/filepath"
	"strings"
)

// DetectShell attempts to detect the current shell
func DetectShell() string {
	// Check SHELL environment variable
	shell := os.Getenv("SHELL")
	if shell != "" {
		// Extract the shell name from the path
		shellName := filepath.Base(shell)
		
		// Return known shell types
		switch shellName {
		case "bash":
			return "bash"
		case "zsh":
			return "zsh"
		case "fish":
			return "fish"
		}
	}
	
	// Check parent process name as fallback
	ppid := os.Getppid()
	procPath := filepath.Join("/proc", string(ppid), "comm")
	if data, err := os.ReadFile(procPath); err == nil {
		procName := strings.TrimSpace(string(data))
		switch procName {
		case "bash":
			return "bash"
		case "zsh":
			return "zsh"
		case "fish":
			return "fish"
		}
	}
	
	// Default to bash if we can't detect
	return "bash"
}
