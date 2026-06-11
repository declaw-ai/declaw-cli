# Declaw OPA Custom-Policy — CLI Examples

This directory contains sample Rego policy files and a runnable shell script
that demonstrate the OPA custom-policy flags on `declaw sandbox create`.

---

## The Three Flags

| Flag | Argument | Effect |
|---|---|---|
| `--opa-policy <file>` | Path to a `.rego` file | Reads the file and sends its contents as `custom_policy.inline_rego`. Sets `custom_policy.enabled = true`. Use for your **primary** policy package. |
| `--opa-policy-module <file>` | Path to a `.rego` file (repeatable) | Reads the file and appends its contents to `custom_policy.inline_modules`. Each file is an **independent Rego package**. Repeat the flag for each additional module. |
| `--opa-fail-closed` | (boolean flag, no argument) | Sets `custom_policy.default_deny = true`. When the OPA evaluator is unreachable or encounters an error, the action is **denied** instead of allowed. Omit this flag to fail-open (the safer default for evaluation errors). |

Any combination of these three flags causes the CLI to set
`custom_policy.enabled = true` automatically — you do not need to pass a
separate `--secure` flag.

---

## `--opa-policy` vs `--opa-policy-module`

Use `--opa-policy` for the **single primary policy** that covers your main
enforcement concern (typically the command gate).  The file you pass here
populates `inline_rego`, which is evaluated as the root policy document.

Use `--opa-policy-module` when you need **multiple independent Rego packages**
active in the same sandbox — for example, one package for `declaw.platform.cmd`
and another for `declaw.platform.network`.  Each `--opa-policy-module` file
becomes one entry in `inline_modules`; all modules are loaded into the same OPA
bundle and evaluated together.

You can combine the two: `--opa-policy` sets the primary module and
`--opa-policy-module` adds supplementary modules.

---

## Rego Contract

All policy packages must follow this interface:

| Package | Input field | Type | Description |
|---|---|---|---|
| `declaw.platform.cmd` | `input.action.command` | `string` | Binary name / argv[0] |
| `declaw.platform.cmd` | `input.action.args` | `[string]` | argv[1..n] |
| `declaw.platform.network` | `input.action.destination` | `string` | Target host (or host:port) |
| `declaw.platform.content` | `input.attributes.scan_results.*` | `object` | Guardrails scan results |
| `declaw.platform.content` | `input.attributes.model` | `string` | Model identifier |

**Rule convention — deny-only sets:**

```rego
deny contains msg if {
    <condition>
    msg := "<human-readable reason>"
}
```

- `deny` is a **set** of reason strings.
- Any non-empty `deny` set causes the action to be blocked.
- Policies may only **tighten** the platform floor — they cannot relax a deny
  that the platform itself already enforces.
- There is no `allow` rule; the platform's built-in rules remain active
  regardless of what custom policies say.

---

## Command Gate vs Network Gate — Enforcement Difference

| Gate | Package | When deny fires | Observed effect |
|---|---|---|---|
| **Command** | `declaw.platform.cmd` | Before the process is spawned | API returns HTTP 403; `declaw sandbox exec` exits non-zero; no process runs |
| **Network** | `declaw.platform.network` | At the TCP connection layer (inside the network namespace) | The **command continues to run** but the egress connection is silently dropped (TCP reset / no route); the command may print a network error in its output |
| **Content** | `declaw.platform.content` | After guardrails scanning | The content is blocked; the command output or LLM response is not returned |

The key distinction: a **command deny is fatal to the process** (403 before
spawn), while a **network deny is silent at the connection level** (the process
lives but the socket goes nowhere).

---

## `--opa-fail-closed` and `default_deny`

`--opa-fail-closed` maps directly to `custom_policy.default_deny = true` in the
API request.

| Evaluator state | `default_deny = false` (fail-open, default) | `default_deny = true` (fail-closed) |
|---|---|---|
| Policy evaluates normally | Deny set respected | Deny set respected |
| OPA unreachable / error | Action **allowed** | Action **denied** |

Use fail-closed (`--opa-fail-closed`) in high-security environments where a
broken policy evaluator should never silently grant access.  Use fail-open
(the default, no flag) during development so a misconfigured policy does not
lock out the sandbox entirely.

---

## Sample Policy Files

| File | Package | What it does |
|---|---|---|
| `deny-dangerous-commands.rego` | `declaw.platform.cmd` | Blocks `rm`, `dd`, `mkfs`, `shred`, `wipefs`; blocks `curl`/`wget` to IMDS endpoints (169.254.169.254, metadata.google.internal, etc.) |
| `command-allowlist.rego` | `declaw.platform.cmd` | Denies any binary **not** in a curated allowlist — the inverse of a blocklist |
| `egress-allowlist.rego` | `declaw.platform.network` | Drops egress connections to destinations that do not match an approved-domain regex list (PyPI, npm, GitHub, OpenAI, Anthropic, etc.) |

---

## Exact Commands from `run.sh`

### Demo 1 — Single inline policy, fail-closed

```bash
# Create the sandbox
SANDBOX_ID=$(declaw sandbox create \
    --template node \
    --opa-policy deny-dangerous-commands.rego \
    --opa-fail-closed \
    --json | python3 -c "import sys,json; print(json.load(sys.stdin)['sandbox_id'])")

# Expect blocked (rm is in _blocked_commands) — non-zero exit
declaw sandbox exec "$SANDBOX_ID" -- rm -rf /tmp/x

# Expect blocked (curl with IMDS destination arg) — non-zero exit
declaw sandbox exec "$SANDBOX_ID" -- curl http://169.254.169.254/latest/meta-data/

# Expect allowed (ls not denied) — exit 0
declaw sandbox exec "$SANDBOX_ID" -- ls /tmp

# Cleanup
declaw sandbox kill "$SANDBOX_ID"
```

### Demo 2 — Multi-module: command allowlist + egress allowlist

```bash
# Create the sandbox with two independent Rego modules (different packages)
SANDBOX_ID=$(declaw sandbox create \
    --template node \
    --opa-policy-module command-allowlist.rego \
    --opa-policy-module egress-allowlist.rego \
    --json | python3 -c "import sys,json; print(json.load(sys.stdin)['sandbox_id'])")

# Expect allowed (ls is in _allowed_commands)
declaw sandbox exec "$SANDBOX_ID" -- ls /tmp

# Expect blocked (id is NOT in _allowed_commands) — non-zero exit
declaw sandbox exec "$SANDBOX_ID" -- id

# Cleanup
declaw sandbox kill "$SANDBOX_ID"
```

### `declaw sandbox exec` invocation form

```
declaw sandbox exec <sandbox-id> -- <argv0> [arg1 arg2 ...]
```

The `--` separator is required to prevent Cobra from interpreting the command
arguments as CLI flags.  `exec` takes the sandbox ID as the first positional
argument, then everything after `--` is treated as the command and its
arguments (joined into a shell string internally by the CLI).

---

## Running the Script

```bash
export DECLAW_API_KEY=sk-...
export DECLAW_DOMAIN=api.example.com   # no https:// prefix
bash run.sh
```

The script:
1. Validates `DECLAW_API_KEY`, `DECLAW_DOMAIN`, and that `declaw` is on `$PATH`.
2. Registers a `trap EXIT` cleanup that kills any sandboxes it created, even if
   the script exits early.
3. Creates two sandboxes (one per demo), runs assertions, then lets the trap
   clean up.

Do not run this against production without understanding the sandbox creation
costs and the API rate limits for your tier.
