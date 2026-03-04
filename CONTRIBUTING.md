# Contributing to cftui

Thank you for your interest in contributing to cftui!

## Getting Started

1. Fork the repository
2. Clone your fork: `git clone git@github.com:YOUR_USERNAME/cloudflare-tui.git`
3. Create a feature branch: `git checkout -b feature/your-feature-name`
4. Make your changes
5. Run tests: `just test`
6. Run linter: `just lint`
7. Commit your changes: `git commit -am 'Add some feature'`
8. Push to the branch: `git push origin feature/your-feature-name`
9. Create a Pull Request

## Development Setup

### Prerequisites
- Go 1.21 or higher
- [just](https://github.com/casey/just) (optional, for task running)
- golangci-lint (for linting)

### Building
```bash
go build ./cmd/cftui
```

### Testing
```bash
go test ./...
```

## Code Guidelines

- Follow the Elm Architecture pattern (Bubble Tea)
- Keep `cmd/cftui/main.go` lightweight; logic goes in `internal/ui/`
- Use `tea.Cmd` for async operations
- Always validate input before modifying operations
- Require confirmation for destructive actions
- Handle errors using the `ErrorMsg` pattern
- Add tests for new functionality
- Run `go fmt` before committing

## Pull Request Process

1. Update the CHANGELOG.md with your changes under `[Unreleased]`
2. Ensure all tests pass
3. Ensure the linter passes
4. Update documentation if needed
5. Reference any related issues in your PR description

## Reporting Bugs

Open an issue with:
- Clear description of the problem
- Steps to reproduce
- Expected vs actual behavior
- Your environment (OS, Go version)

## Feature Requests

Open an issue describing:
- The feature you'd like to see
- Why it would be useful
- Any implementation ideas

## Questions?

Feel free to open an issue for questions or discussion.
