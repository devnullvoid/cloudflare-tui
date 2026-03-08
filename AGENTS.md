# cftui: Agent Instructions

You are an expert Go developer and TUI designer assisting with `cftui`.

## Architecture & Design Patterns

### 1. Unified CLI/TUI State
- All application state (API client, Logger, Config) is managed by a centralized `CLI` struct in `internal/cli/root.go`.
- Subcommands should use the global `app` instance to access shared resources.
- Use `PersistentPreRunE` in the root command for initialization logic (logging, API client setup, mock server).

### 2. Elm Architecture (Bubble Tea)
- Strictly follow the Model-Update-View pattern.
- All TUI logic resides in `internal/ui/`.
- **Performance**: Use **pointer receivers** for all methods on `Model` and list items (`ZoneItem`, `RecordItem`) to avoid struct copying.
- **Async**: Use `tea.Cmd` for all Cloudflare API calls.
- **State Recovery**: Always implement error-clearing logic in `Update()` to return users from "Loading" states back to safe list views upon any keypress after an error.

### 3. Dynamic Forms & Sub-menus
- **Dynamic Fields**: Use `initializeInputs` in `internal/ui/record.go` to rebuild form fields when the record type changes (e.g., adding Priority for MX).
- **Searchable Selectors**: Use the `PickingTypeState` pattern (a searchable `list.Model` overlay) for fields with many valid options instead of raw text input.
- **Feature Flags**: Use helper methods like `isProxiedSupported()` or `isFlattenSupported()` to hide/show UI elements based on the current record type.

### 4. Engineering Standards
- **Logging**: Use `charmbracelet/log`. Follow the XDG State Home spec (`~/.local/state/cftui/`) for default log paths. Implement detailed debug logging for all API traffic using the custom `loggingRoundTripper`.
- **Styles**: Use the centralized `Theme` struct in `internal/ui/model.go`. Ensure all new components support the built-in color schemes.
- **CLI**: Use `spf13/cobra` for routing and `spf13/viper` for configuration/env-vars. Use `lipgloss/table` for user-friendly CLI output.

## Code Quality & Safety
- **Linter**: Ensure all code passes `golangci-lint` (v2 configuration). Use `just check` to verify.
- **Error Handling**: Explicitly handle or ignore all error returns (e.g., `_ = logFile.Close()`).
- **Confirmations**: Always require user confirmation for destructive or modifying operations.
- **Name Confirmation**: For critical operations like **Zone Deletion**, require the user to type the domain name exactly to proceed.
- **Validation**: Validate required fields (Name, Content) before initiating API calls.

## Development Workflow
1. **Research**: Review `internal/ui/model.go` for state definitions and `internal/cli/root.go` for shared state.
2. **Mock Testing**: Always test UI changes using the `--mock` flag to avoid hitting production Cloudflare accounts during development.
3. **Execute**: Maintain pointer consistency. If you change a method to a pointer receiver, ensure type assertions in `Update()` are updated to pointer types.
4. **Validate**: Always run `just check` and `just build` to ensure no lint regressions, test failures, or build failures occur.
