---
name: cftui-cli
description: Use when querying or managing Cloudflare DNS records and zones via the cftui CLI — listing zones, creating/deleting zones, listing/creating/updating/deleting DNS records. Requires a Cloudflare API token. Use for automation, scripting, or any headless Cloudflare DNS management without the TUI.
license: MIT
metadata:
  author: github.com/devnullvoid/cloudflare-tui
  version: "1.0"
---

# cftui CLI

`cftui` is a terminal UI and CLI for managing Cloudflare DNS records. When invoked with a subcommand it runs non-interactively, making it suitable for scripts and AI agent workflows.

## Installation

```bash
# Via Go (recommended)
go install github.com/devnullvoid/cloudflare-tui/cmd/cftui@latest

# Via AUR (Arch Linux)
paru -S cftui-bin

# Via skill (Claude Code / any skills.sh-compatible agent)
npx skills add devnullvoid/cloudflare-tui
```

## When to Use

- Listing all zones accessible by a Cloudflare API token
- Creating or deleting Cloudflare zones
- Triggering activation checks for pending zones
- Listing, creating, updating, or deleting DNS records for a zone
- Scripting or automating DNS changes without the interactive TUI

**Not for:** Cloudflare account settings, firewall rules, Workers, Pages, or anything outside DNS/zone management — use the Cloudflare dashboard or `wrangler` CLI for those.

## Prerequisites

- `cftui` installed and on `$PATH`
- `CLOUDFLARE_API_TOKEN` environment variable set to a valid API token

### Required API Token Permissions

| Scope | Permission | Purpose |
|-------|-----------|---------|
| DNS | Edit | All DNS record operations |
| Zone | Read | List zones |
| Zone | Edit | Add or delete zones (optional) |

```bash
export CLOUDFLARE_API_TOKEN=your_token_here
```

## Quick Reference

| Command | Purpose |
|---------|---------|
| `cftui zones list` | List all zones |
| `cftui zones create <name>` | Create a new zone |
| `cftui zones delete <zone>` | Delete a zone |
| `cftui zones check <zone>` | Trigger activation check for a pending zone |
| `cftui records list <zone>` | List DNS records for a zone |
| `cftui records create <zone> -t TYPE -n NAME -c CONTENT` | Create a DNS record |
| `cftui records update <zone> <record-id> -t TYPE -n NAME -c CONTENT` | Update a DNS record |
| `cftui records delete <zone> <record-id>` | Delete a DNS record |

`<zone>` accepts either a zone name (e.g. `example.com`) or a zone ID.

## Global Flags

These flags work with every subcommand:

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--format` | `-f` | `table` | Output format: `table`, `json`, or `yaml` |
| `--theme` | `-t` | `ansi` | Color theme for table output: `ansi`, `mocha`, `nord`, `dracula`, `rose-pine`, `tokyo-night`, `gruvbox`, `everforest` |
| `--log` | | `~/.local/state/cftui/cftui.log` | Path to log file |
| `--debug` | | `false` | Enable debug logging (logs raw HTTP requests/responses) |
| `--mock` | | `false` | Use a local mock API instead of hitting Cloudflare |

**Prefer `--format json` when parsing output in scripts or agents.** Table output is for human review.

## Output Format

**JSON (`--format json`):** structured output to stdout; errors to stderr with non-zero exit.

**YAML (`--format yaml`):** YAML-serialized output to stdout.

**Table (`--format table`):** styled human-readable table to stdout.

## Zones

```bash
# List all zones (table)
cftui zones list

# List all zones (JSON — preferred for scripting)
cftui zones list --format json

# Create a new zone
cftui zones create example.com

# Delete a zone by name
cftui zones delete example.com

# Delete a zone by ID
cftui zones delete abc123def456

# Trigger an activation check for a pending zone
cftui zones check example.com
```

**JSON shape — zones list:**
```json
[
  {
    "id": "abc123def456",
    "name": "example.com",
    "status": "active",
    "paused": false,
    "name_servers": ["ns1.cloudflare.com", "ns2.cloudflare.com"]
  }
]
```

`status` is `"active"`, `"pending"`, `"initializing"`, or `"moved"`. Use `zones check` to re-trigger nameserver verification for `"pending"` zones.

## DNS Records

```bash
# List records for a zone (table)
cftui records list example.com

# List records (JSON)
cftui records list example.com --format json

# Create an A record (proxied)
cftui records create example.com --type A --name www --content 1.2.3.4 --proxied

# Create a CNAME record (not proxied)
cftui records create example.com --type CNAME --name api --content target.example.com

# Create a TXT record
cftui records create example.com --type TXT --name "_dmarc" --content "v=DMARC1; p=none"

# Update a record by ID
cftui records update example.com r1abc123 --type A --name www --content 5.6.7.8 --proxied

# Delete a record by ID
cftui records delete example.com r1abc123
```

**Record flags for create and update:**

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--type` | `-t` | `A` | DNS record type: `A`, `AAAA`, `CNAME`, `TXT`, `MX`, `SRV`, `CAA`, etc. |
| `--name` | `-n` | *(required)* | Record name (e.g. `www`, `api`, `@`) |
| `--content` | `-c` | *(required)* | Record content (IP address, hostname, or text value) |
| `--proxied` | `-p` | `false` | Proxy traffic through Cloudflare |

**JSON shape — records list:**
```json
[
  {
    "id": "r1abc123",
    "type": "A",
    "name": "www.example.com",
    "content": "1.2.3.4",
    "proxied": true,
    "ttl": 1
  },
  {
    "id": "r2def456",
    "type": "CNAME",
    "name": "api.example.com",
    "content": "target.example.com",
    "proxied": false,
    "ttl": 3600
  }
]
```

`ttl` of `1` means "Auto" (Cloudflare-managed TTL). `proxied: true` means traffic flows through Cloudflare's network.

## Shell Completion

```bash
# Bash
source <(cftui completion bash)

# Zsh
source <(cftui completion zsh)

# Fish
cftui completion fish | source

# PowerShell
cftui completion powershell | Out-String | Invoke-Expression
```

Zone name completion is supported for all `records` subcommands.

## Mock Mode

Use `--mock` to run against a local in-memory mock server — no API token needed, no real API calls made. Useful for testing workflows or demonstrating behavior.

```bash
cftui zones list --mock
cftui records list mock-zone.com --mock
```

Mock zones: `mock-zone.com` (active) and `pending-mock.io` (pending).
Mock records: `www` A `1.2.3.4` (proxied), `api` CNAME `target.mock.com` (not proxied).

## Common Agent Patterns

```bash
# Get all zone IDs as a list
cftui zones list --format json | jq -r '.[].id'

# Find a zone by name
cftui zones list --format json | jq '.[] | select(.name == "example.com")'

# Get all A records for a zone
cftui records list example.com --format json | jq '[.[] | select(.type == "A")]'

# Get the content of a specific record by name
cftui records list example.com --format json | jq -r '.[] | select(.name == "www.example.com") | .content'

# Get record IDs for all CNAME records (useful before updating/deleting)
cftui records list example.com --format json | jq -r '.[] | select(.type == "CNAME") | .id'

# Create a record and capture its ID
record_id=$(cftui records create example.com --type A --name staging --content 10.0.0.1 --format json | jq -r '.id')

# Check if any zones are in a pending state
cftui zones list --format json | jq '[.[] | select(.status == "pending")] | length'

# Trigger activation checks for all pending zones
cftui zones list --format json | jq -r '.[] | select(.status == "pending") | .name' | while read zone; do
  cftui zones check "$zone"
done
```

## Troubleshooting

| Problem | Likely Cause | Fix |
|---------|-------------|-----|
| `CLOUDFLARE_API_TOKEN environment variable is required` | Token not set | `export CLOUDFLARE_API_TOKEN=your_token` |
| `failed to list zones: ...` | Invalid token or missing Zone Read permission | Verify token has **Zone: Read** permission |
| `could not find zone with name or ID: ...` | Zone doesn't exist or token can't access it | Check zone name/ID; verify token scope |
| `failed to create record: ...` | Missing DNS Edit permission or invalid record values | Verify token has **DNS: Edit**; check `--type`, `--name`, `--content` |
| `failed to create zone: ...` | Missing Zone Edit permission | Add **Zone: Edit** permission to the token |
| Table output is garbled | Terminal doesn't support ANSI | Use `--format json` or `--format yaml` instead |
