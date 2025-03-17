# more LLM backends

- [ ] Add some validation to the model set in config
- [ ] Add per model configuration
    - Maybe profiles? switchable with a cli switch
- [ ] Allow setting multiple models maybe with short name? and a switch to switch model.
- [ ] Add additional LLM backends
- [ ] Use `llm` as a possible backend

# Shell integration

- [ ] Move the shell integration scripts into separate files and embed them with https://pkg.go.dev/embed.
    - That would allow users to copy or modify the script themselves if they want. 
- [ ] Ensure that the shell integration gets passed the flags correctly to the program
- [ ] Test other shell integrations bash, fish

# History

- [ ] Log success or execution of generated commands to sqlite history
    - Can use post execution hook for than in zsh. Add a subcommand to log that to history.
- [ ] Add estimate cost calculation and add this to history log

# History UI

- [ ] Integrate history with fzf?
- [ ] Add a better UI for displaying history
    - See https://github.com/charmbracelet/bubbletea and https://charm.sh/
- [ ] history command should return in reverse chronological order

# More context

- [ ] Add more context to the command
- [ ] Add shell history to context
- [ ] Investigate the use of MCP to provide context
- [ ] extra prompts for ecosystem preferences?
