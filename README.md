# Declaw CLI

Sandbox any MCP server in one line. Drop-in for Claude Desktop, Cursor, Windsurf, Claude Code, and every MCP client.

## Before / After

Your existing MCP config — no sandbox:

```json
{
  "mcpServers": {
    "github": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-github"],
      "env": { "GITHUB_PERSONAL_ACCESS_TOKEN": "ghp_..." }
    }
  }
}
```

Add `declaw mcp --` in front — now it runs in a Firecracker microVM:

```json
{
  "mcpServers": {
    "github": {
      "command": "declaw",
      "args": ["mcp", "--template", "node", "--network-allow", "api.github.com,github.com", "--", "npx", "-y", "@modelcontextprotocol/server-github"],
      "env": { "GITHUB_PERSONAL_ACCESS_TOKEN": "ghp_..." }
    }
  }
}
```

One prefix. No code changes. The MCP server runs inside a hardware-isolated sandbox.

## Why

MCP servers run as subprocesses with full host access — no sandbox, no permission model. Claude Desktop Extensions had a [zero-click RCE](https://layerxsecurity.com/blog/claude-desktop-extensions-rce/) rated CVSS 10/10 (LayerX, Feb 2026). Cursor had [CVE-2025-54135](https://www.tenable.com/cve/CVE-2025-54135) (CurXecute, CVSS 8.6) and [CVE-2025-54136](https://www.tenable.com/cve/CVE-2025-54136) (MCPoison, CVSS 7.2). `declaw mcp` wraps any stdio MCP server in a Firecracker microVM — the server runs unchanged, it just can't touch your machine.

## Install

**Option 1 — Download binary**

Grab the latest from [Releases](https://github.com/declaw-ai/declaw-cli/releases). Pick the right archive for your platform:

| Platform | Archive |
|----------|---------|
| macOS Apple Silicon | `declaw-darwin-arm64.tar.gz` |
| macOS Intel | `declaw-darwin-amd64.tar.gz` |
| Linux x86_64 | `declaw-linux-amd64.tar.gz` |
| Linux ARM64 | `declaw-linux-arm64.tar.gz` |
| Windows x86_64 | `declaw-windows-amd64.zip` |
| Windows ARM64 | `declaw-windows-arm64.zip` |

Extract and move to your PATH:

```bash
tar -xzf declaw-darwin-arm64.tar.gz   # or unzip on Windows
sudo mv declaw /usr/local/bin/
```

**Option 2 — Shell script**

```bash
curl -fsSL https://raw.githubusercontent.com/declaw-ai/declaw-cli/main/install.sh | sh
```

**Option 3 — Go install**

```bash
go install github.com/declaw-ai/declaw-cli/cmd/declaw@latest
```

**Option 4 — Build from source**

```bash
git clone https://github.com/declaw-ai/declaw-cli.git
cd declaw-cli && make build
```

After install, authenticate:

```bash
declaw auth login
# or
export DECLAW_API_KEY=your-api-key
```

## Client Setup

### Claude Desktop

Config path: `~/Library/Application Support/Claude/claude_desktop_config.json` (macOS) or `%APPDATA%\Claude\claude_desktop_config.json` (Windows)

```json
{
  "mcpServers": {
    "github": {
      "command": "declaw",
      "args": ["mcp", "--template", "node", "--network-allow", "api.github.com,github.com", "--", "npx", "-y", "@modelcontextprotocol/server-github"],
      "env": { "GITHUB_PERSONAL_ACCESS_TOKEN": "ghp_..." }
    }
  }
}
```

### Cursor

Config path: `~/.cursor/mcp.json`

```json
{
  "mcpServers": {
    "github": {
      "command": "declaw",
      "args": ["mcp", "--template", "node", "--network-allow", "api.github.com,github.com", "--", "npx", "-y", "@modelcontextprotocol/server-github"],
      "env": { "GITHUB_PERSONAL_ACCESS_TOKEN": "ghp_..." }
    }
  }
}
```

### Claude Code

```bash
claude mcp add github -- declaw mcp --template node --network-allow api.github.com,github.com -- npx -y @modelcontextprotocol/server-github
```

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--network-allow <hosts>` | deny-all | Comma-separated outbound hostname allowlist |
| `--template <name>` | `base` | Sandbox template |
| `--timeout <seconds>` | `86400` | Sandbox timeout (default 24h) |
| `--env KEY=VAL` | — | Environment variable to forward (repeatable) |
| `--verbose` | off | Diagnostic logging to stderr |

Network is **deny-all by default**. MCP servers that need internet access (github, brave-search, slack, etc.) require `--network-allow`. Servers that don't (filesystem, memory, time) work without it.

## How is this different?

| | Declaw | nilbox | E2B |
|---|---|---|---|
| Isolation | Firecracker microVM (hardware) | Linux VM (unspecified) | Firecracker |
| Install | CLI binary, one config edit | Desktop app, one-click | Python/TS SDK |
| Requires account | Yes (API key) | No | Yes |
| Local/offline | Not yet | Yes | No |
| Audit logs | Yes | No | No |
| Prompt injection defense | Opt-in via policy | No | No |
| PII redaction | Opt-in via policy | No | No |
| Best for | Teams needing compliance + isolation | Individual devs wanting free local sandbox | Developers building agent products |

## General Commands

```
declaw sandbox create|list|info|kill   Manage sandboxes
declaw sandbox exec|connect|files      Interact with sandboxes
declaw sandbox pause|resume            Lifecycle management
declaw template list|build|info|delete Manage templates
declaw volume create|list|get|delete   Manage volumes
declaw account info|usage|api-keys     Account management
declaw auth login|logout|status        Authentication
declaw mcp -- <command>                Sandbox an MCP server
```

Use `declaw <command> --help` for details on any command. All list/info commands support `--json` for machine-readable output.

## Links

- [Docs](https://docs.declaw.ai)
- [Website](https://declaw.ai)

## License

Apache 2.0 — see [LICENSE](LICENSE) for details.
