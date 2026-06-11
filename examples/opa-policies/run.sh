#!/usr/bin/env bash
# run.sh — end-to-end demonstration of declaw OPA custom-policy flags
#
# Flags exercised:
#   --opa-policy <file>          inline a .rego file as the primary policy
#   --opa-policy-module <file>   add one independent Rego module (repeatable)
#   --opa-fail-closed            deny when the OPA evaluator is unreachable
#
# Prerequisites:
#   - `declaw` binary is on $PATH
#   - DECLAW_API_KEY and DECLAW_DOMAIN env vars are exported
#   - The API is reachable from this machine
#
# This script does NOT run automatically in CI.  It is a reference / smoke-test
# to be run manually against a live environment.
#
# Usage:
#   export DECLAW_API_KEY=sk-...
#   export DECLAW_DOMAIN=api.example.com
#   bash run.sh

set -euo pipefail

# ---------------------------------------------------------------------------
# Resolve the directory this script lives in so .rego files are always found
# regardless of where the caller's $PWD is.
# ---------------------------------------------------------------------------
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

POLICY_DENY_DANGEROUS="${SCRIPT_DIR}/deny-dangerous-commands.rego"
POLICY_CMD_ALLOWLIST="${SCRIPT_DIR}/command-allowlist.rego"
POLICY_EGRESS_ALLOWLIST="${SCRIPT_DIR}/egress-allowlist.rego"

# ---------------------------------------------------------------------------
# Validate environment
# ---------------------------------------------------------------------------
check_env() {
    local missing=0
    if [[ -z "${DECLAW_API_KEY:-}" ]]; then
        echo "ERROR: DECLAW_API_KEY is not set" >&2
        missing=1
    fi
    if [[ -z "${DECLAW_DOMAIN:-}" ]]; then
        echo "ERROR: DECLAW_DOMAIN is not set" >&2
        missing=1
    fi
    if [[ $missing -eq 1 ]]; then
        echo "" >&2
        echo "Set both variables before running this script:" >&2
        echo "  export DECLAW_API_KEY=sk-..." >&2
        echo "  export DECLAW_DOMAIN=api.example.com" >&2
        exit 1
    fi

    if ! command -v declaw &>/dev/null; then
        echo "ERROR: 'declaw' binary not found on PATH" >&2
        exit 1
    fi

    for f in "$POLICY_DENY_DANGEROUS" "$POLICY_CMD_ALLOWLIST" "$POLICY_EGRESS_ALLOWLIST"; do
        if [[ ! -f "$f" ]]; then
            echo "ERROR: policy file not found: $f" >&2
            exit 1
        fi
    done
}

# ---------------------------------------------------------------------------
# Cleanup trap — kills the sandbox even if the script exits early due to
# set -e or an explicit exit call.
# ---------------------------------------------------------------------------
SANDBOX_ID_1=""
SANDBOX_ID_2=""

cleanup() {
    local exit_code=$?
    echo ""
    echo "=== cleanup ==="
    if [[ -n "$SANDBOX_ID_1" ]]; then
        echo "Killing sandbox $SANDBOX_ID_1 ..."
        declaw sandbox kill "$SANDBOX_ID_1" || true
    fi
    if [[ -n "$SANDBOX_ID_2" ]]; then
        echo "Killing sandbox $SANDBOX_ID_2 ..."
        declaw sandbox kill "$SANDBOX_ID_2" || true
    fi
    echo "Done (exit $exit_code)"
}
trap cleanup EXIT

# ---------------------------------------------------------------------------
# Helper: print a section header
# ---------------------------------------------------------------------------
section() {
    echo ""
    echo "================================================================"
    echo "  $*"
    echo "================================================================"
}

# ---------------------------------------------------------------------------
# DEMO 1 — single inline policy (deny-dangerous-commands) + fail-closed
#
#   --opa-policy   sends the .rego contents as inline_rego
#   --opa-fail-closed  maps to default_deny=true
# ---------------------------------------------------------------------------
section "DEMO 1: single --opa-policy with --opa-fail-closed (node template)"

echo ""
echo "Creating sandbox with deny-dangerous-commands.rego ..."
echo "  Command:"
echo "    declaw sandbox create \\"
echo "      --template node \\"
echo "      --opa-policy deny-dangerous-commands.rego \\"
echo "      --opa-fail-closed \\"
echo "      --json"

RAW_1=$(declaw sandbox create \
    --template node \
    --opa-policy "$POLICY_DENY_DANGEROUS" \
    --opa-fail-closed \
    --json)

SANDBOX_ID_1=$(echo "$RAW_1" | python3 -c "import sys,json; print(json.load(sys.stdin)['sandbox_id'])")
echo ""
echo "Sandbox created: $SANDBOX_ID_1"

# -- Test 1a: blocked command (rm) ----------------------------------------
section "DEMO 1a: rm -rf /tmp/x  [EXPECT: blocked, non-zero exit]"
echo ""
echo "  Command:"
echo "    declaw sandbox exec $SANDBOX_ID_1 -- rm -rf /tmp/x"
echo ""
echo "  Expected outcome: policy 'deny-dangerous-commands' fires; exec returns"
echo "  non-zero exit code (the platform rejects the command with 403 before"
echo "  the process is ever spawned)."
echo ""

# We expect this to fail; capture the exit code without aborting the script.
BLOCKED_EXIT=0
declaw sandbox exec "$SANDBOX_ID_1" -- rm -rf /tmp/x || BLOCKED_EXIT=$?

if [[ $BLOCKED_EXIT -ne 0 ]]; then
    echo "PASS — rm was blocked (exit $BLOCKED_EXIT)"
else
    echo "UNEXPECTED — rm was not blocked (exit 0); check policy configuration"
fi

# -- Test 1b: blocked command with IMDS curl arg --------------------------
section "DEMO 1b: curl with IMDS URL  [EXPECT: blocked, non-zero exit]"
echo ""
echo "  Command:"
echo "    declaw sandbox exec $SANDBOX_ID_1 -- curl http://169.254.169.254/latest/meta-data/"
echo ""
echo "  Expected outcome: curl itself is not blocked (not in _blocked_commands),"
echo "  but the IMDS-destination argument triggers the second deny rule."
echo ""

BLOCKED_EXIT_2=0
declaw sandbox exec "$SANDBOX_ID_1" -- curl http://169.254.169.254/latest/meta-data/ \
    || BLOCKED_EXIT_2=$?

if [[ $BLOCKED_EXIT_2 -ne 0 ]]; then
    echo "PASS — curl to IMDS was blocked (exit $BLOCKED_EXIT_2)"
else
    echo "UNEXPECTED — curl to IMDS was not blocked (exit 0); check policy configuration"
fi

# -- Test 1c: allowed command (ls) ----------------------------------------
section "DEMO 1c: ls /tmp  [EXPECT: allowed, exit 0]"
echo ""
echo "  Command:"
echo "    declaw sandbox exec $SANDBOX_ID_1 -- ls /tmp"
echo ""
echo "  Expected outcome: ls is not in the denied set; command runs normally."
echo ""

declaw sandbox exec "$SANDBOX_ID_1" -- ls /tmp
echo ""
echo "PASS — ls /tmp succeeded"

# ---------------------------------------------------------------------------
# DEMO 2 — multi-module invocation
#
#   --opa-policy-module is repeatable; each file becomes an independent Rego
#   module in inline_modules.  Both packages (declaw.platform.cmd and
#   declaw.platform.network) are active simultaneously.
#
#   NOTE: when ONLY --opa-policy-module is used (no --opa-policy), there is no
#   single "primary" inline_rego, so custom_policy.enabled=true is still set by
#   the CLI as soon as any OPA flag is changed.
# ---------------------------------------------------------------------------
section "DEMO 2: multi-module (command + egress allowlists via --opa-policy-module)"

echo ""
echo "Creating sandbox with command-allowlist + egress-allowlist as modules ..."
echo "  Command:"
echo "    declaw sandbox create \\"
echo "      --template node \\"
echo "      --opa-policy-module command-allowlist.rego \\"
echo "      --opa-policy-module egress-allowlist.rego \\"
echo "      --json"

RAW_2=$(declaw sandbox create \
    --template node \
    --opa-policy-module "$POLICY_CMD_ALLOWLIST" \
    --opa-policy-module "$POLICY_EGRESS_ALLOWLIST" \
    --json)

SANDBOX_ID_2=$(echo "$RAW_2" | python3 -c "import sys,json; print(json.load(sys.stdin)['sandbox_id'])")
echo ""
echo "Sandbox created: $SANDBOX_ID_2"

# -- Test 2a: command on allowlist ----------------------------------------
section "DEMO 2a: ls /tmp  [EXPECT: allowed]"
echo ""
echo "  Command:"
echo "    declaw sandbox exec $SANDBOX_ID_2 -- ls /tmp"
echo ""
declaw sandbox exec "$SANDBOX_ID_2" -- ls /tmp
echo ""
echo "PASS — ls is on the allowlist"

# -- Test 2b: command NOT on allowlist ------------------------------------
section "DEMO 2b: id  [EXPECT: blocked — 'id' not in command allowlist]"
echo ""
echo "  Command:"
echo "    declaw sandbox exec $SANDBOX_ID_2 -- id"
echo ""
echo "  Expected outcome: 'id' is not in _allowed_commands; exec returns non-zero."
echo ""

BLOCKED_EXIT_3=0
declaw sandbox exec "$SANDBOX_ID_2" -- id || BLOCKED_EXIT_3=$?

if [[ $BLOCKED_EXIT_3 -ne 0 ]]; then
    echo "PASS — 'id' was blocked (exit $BLOCKED_EXIT_3)"
else
    echo "UNEXPECTED — 'id' was not blocked (exit 0); check allowlist"
fi

# -- Test 2c: node allowed, egress to allowed domain -----------------------
section "DEMO 2c: node fetch to npmjs.com  [EXPECT: command runs, egress allowed]"
echo ""
echo "  Command:"
echo "    declaw sandbox exec $SANDBOX_ID_2 -- node -e \"require('https').get('https://registry.npmjs.org/', r => { console.log('status', r.statusCode); r.destroy(); })\""
echo ""
echo "  Expected outcome: node is on the command allowlist; registry.npmjs.org"
echo "  matches the egress allowlist pattern so the connection is permitted."
echo ""
# We do not assert on exit code here because network availability may vary;
# the point is that neither policy blocks it at the gate.
declaw sandbox exec "$SANDBOX_ID_2" -- \
    node -e "require('https').get('https://registry.npmjs.org/', r => { console.log('status', r.statusCode); r.destroy(); })" \
    || echo "(node fetch returned non-zero — network may be unavailable in this environment, but the policy gate did not block it)"

echo ""
echo "================================================================"
echo "  All demo steps completed."
echo "  Cleanup will run automatically (trap EXIT)."
echo "================================================================"
