# more LLM backends

- [ ] Add some validation to the model set in config
- [ ] Add per model configuration
    - Maybe profiles? switchable with a cli switch
- [ ] Allow setting multiple models maybe with short name? and a switch to switch model.
- [ ] Add additional LLM backends
- [ ] Use `llm` as a possible backend

# Shell integration / more info in history

- [ ] Log success or execution of generated commands to sqlite history
    - Can use post execution hook for than in zsh. Add a subcommand to log that to history.
- [ ] Integrate history with fzf
- [ ] Ensure that the shell integration gets passed the flags correctly to the program
- [ ] Test other shell integrations bash, fish
- [ ] Add estimate cost calculation and add this to history log
- [ ] history command should return in reverse chronological order

# More context

- [ ] Add more context to the command
- [ ] Add shell history to context
- [ ] Investigate the use of MCP to provide context
- [ ] extra prompts for ecosystem preferences?
