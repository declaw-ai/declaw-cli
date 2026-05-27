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

Add `declaw mcp` in front — now it runs in a Firecracker microVM:

```json
{
  "mcpServers": {
    "github": {
      "command": "declaw",
      "args": ["mcp", "--env", "GITHUB_PERSONAL_ACCESS_TOKEN", "--network-allow", "registry.npmjs.org,api.github.com,github.com,codeload.github.com", "--", "npx", "-y", "@modelcontextprotocol/server-github"],
      "env": { "GITHUB_PERSONAL_ACCESS_TOKEN": "ghp_..." }
    }
  }
}
```

The MCP server runs inside a hardware-isolated sandbox. Only the environment variables you explicitly forward with `--env` reach the sandbox, and network is deny-all unless you allowlist specific hosts.

## Why

MCP servers run as subprocesses with full host access — no sandbox, no permission model. Claude Desktop Extensions had a [zero-click RCE](https://layerxsecurity.com/blog/claude-desktop-extensions-rce/) rated CVSS 10/10 (LayerX, Feb 2026). Cursor had [CVE-2025-54135](https://www.tenable.com/cve/CVE-2025-54135) (CurXecute, CVSS 9.8) and [CVE-2025-54136](https://www.tenable.com/cve/CVE-2025-54136) (MCPoison, CVSS 8.8). `declaw mcp` wraps any stdio MCP server in a Firecracker microVM — the server runs unchanged, it just can't touch your machine.

## Install

**Option 1 — Download binary**

Grab the latest from [Releases](https://github.com/declaw-ai/declaw-cli/releases). Pick the right binary for your platform:

| Platform | Binary |
|----------|--------|
| macOS Apple Silicon | `declaw-darwin-arm64` |
| macOS Intel | `declaw-darwin-amd64` |
| Linux x86_64 | `declaw-linux-amd64` |
| Linux ARM64 | `declaw-linux-arm64` |
| Windows x86_64 | `declaw-windows-amd64.exe` |
| Windows ARM64 | `declaw-windows-arm64.exe` |

Download and move to your PATH:

```bash
chmod +x declaw-darwin-arm64
sudo mv declaw-darwin-arm64 /usr/local/bin/declaw
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

After install, sign up and authenticate:

```bash
# 1. Create a free account at https://console.declaw.ai
# 2. Copy your API key from the dashboard
# 3. Authenticate:
declaw auth login
```

## Client Setup

### Claude Desktop

Config path: `~/Library/Application Support/Claude/claude_desktop_config.json` (macOS) or `%APPDATA%\Claude\claude_desktop_config.json` (Windows)

```json
{
  "mcpServers": {
    "github": {
      "command": "declaw",
      "args": ["mcp", "--env", "GITHUB_PERSONAL_ACCESS_TOKEN", "--network-allow", "registry.npmjs.org,api.github.com,github.com,codeload.github.com", "--", "npx", "-y", "@modelcontextprotocol/server-github"],
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
      "args": ["mcp", "--env", "GITHUB_PERSONAL_ACCESS_TOKEN", "--network-allow", "registry.npmjs.org,api.github.com,github.com,codeload.github.com", "--", "npx", "-y", "@modelcontextprotocol/server-github"],
      "env": { "GITHUB_PERSONAL_ACCESS_TOKEN": "ghp_..." }
    }
  }
}
```

### Windsurf

Config path: `~/.codeium/windsurf/mcp_config.json` — same JSON structure as above.

### Claude Code

```bash
claude mcp add github -e GITHUB_PERSONAL_ACCESS_TOKEN=ghp_... -- declaw mcp --env GITHUB_PERSONAL_ACCESS_TOKEN --network-allow registry.npmjs.org,api.github.com,github.com,codeload.github.com -- npx -y @modelcontextprotocol/server-github
```

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--network-allow <hosts>` | deny-all | Comma-separated outbound hostname allowlist |
| `--template <name>` | `mcp-server` | Sandbox template (includes Node.js + Python) |
| `--timeout <seconds>` | `86400` | Sandbox timeout (default 24h) |
| `--env KEY` or `--env KEY=VAL` | — | Environment variable to forward (repeatable). `KEY` reads from host env; `KEY=VAL` sets explicitly. |
| `--verbose` | off | Diagnostic logging to stderr |

Network is **deny-all by default**. MCP servers that connect to external APIs (GitHub, Slack, Brave Search, etc.) need `--network-allow` to reach their endpoints. This is the key security property: credentials passed to the server can only reach hosts you explicitly permit.

## Custom dependencies

The default `mcp-server` template includes Node.js and Python, which covers most MCP servers. If your server needs additional system packages (e.g., `ffmpeg`, `chromium`, native libraries), build a custom template:

```bash
# Create a Dockerfile
echo 'FROM declaw/mcp-server:latest
RUN apt-get update && apt-get install -y ffmpeg' > Dockerfile

# Build it (returns a template ID)
declaw template build --dockerfile Dockerfile

# Use the template ID from the build output
declaw mcp --template <template-id> -- your-server-command
```

See `declaw template build --help` for details.

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
