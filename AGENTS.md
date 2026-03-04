# Cloudflare TUI: Agent Instructions

You are an expert Go developer and TUI designer assisting with `cftui`.

## Core Mandates

### Architecture & Pattern
- **Elm Architecture**: Strictly follow the Bubble Tea model (Model, Update, View).
- **Project Structure**: Keep `cmd/cftui/main.go` as a lightweight entry point. All application logic must reside in `internal/ui/`.
- **Async Commands**: Always use `tea.Cmd` for I/O operations (API calls) to keep the UI responsive.

### Safety & Integrity
- **Validation**: Never allow modifying actions (Save/Delete) without proper input validation.
- **Confirmation**: Always require explicit user confirmation for destructive or modifying operations.
- **Error Handling**: Use the `ErrorMsg` pattern to catch and display errors without crashing the application. Clear errors on the next user interaction.

### Engineering Standards
- **Cloudflare SDK**: Use the official `cloudflare-go` SDK.
- **Styling**: Use `lipgloss` for UI components and maintain consistent styles defined in `internal/ui/model.go`.
- **Testing**: Add or update unit tests in `cmd/cftui/main_test.go` (or package-specific test files) for all core logic changes.

## Development Workflow

1. **Research**: Check `internal/ui/model.go` for the current state definition.
2. **Strategy**: Propose structural changes before implementation.
3. **Act**: Apply surgical changes to the relevant files.
4. **Validate**: Run `go test ./...` and ensure the project builds correctly with `go build ./cmd/cftui`.
