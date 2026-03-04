# cftui: Cloudflare TUI

> I was tired of logging in to the cloudflare dashboard every time I needed to edit DNS entries, using terraform was overkill. Avoid context switching away from the terminal. Thus: cftui came into being.

`cftui` is a fast, terminal-based user interface for managing your Cloudflare DNS records. Built with Go and the [Bubble Tea](https://github.com/charmbracelet/bubbletea) framework.

## Features

- **Zone Selection**: Quickly browse and search through all your Cloudflare zones.
- **DNS Management**: List, add, edit, and delete DNS records (A, CNAME, TXT, etc.).
- **Async Operations**: Non-blocking API calls ensure the UI stays responsive.
- **Safe Operations**: Validation and confirmation prompts for all destructive or modifying actions.
- **Keyboard-Centric**: Fully navigable via keyboard for maximum efficiency.

## Installation

### Prerequisites

- Go 1.21 or higher.
- A Cloudflare API Token with `Zone.DNS` (Edit) and `Zone.Zone` (Read) permissions.

### Build from Source

```bash
git clone https://github.com/devnullvoid/cloudflare-tui.git
cd cloudflare-tui
go build ./cmd/cftui
```

## Usage

1. Set your Cloudflare API token as an environment variable:
   ```bash
   export CLOUDFLARE_API_TOKEN=your_token_here
   ```

### Interactive TUI Mode

By default, running `cftui` without arguments launches the interactive interface.

```bash
./cftui
# or
./cftui tui
```

**Keybindings**
- **Arrows/Vim (j,k)**: Navigate lists.
- **Enter**: Select zone / Edit record / Confirm action.
- **'a'**: Add a new DNS record.
- **'d'**: Delete the selected DNS record.
- **'esc' / Backspace**: Go back to the previous view.
- **Tab / Shift+Tab**: Navigate form fields.
- **Ctrl+C**: Quit.

### Headless CLI Mode

`cftui` also provides powerful CLI commands for scripting and outputting structured data.

**Global Flags**
- `-f, --format string`: Output format. Options: `table` (default), `json`, `yaml`.

**Commands**
- `cftui help`: Show help text and available commands.
- `cftui zones list`: List all zones accessible by the token.
- `cftui records list <zone-name-or-id>`: List all records for a zone.
- `cftui records create <zone-name-or-id> --name <name> --content <content> [--type <type>] [--proxied]`: Create a record.
- `cftui records update <zone-name-or-id> <record-id> --name <name> --content <content> [--type <type>] [--proxied]`: Update a record.
- `cftui records delete <zone-name-or-id> <record-id>`: Delete a record.

*Example:*
```bash
# List zones in YAML format
./cftui zones list -f yaml

# List records using domain name
./cftui records list example.com

# Create a new A record using domain name
./cftui records create example.com --type A --name api --content 1.1.1.1 --proxied
```

## Shell Completion

`cftui` supports shell completion for Bash, Zsh, Fish, and PowerShell, including **dynamic completion for your Cloudflare domain names**.

To enable completion for your current session:

- **Zsh**: `source <(./cftui completion zsh)`
- **Bash**: `source <(./cftui completion bash)`
- **Fish**: `./cftui completion fish | source`

To make it permanent, add the appropriate command to your shell's configuration file (e.g., `~/.zshrc`, `~/.bashrc`).

## License


MIT
