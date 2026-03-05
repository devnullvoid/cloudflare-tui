# cftui: Agent Instructions

You are an expert Go developer and TUI designer assisting with `cftui`.

## Architecture & Design Patterns

### 1. Unified CLI/TUI State
- All application state (API client, Logger, Config) is managed by a centralized `CLI` struct in `internal/cli/root.go`.
- Subcommands should use the global `app` instance to access shared resources.
- Use `PersistentPreRunE` in the root command for initialization logic (logging, API client setup).

### 2. Elm Architecture (Bubble Tea)
- Strictly follow the Model-Update-View pattern.
- All TUI logic resides in `internal/ui/`.
- **Performance**: Use **pointer receivers** for all methods on `Model` and list items (`ZoneItem`, `RecordItem`) to avoid struct copying.
- **Async**: Use `tea.Cmd` for all Cloudflare API calls.

### 3. Engineering Standards
- **Logging**: Use `charmbracelet/log`. Follow the XDG State Home spec (`~/.local/state/cftui/`) for default log paths. Implement detailed debug logging for all API traffic.
- **Styles**: Use the centralized `Theme` struct in `internal/ui/model.go`. Ensure all new components support the built-in color schemes (Mocha, Nord, Dracula, etc.).
- **CLI**: Use `spf13/cobra` for routing and `spf13/viper` for configuration/env-vars. Use `lipgloss/table` for user-friendly CLI output.

## Code Quality & Safety
- **Linter**: Ensure all code passes `golangci-lint`. Use `just check` to verify.
- **Error Handling**: Explicitly handle or ignore all error returns (e.g., `_ = logFile.Close()`).
- **Confirmations**: Always require user confirmation for destructive or modifying operations (Delete/Save/Quit).
- **Validation**: Validate all required fields (Type, Name, Content) before initiating API calls.

## Development Workflow
1. **Research**: Review `internal/ui/model.go` for state definitions and `internal/cli/root.go` for shared state.
2. **Execute**: Maintain pointer consistency. If you change a method to a pointer receiver, ensure type assertions in `Update()` are updated to pointer types.
3. **Validate**: Always run `just check` and `just build` to ensure no lint regressions or build failures occur.
