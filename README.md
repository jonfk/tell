# tell

TELL: Terminal English Language Liason

A command that takes plain english on the cli and returns the command to execute to achieve what was prompted.

## Install

Add to your shell config. It adds a command call tellme that puts the command on your shell prompt ready to execute.
```
eval "$(tell env zsh)"
```


## NOTES

- Use a yaml config file called `tell.yaml` in `$XDG_CONFIG_HOME/tell-llm` and fallback to `$HOME/.config/tell-llm`
if it doesn't exist.
    - A few configurations I can think of:
        - system prompt
        - llm API key
        - extra_instructions that would contain
            - list of commands and recommendations on how to use them
            - projects that might be encountered, etc
- `print -z "your command here"` will put the command on the next prompt when called from within a script or command.
    - is this possible in other shells?
- Should the contents of the current directory be put into the context? Maybe it should be possible with a flag.
- The system prompt should specify the programs installed that I would prefer to use. For example: 
    - rg and fd for searching
    - common languages I program in and therefore their ecosystems
        - Notes for some ecosystems. e.g. use uv for package management in python projects
- Put a few notes on what makes a good command.
    - e.g. split by line with `\` for long commands or pipes sequences.
    - brainstorm with claude on what else could be good suggestions for the system prompt.
- Try to get structured output such as:
    - short explanation of the command(s) printed
    - command applied on the prompt
- Provide a flag that would print debug information such as the number of tokens used and the cost.
- Add logging to an sqlite db
- Default to using anthropic as the llm provider. We can add more providers later.
- debug logging should be done using log/slog
    - We should log to a log file whenever tell is called 

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
