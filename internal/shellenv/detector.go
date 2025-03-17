package shellenv

import (
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// DetectShell attempts to detect the current shell
func DetectShell() string {
	// Check SHELL environment variable
	shell := os.Getenv("SHELL")
	if shell != "" {
		// Extract the shell name from the path
		shellName := filepath.Base(shell)

		slog.Debug("Detected shell from SHELL env var", "path", shell, "name", shellName)

		// Return known shell types
		switch shellName {
		case "bash":
			return "bash"
		case "zsh":
			return "zsh"
		}
	}

	// Check parent process name as fallback
	ppid := os.Getppid()
	procPath := filepath.Join("/proc", strconv.Itoa(ppid), "comm")
	if data, err := os.ReadFile(procPath); err == nil {
		procName := strings.TrimSpace(string(data))
		slog.Debug("Detected shell from parent process", "ppid", ppid, "name", procName)
		switch procName {
		case "bash":
			return "bash"
		case "zsh":
			return "zsh"
		}
	} else {
		slog.Debug("Failed to read parent process info", "error", err)
	}

	slog.Info("Could not detect shell, defaulting to bash")
	// Default to bash if we can't detect
	return "bash"
}
