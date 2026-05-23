# cftui: Cloudflare TUI

> I was tired of logging in to the cloudflare dashboard every time I needed to edit DNS entries, using terraform was overkill. Avoid context switching away from the terminal. Thus: cftui came into being.

`cftui` is a fast, terminal-based user interface for managing your Cloudflare DNS records. Built with Go and the [Bubble Tea](https://github.com/charmbracelet/bubbletea) framework.

## Features

- **Zone Management**: List, add, and delete zones. View nameservers and verification keys.
- **DNS Management**: List, add, edit, and delete DNS records (A, CNAME, TXT, MX, SRV, CAA).
- **Interactive Type Picker**: Searchable list for selecting DNS record types.
- **Async Operations**: Non-blocking API calls ensure the UI stays responsive.
- **Safe Operations**: Multi-stage validation and confirmation prompts.
- **Themable**: 8 built-in color schemes.
- **Headless Mode**: Scriptable CLI commands for automation.

## Installation

### From Releases

Download the latest pre-compiled binary for your platform from the [Releases](https://github.com/devnullvoid/cloudflare-tui/releases) page.

### Go Install

```bash
go install github.com/devnullvoid/cloudflare-tui/cmd/cftui@latest
```

### AUR (Arch Linux)

```bash
paru -S cftui-bin
```

## Usage

1. Set your Cloudflare API token as an environment variable:
   ```bash
   export CLOUDFLARE_API_TOKEN=your_token_here
   ```

### API Permissions

Your Cloudflare API Token needs the following permissions:

- **DNS**: `Edit` (Required for all DNS operations)
- **Zone**: `Read` (Required to list zones)
- **Zone**: `Edit` (Optional: Only required to **add** or **delete** zones)

### Interactive TUI Mode

By default, running `cftui` without arguments launches the interactive interface.

```bash
./cftui
```

**Keybindings (Zone List)**
- **'a'**: Add a new zone.
- **'d'**: Delete the selected zone (requires name confirmation).
- **'r'**: Trigger zone activation check.
- **'i'**: View detailed zone info (nameservers, etc.).

### Headless CLI Mode

**Zone Commands**
- `cftui zones list`
- `cftui zones create <name>`
- `cftui zones delete <zone-name-or-id>`
- `cftui zones check <zone-name-or-id>`

**Record Commands**
- `cftui records list <zone-name-or-id>`
- `cftui records create <zone-name-or-id> --name <name> --type <type> [flags]`
- `cftui records update <zone-name-or-id> <record-id> --name <name> --type <type> [flags]`
- `cftui records delete <zone-name-or-id> <record-id>`

**Universal Record Flags**

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--type` | | `A` | Record type (`A`, `AAAA`, `CNAME`, `TXT`, `MX`, `SRV`, `CAA`, …) |
| `--name` | `-n` | *(required)* | Record name (e.g. `www`, `@`) |
| `--content` | `-c` | | IP, hostname, or text value (not used for SRV/CAA) |
| `--proxied` | `-p` | `false` | Proxy through Cloudflare |
| `--ttl` | | `1` | TTL in seconds (`1` = auto) |
| `--comment` | | | Optional description |

**Type-Specific Flags**

| Flag | Applies to | Description |
|------|-----------|-------------|
| `--priority` | MX, SRV | Priority value |
| `--flatten-cname` | CNAME | Flatten CNAME at zone apex |
| `--service` | SRV | Service name (e.g. `_sip`) |
| `--proto` | SRV | Protocol (e.g. `_tcp`) |
| `--weight` | SRV | Weight |
| `--port` | SRV | Port |
| `--target` | SRV | Target hostname |
| `--tag` | CAA | Tag (`issue`, `issuewild`, `iodef`) |
| `--caa-flags` | CAA | Flags (`0` or `128`) |
| `--value` | CAA | CA value (e.g. `letsencrypt.org`) |

**Global Flags**
- `--format, -f`: Output format: `table` (default), `json`, `yaml`.
- `--mock`: Use a local mock API for testing without hitting Cloudflare.
- `--theme`: Color theme. Options: `ansi`, `mocha`, `nord`, `dracula`, etc.

## Shell Completion

`cftui` supports shell completion for Bash, Zsh, Fish, and PowerShell.

```bash
# Zsh
source <(./cftui completion zsh)
```

## License

MIT
