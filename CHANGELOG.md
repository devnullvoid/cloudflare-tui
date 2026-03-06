# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- **Searchable Record Type Picker**: Select record types from a fuzzy-searchable list instead of manual text entry.
- **Advanced Record Types**: Full support for MX, SRV, and CAA records with type-specific fields.
- **Dynamic Form Fields**: Form inputs now automatically adapt based on the selected record type (e.g., Priority for MX, Service/Proto for SRV).
- **Multi-Theme Support**: 8 built-in color schemes including Catppuccin Mocha, Nord, Dracula, Tokyo Night, Gruvbox, and more.
- **Default ANSI Theme**: Respects terminal native color schemes by default.
- **Interactive Filtering**: Added `/` keybinding to filter zones and records in the TUI.
- **XDG Compliance**: Logs are now stored in `XDG_STATE_HOME` (`~/.local/state/cftui/`).
- **Detailed Logging**: Support for `CFTUI_DEBUG` and `--debug` to log raw API requests and responses.
- **Quit Confirmation**: Safety prompt to prevent accidental exits.

### Fixed
- Fixed a startup panic caused by uninitialized UI components during the first frame.
- Fixed visual filtering highlights by placing the record name at the start of the title.
- Fixed focus ordering in forms for more logical navigation.
- Fixed `goreleaser` configuration to comply with v2 standards.

## [0.1.0] - 2026-03-04

### Added
- Interactive TUI mode for managing Cloudflare DNS records
- Zone selection and browsing
- DNS record management (list, add, edit, delete)
- Support for A, CNAME, TXT, and other DNS record types
- Headless CLI mode with structured output (table, JSON, YAML)
- Shell completion for Bash, Zsh, Fish, and PowerShell
- Dynamic completion for Cloudflare domain names
- Async operations for responsive UI
- Input validation and confirmation prompts
- Keyboard-centric navigation

[0.1.0]: https://github.com/devnullvoid/cloudflare-tui/releases/tag/v0.1.0
