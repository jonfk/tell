- [x] Modify the llm client to parse the expected structured output as specified in the system prompt, in the client when handling the response. It should be xml with command, description, explanation. Make sure that the output matches the expected structured output. If it doesn't return an error.
- [x] Fix shell integration and actual output of `tell prompt` to separate the command, description, etc with a delimiter such as `---TELL_DELIMITER_8e3a51f9---`
    - [x] Add alternate output format to prompt such as json with a `--json` flag
- [ ] Add some validation to the model set in config
- [ ] Log success or execution of generated commands to sqlite history
- [ ] Allow setting multiple models maybe with short name? and a switch to switch model.
- [ ] Integrate history with fzf
- [ ] Add more context to the command
- [ ] Add shell history to context
- [ ] Ensure that the shell integration gets passed the flags correctly to the program
- [ ] Test other shell integrations bash, fish
- [ ] Investigate the use of MCP to provide context
- [x] Only print logs if debug is enabled
    - [ ] ~output logs to a file that follows the XDG spec `$XDG_STATE_HOME/tell-llm/logs/` and fallback `$HOME/.local/state/tell-llm/logs/`~
    - Decided not to implement because I don't see the value of keeping this kind of log files around and the complication of log rotation.
- [ ] Save history of requests and responses in an sqlite db `$XDG_DATA_HOME/tell-llm/history.db` and fallback `$HOME/.local/share/tell-llm/history.db`
- [ ] Add estimate cost calculation and add this to history log

## Log prompt history to sqlite

Given the following project implement logging the prompt from the user and generated response from LLM to an sqlite db.
Use the following schema:

```
-- Schema for tell command history
CREATE TABLE IF NOT EXISTS command_history (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
    prompt TEXT NOT NULL,           -- User's natural language input
    command TEXT NOT NULL,          -- Generated shell command
    details TEXT,                   -- Command explanation
    show_details BOOLEAN DEFAULT 0, -- Whether details were shown

    -- LLM API information
    error_message TEXT,             -- Error message if failed
    model TEXT,                     -- LLM model used
    tokens_used INTEGER DEFAULT 0,  -- Token count

    -- For filtering and searching
    favorite BOOLEAN DEFAULT 0,     -- Allow users to mark favorite commands
);

-- Index for faster searches
CREATE INDEX IF NOT EXISTS idx_command_history_prompt ON command_history(prompt);
CREATE INDEX IF NOT EXISTS idx_command_history_command ON command_history(command);
CREATE INDEX IF NOT EXISTS idx_command_history_timestamp ON command_history(timestamp);
```
