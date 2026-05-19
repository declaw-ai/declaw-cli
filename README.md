# Declaw CLI

Command-line interface for [Declaw](https://declaw.ai) — security-first sandboxing for AI agents.

## Installation

Download the latest binary from [Releases](https://github.com/declaw-ai/declaw-cli/releases), or build from source:

```bash
go install github.com/declaw-ai/declaw-cli/cmd/declaw@latest
```

## Quick start

```bash
# Authenticate
declaw auth login

# Create a sandbox
declaw sandbox create --template python --timeout 300

# Run a command
declaw sandbox exec <sandbox-id> -- echo "hello from declaw"

# Interactive terminal
declaw sandbox connect <sandbox-id>

# Clean up
declaw sandbox kill <sandbox-id>
```

Set your API key via environment variable:

```bash
export DECLAW_API_KEY=your-api-key
```

## Commands

```
declaw auth login|logout|status       Manage authentication
declaw sandbox create|list|info|kill  Manage sandboxes
declaw sandbox exec|connect|files     Interact with sandboxes
declaw sandbox pause|resume           Lifecycle management
declaw template list|build|info|delete  Manage templates
declaw volume create|list|get|delete  Manage volumes
declaw account info|usage|api-keys    Account management
```

Use `declaw <command> --help` for details on any command.

All list/info commands support `--json` for machine-readable output.

## Documentation

Full reference: [docs.declaw.ai](https://docs.declaw.ai)

## License

Apache 2.0 — see [LICENSE](LICENSE) for details.
