# tell

TELL: Terminal English Language Liason

A command that takes plain english on the cli and returns the command to execute to achieve what was prompted.

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

## Notes on how to put the result on the prompt

In zsh you can use `print -z` put some text on the prompt without executing it. But this must be done from a zsh
function or a sourced script. It cannot be executed. To do that we could expose a script when `tell zsh env` is called
which would call the tell command, take the structured output to print to the terminal and put the command using `print -z`

Something similar could be done in bash;
- you can use readline to manipulate the command line buffer:
```bash
# In a bash script
bind '"\C-x\C-r": "your command here"'
echo -n "\C-x\C-r"
```
- A more common approach is to use the shell history:
```bash

# In a bash script
echo "your command here" >> ~/.bash_history
history -r
```

Fish can use:
- `commandline -i "your command here"`. The `-i` flag inserts text at the cursor position
- `commandline "your command here"  # Replace the entire command line`

## Command line interface and flags

```
tell - Terminal English Language Liaison

Usage:
  tell [flags] <prompt>        Convert English to shell commands
  tell env [shell]             Print shell integration script for specified shell
  tell config edit             Edit configuration file
  tell history [query]         Show command history with optional search

Flags:
  -c, --context              Include current directory contents in prompt
  -d, --debug                Show debug information (tokens used, cost)
  -f, --format string        Output format: text|json (default "text")
  -s, --shell string         Target shell: zsh|bash|fish (default "auto")
  -e, --execute              Execute command immediately
  -n, --no-explain           Skip command explanation
  -i, --init                 Create default configuration file
  -v, --version              Show version information
  -h, --help                 Show help information
```


## Default System Prompt

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

## Main Command Flow (cmd/tell/main.go)

```
func main() {
    // 1. Parse CLI flags and arguments
    // 2. Load configuration
    // 3. Handle subcommands (env, config, history)
    // 4. For main command:
    //    a. Gather context if --context flag
    //    b. Prepare prompt
    //    c. Call LLM client
    //    d. Parse response
    //    e. Display result
    //    f. Log to SQLite
    //    g. Handle shell integration if needed
}
```

## Configuration

```
type Config struct {
    AnthropicAPIKey         string            `yaml:"anthropic_api_key"`
    LLMModel          string            `yaml:"llm_model"`    // Model to use
    PreferredCommands []string          `yaml:"preferred_commands"`
    ExtraInstructions []string            `yaml:"extra_instructions"`
}

func LoadConfig() (*Config, error) {
    // Try XDG_CONFIG_HOME first, then fall back to HOME/.config
    // Parse YAML into Config struct
}
```
