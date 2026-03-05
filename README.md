# cftui: Cloudflare TUI

> I got tired of logging in to the Cloudflare dashboard for every DNS edit, and Terraform felt like overkill. I built cftui to make DNS changes fast, without context switching or heavy config/state management.

`cftui` is a fast, terminal-based user interface for managing your Cloudflare DNS records. Built with Go and the [Bubble Tea](https://github.com/charmbracelet/bubbletea) framework.

## Features

- **Zone Selection**: Quickly browse and search through all your Cloudflare zones.
- **DNS Management**: List, add, edit, and delete DNS records (A, CNAME, TXT, etc.).
- **Async Operations**: Non-blocking API calls ensure the UI stays responsive.
- **Safe Operations**: Validation and confirmation prompts for all destructive or modifying actions.
- **Keyboard-Centric**: Fully navigable via keyboard for maximum efficiency.
- **Themable**: Multiple built-in color schemes (Catppuccin, Nord, Dracula, Gruvbox, etc.).
- **Headless Mode**: Scriptable CLI commands for automation.

## Installation

### From Releases

Download the latest pre-compiled binary for your platform from the [Releases](https://github.com/devnullvoid/cloudflare-tui/releases) page.

### Go Install

```bash
go install github.com/devnullvoid/cloudflare-tui/cmd/cftui@latest
```

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
- **'q'**: Quit (with confirmation).
- **'/'**: Search/Filter lists.
- **Ctrl+C**: Force Quit.

### Headless CLI Mode

`cftui` also provides powerful CLI commands for scripting and outputting structured data.

**Global Flags**

- `-f, --format string`: Output format. Options: `table` (default), `json`, `yaml`.
- `-t, --theme string` : Color theme. Options: `ansi` (default), `mocha`, `nord`, `dracula`, `rose-pine`, `tokyo-night`, `gruvbox`, `everforest`.
- `--log string`      : Path to log file (default: `~/.config/cftui/cftui.log` on Linux).
- `--debug`           : Enable debug logging (verbose).

**Environment Variables**

- `CLOUDFLARE_API_TOKEN`: Your Cloudflare API token (Required).
- `CFTUI_THEME`: Set the default theme (e.g., `mocha`, `nord`).
- `CFTUI_LOG`: Set the log file path.
- `CFTUI_DEBUG`: Set to `true` to enable debug logging.

*Example:*

```bash
# List records using domain name
./cftui records list example.com

# Run TUI with verbose debug logging
./cftui --debug --log ./debug.log
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
