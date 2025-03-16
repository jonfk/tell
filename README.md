# tell

TELL: Terminal English Language Liason

A command that takes plain english on the cli and returns the command to execute to achieve what was prompted.

## Install

Add to your shell config. It adds a command call tellme that puts the command on your shell prompt ready to execute.
```
eval "$(tell env zsh)"
```


## NOTES


## Proposed Architecture

```
tell/
├── cmd/
│   └── tell/
│       └── main.go           # Entry point, command-line parsing
│
├── internal/
│   ├── config/               # Configuration management
│   │   ├── config.go         # Load/save config
│   │   └── defaults.go       # Default configuration values
│   │
│   ├── llm/                  # LLM client implementation
│   │   ├── client.go         # API client for LLM provider
│   │   ├── prompt.go         # Prompt construction
│   │   └── response.go       # Response parsing/handling
│   │
│   ├── shell/                # Shell integration
│   │   ├── detector.go       # Auto-detect current shell
│   │   ├── integration.go    # Generate integration scripts
│   │   ├── zsh.go            # Zsh-specific functionality
│   │   ├── bash.go           # Bash-specific functionality
│   │   └── fish.go           # Fish-specific functionality
│   │
│   ├── context/              # Context gathering
│   │   ├── directory.go      # Directory listing/scanning
│   │   └── environment.go    # Environment variables
│   │
│   └── storage/              # Persistent storage
│       ├── sqlite.go         # SQLite database operations
│       └── history.go        # Command history management
│
├── pkg/                      # Public API (if needed)
│   └── models/               # Data models
│       └── command.go        # Command representation
│
├── go.mod                    # Go module definition
├── go.sum                    # Go module checksums
└── README.md                 # Documentation
```
