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
2. Run the application:
   ```bash
   ./cftui
   ```

### Keybindings

- **Arrows/Vim (j,k)**: Navigate lists.
- **Enter**: Select zone / Edit record / Confirm action.
- **'a'**: Add a new DNS record.
- **'d'**: Delete the selected DNS record.
- **'esc' / Backspace**: Go back to the previous view.
- **Tab / Shift+Tab**: Navigate form fields.
- **Ctrl+C**: Quit.

## License

MIT
