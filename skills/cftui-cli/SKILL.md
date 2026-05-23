---
name: cftui-cli
description: Use when querying or managing Cloudflare DNS records and zones via the cftui CLI — listing zones, creating/deleting zones, listing/creating/updating/deleting DNS records (A, AAAA, CNAME, TXT, MX, SRV, CAA, and more). Requires a Cloudflare API token. Use for automation, scripting, or any headless Cloudflare DNS management without the TUI.
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
- Listing, creating, updating, or deleting DNS records for a zone (all types)
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
| `cftui records create <zone> -n NAME -t TYPE [flags]` | Create a DNS record |
| `cftui records update <zone> <record-id> -n NAME -t TYPE [flags]` | Update a DNS record |
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

# Delete a zone by name or ID
cftui zones delete example.com

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

### Universal Flags (all record types)

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--type` | | `A` | Record type (`A`, `AAAA`, `CNAME`, `TXT`, `MX`, `SRV`, `CAA`, …) |
| `--name` | `-n` | *(required)* | Record name (e.g. `www`, `@`, `_sip._tcp`) |
| `--content` | `-c` | | IP, hostname, or text value (not used for SRV/CAA) |
| `--proxied` | `-p` | `false` | Proxy traffic through Cloudflare |
| `--ttl` | | `1` | TTL in seconds (`1` = auto) |
| `--comment` | | | Optional description for the record |

### Type-Specific Flags

| Flag | Applies to | Description |
|------|-----------|-------------|
| `--flatten-cname` | CNAME | Flatten CNAME at zone apex |
| `--priority` | MX, SRV | Priority value |
| `--service` | SRV | Service name (e.g. `_sip`) |
| `--proto` | SRV | Protocol (e.g. `_tcp`, `_udp`) |
| `--weight` | SRV | Weight |
| `--port` | SRV | Port number |
| `--target` | SRV | Target hostname |
| `--tag` | CAA | Tag: `issue`, `issuewild`, or `iodef` |
| `--caa-flags` | CAA | Flags: `0` (non-critical) or `128` (critical) |
| `--value` | CAA | CA value (e.g. `letsencrypt.org`) |

### A / AAAA Records

```bash
cftui records create example.com --type A --name www --content 1.2.3.4 --proxied
cftui records create example.com --type A --name www --content 1.2.3.4 --ttl 300
cftui records create example.com --type AAAA --name www --content 2001:db8::1
```

### CNAME Records

```bash
cftui records create example.com --type CNAME --name api --content target.example.com
cftui records create example.com --type CNAME --name @ --content example.github.io --flatten-cname
```

### TXT Records

```bash
cftui records create example.com --type TXT --name "_dmarc" --content "v=DMARC1; p=none"
cftui records create example.com --type TXT --name "@" --content "v=spf1 include:_spf.google.com ~all"
```

### MX Records

MX records use `--content` for the mail server hostname and `--priority` for the priority value.

```bash
cftui records create example.com --type MX --name "@" --content mail.example.com --priority 10
cftui records create example.com --type MX --name "@" --content alt1.aspmx.l.google.com --priority 20
```

### SRV Records

SRV records do not use `--content`. Use the SRV-specific flags instead.

```bash
cftui records create example.com \
  --type SRV \
  --name "_sip._tcp" \
  --service _sip \
  --proto _tcp \
  --priority 10 \
  --weight 20 \
  --port 5060 \
  --target sip.example.com
```

### CAA Records

CAA records do not use `--content`. Use `--tag`, `--value`, and optionally `--caa-flags`.

```bash
# Allow Let's Encrypt to issue certificates
cftui records create example.com --type CAA --name "@" --tag issue --value "letsencrypt.org"

# Wildcard issuance policy
cftui records create example.com --type CAA --name "@" --tag issuewild --value ";"

# Incident report URL
cftui records create example.com --type CAA --name "@" --tag iodef --value "mailto:security@example.com"
```

### Listing Records

```bash
# List records (table)
cftui records list example.com

# List records (JSON)
cftui records list example.com --format json
```

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
    "type": "MX",
    "name": "example.com",
    "content": "mail.example.com",
    "proxied": false,
    "ttl": 3600
  }
]
```

`ttl` of `1` means "Auto" (Cloudflare-managed TTL). `proxied: true` means traffic flows through Cloudflare's network.

### Updating Records

The same type-specific flags apply to `records update`. The second positional argument is the record ID.

```bash
# Update an A record's content
cftui records update example.com r1abc123 --type A --name www --content 5.6.7.8 --proxied

# Update an MX record's priority
cftui records update example.com r2def456 --type MX --name "@" --content mail.example.com --priority 5

# Add a comment to a record
cftui records update example.com r1abc123 --type A --name www --content 1.2.3.4 --comment "Primary web server"
```

### Deleting Records

```bash
cftui records delete example.com r1abc123
```

## Tab Completion

Tab completion is available for zone names, record IDs, and `--type` flag values:

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

With completion active:
- `cftui records list <TAB>` — lists your zone names with status
- `cftui records delete example.com <TAB>` — lists record IDs for that zone
- `cftui records create example.com --type <TAB>` — lists all supported record types with descriptions

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

# Upsert pattern: delete existing record then recreate
old_id=$(cftui records list example.com --format json | jq -r '.[] | select(.name == "www.example.com" and .type == "A") | .id')
cftui records delete example.com "$old_id"
cftui records create example.com --type A --name www --content 5.6.7.8 --proxied

# Check if any zones are in a pending state
cftui zones list --format json | jq '[.[] | select(.status == "pending")] | length'

# Trigger activation checks for all pending zones
cftui zones list --format json | jq -r '.[] | select(.status == "pending") | .name' | while read zone; do
  cftui zones check "$zone"
done

# List all MX records across a zone
cftui records list example.com --format json | jq '[.[] | select(.type == "MX")]'
```

## Troubleshooting

| Problem | Likely Cause | Fix |
|---------|-------------|-----|
| `CLOUDFLARE_API_TOKEN environment variable is required` | Token not set | `export CLOUDFLARE_API_TOKEN=your_token` |
| `failed to list zones` | Invalid token or missing Zone Read permission | Verify token has **Zone: Read** |
| `could not find zone with name or ID: ...` | Zone doesn't exist or token can't access it | Check zone name/ID; verify token scope |
| `failed to create record` | Missing DNS Edit permission or invalid values | Verify token has **DNS: Edit**; check flags |
| `--content is required for MX records` | Forgot `--content` for MX | Pass `--content mail.example.com` |
| `--service and --target are required for SRV records` | Missing SRV-specific flags | Add `--service` and `--target` |
| `--tag and --value are required for CAA records` | Missing CAA-specific flags | Add `--tag` and `--value` |
| Table output is garbled | Terminal doesn't support ANSI | Use `--format json` or `--format yaml` |
