# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.4.0] - 2026-05-23

### Added
- **Full CLI Record Type Support**: `records create` and `records update` now support all record types available in the TUI — MX, SRV, CAA, CNAME (with flattening), A, AAAA, TXT, and more.
- **Type-Specific CLI Flags**: New flags for structured record types — `--priority` (MX/SRV), `--service`, `--proto`, `--weight`, `--port`, `--target` (SRV), `--tag`, `--caa-flags`, `--value` (CAA), `--flatten-cname` (CNAME).
- **TTL and Comment Flags**: `--ttl` and `--comment` are now supported on `records create` and `records update`.
- **Tab Completion for Record IDs**: `records update` and `records delete` now complete record IDs from the API when a zone is already provided.
- **Tab Completion for `--type`**: All valid DNS record type values are offered as completions with descriptions.
- **Zone Completion with Status**: Zone name completions now include zone status as a description.
- **Agent Skill**: Added `skills/cftui-cli/SKILL.md` — install via `npx skills add devnullvoid/cloudflare-tui` to give AI agents full knowledge of the CLI interface.

### Changed
- CLI flag `--type` on record commands no longer has a `-t` shorthand (conflicts with the global `--theme` flag).
- `--content` is no longer statically required; validation is now type-aware (SRV and CAA use structured data fields instead of content).

## [0.3.0] - 2026-03-07

### Added
- **Searchable Record Type Picker**: Select record types from a fuzzy-searchable list instead of manual text entry.
- **Advanced Record Types**: Full support for MX, SRV, and CAA records with type-specific fields.
- **Dynamic Form Fields**: Form inputs now automatically adapt based on the selected record type (e.g., Priority for MX, Service/Proto for SRV).
- **TTL Support**: Added TTL field to all record types (defaults to 1 for Auto).
- **Record Comments**: Support for adding and editing comments on DNS records.
- **CNAME Flattening**: Added support for Cloudflare's CNAME flattening setting.
- **Zone Management**: Added ability to add new zones, delete existing zones (with name confirmation), and trigger activation checks.
- **Zone Info View**: Dedicated screen to view nameservers, status, and verification keys for any zone.
- **Mock Mode**: Added `--mock` flag to run the TUI against a local mock server for safe testing.
- **CLI Zone Commands**: Mirrored zone management functionality in the headless CLI (`zones create`, `zones delete`, `zones check`).
- **AUR Support**: Added GoReleaser configuration for automated AUR (`cftui-bin`) publishing.

### Fixed
- Fixed a startup panic caused by uninitialized UI components during the first frame.
- Fixed visual filtering highlights by placing the record name at the start of the title.
- Fixed focus ordering in forms for more logical navigation.
- Fixed `goreleaser` configuration to comply with v2 standards.
- Aligned `Esc` and `q` behavior for quitting on the main screen.
- Fixed "stuck on loading" state by implementing safe error recovery transitions.

## [0.2.0] - 2026-03-05

### Added
- **Multi-Theme Support**: 8 built-in color schemes including Catppuccin Mocha, Nord, Dracula, Tokyo Night, Gruvbox, and more.
- **Default ANSI Theme**: Respects terminal native color schemes by default.
- **Interactive Filtering**: Added `/` keybinding to filter zones and records in the TUI.
- **XDG Compliance**: Logs are now stored in `XDG_STATE_HOME` (`~/.local/state/cftui/`).
- **Detailed Logging**: Support for `CFTUI_DEBUG` and `--debug` to log raw API requests and responses.
- **Quit Confirmation**: Safety prompt to prevent accidental exits.

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

[0.4.0]: https://github.com/devnullvoid/cloudflare-tui/releases/tag/v0.4.0
[0.3.0]: https://github.com/devnullvoid/cloudflare-tui/releases/tag/v0.3.0
[0.2.0]: https://github.com/devnullvoid/cloudflare-tui/releases/tag/v0.2.0
[0.1.0]: https://github.com/devnullvoid/cloudflare-tui/releases/tag/v0.1.0
