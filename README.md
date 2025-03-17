# Tell - Terminal English Language Liaison

[![License](https://img.shields.io/github/license/jonfk/tell)](https://github.com/jonfk/tell/blob/main/LICENSE.txt)
[![Go Report Card](https://goreportcard.com/badge/github.com/jonfk/tell)](https://goreportcard.com/report/github.com/jonfk/tell)
[![Go Version](https://img.shields.io/github/go-mod/go-version/jonfk/tell)](https://github.com/jonfk/tell)
<!-- TODO: remove when making first release [![Latest Release](https://img.shields.io/github/v/release/jonfk/tell)](https://github.com/jonfk/tell/releases) -->
<!-- TODO: Add more badges as appropriate: Go version, release version, build status, etc. -->

Tell is a command-line tool that converts natural language into shell commands. Simply describe what you want to do in plain English, and Tell will generate the appropriate command to execute.

## Features

- **Natural Language to Shell Commands**: Convert plain English descriptions into executable shell commands
- **Smart Command Explanation**: Get detailed explanations of complex or obscure commands
    - Let the LLM decides on whether to show details or not, or pass `--no-explain` to suppress always.
- **Command History**: Browse, search, and manage your command history
- **Favorites**: Mark and filter your most useful command translations
- **Multi-shell Support**: Works with bash and zsh shells
    - Contributions welcomed for more shells
- **Seamless Shell Integration**: Easy shell integration that allows you to put generated commands directly on your prompt
- **Continuation Mode**: Build upon previous commands for complex operations
- **JSON Output Format**: Structured output for programmatic use

## Installation

### Prerequisites

- Go 1.21 or later
- [Just](https://github.com/casey/just) command runner (optional, for easier building)
- An Anthropic API key (for Claude API access)

### From Source

```bash
# Clone the repository
git clone https://github.com/jonfk/tell.git
cd tell

# Using Just (recommended)
just install       # Installs to ~/.local/bin/tell
# OR
just install-gopath  # Installs to $GOPATH/bin/tell

# Using Go directly
go build -o tell ./cmd/tell
cp tell /usr/local/bin/ # Or another directory in your PATH
```

### Shell Integration

For the best experience, add shell integration to your shell configuration file:

```bash
# For zsh (add to ~/.zshrc)
eval "$(tell env zsh)"

# For bash (add to ~/.bashrc)
eval "$(tell env bash)"

# For fish (add to ~/.config/fish/config.fish)
tell env fish | source
```

This adds a `tellme` command that puts the generated command directly on your shell prompt, ready to execute.

## Configuration

Tell requires an Anthropic API key to work. You can set up your configuration in one of the following ways:

### Initial Setup

```bash
tell config init    # Create the default configuration
tell config edit    # Open the configuration in your editor
```

### Configuration Options

Configuration is stored in `~/.config/tell-llm/tell.yaml` (or `$XDG_CONFIG_HOME/tell-llm/tell.yaml` if set):

```yaml
anthropic_api_key: "your_api_key_here"
llm_model: "claude-3-haiku-20240307"
preferred_commands:
  - rg
  - fd
  - find
  - grep
  - awk
  - sed
extra_instructions:
  - "Prefer using modern alternatives like ripgrep (rg) instead of grep when available"
  - "For Python projects, recommend using uv for package management"
```

You can also set your API key via the `ANTHROPIC_API_KEY` environment variable. `tell` will first check against 
the config file and then against the environment variable if none is set there.

## Usage

### Basic Usage

```bash
# Generate a command from natural language
tell prompt "find all PDF files in the current directory modified in the last 7 days"

# Generate a command with detailed explanation disabled
tell prompt --no-explain "find all PDF files in the current directory modified in the last 7 days"

# Get JSON output
tell prompt --format json "find all PDF files in the current directory modified in the last 7 days"

# Continue from your most recent command
tell prompt --continue "but only those larger than 5MB"
```

### Working with History

```bash
# View recent commands
tell history

# Search history
tell history "pdf files"

# Show only favorite commands
tell history --favorites

# View details of a specific history entry
tell history show 42

# Mark/unmark a command as favorite
tell history favorite 42

# Delete a history entry
tell history delete 42
```

### Shell Integration

The shell integration adds a `tellme` command that puts the generated command directly on your shell prompt. You 
can create an alias such as `alias t=tellme` for even shorter prompting.

```bash
# This will generate a command and place it on your prompt
tellme find all PDF files created today
```

## Examples

Here are some examples of what you can do with Tell:

```
$ tell prompt "find large log files and compress them"
find /var/log -type f -name "*.log" -size +10M -exec gzip {} \;

This command finds all files in /var/log with the .log extension that are larger than 10MB 
and compresses them using gzip. The -exec flag allows us to run gzip on each file found.

$ tell prompt "show me the disk usage sorted by size"
du -h | sort -hr | head -n 20

This command shows the disk usage of files and directories in human-readable format (-h),
sorts them by size in reverse order (largest first), and shows only the top 20 results.
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE.txt) file for details.

## Acknowledgments

- [Anthropic](https://www.anthropic.com/) for the Claude API
- [Cobra](https://github.com/spf13/cobra) for the CLI framework
