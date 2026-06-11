# command-allowlist.rego
#
# Package: declaw.platform.cmd
#
# Purpose:
#   Implements a strict command allowlist: only the binaries listed in
#   `_allowed_commands` may be executed.  Every other binary is denied.
#
#   Use this as your primary policy (`--opa-policy`) when you need the
#   tightest possible command-execution surface.  Pair it with
#   egress-allowlist.rego via `--opa-policy-module` to also control network
#   egress in the same sandbox.
#
# Input contract (declaw.platform.cmd):
#   input.action.command  string   — the binary name / argv[0]
#   input.action.args     [string] — argv[1..n]
#
# Enforcement point: COMMAND gate
#   Rejected commands return HTTP 403 to the API caller and a non-zero exit
#   code when using `declaw sandbox exec`.
#
# Customisation:
#   Edit _allowed_commands to match the specific tooling your workload
#   requires.  Keep the set as small as possible.

package declaw.platform.cmd

import future.keywords.contains
import future.keywords.if

# ---------------------------------------------------------------------------
# Allowlist — only these binaries may be invoked.
# ---------------------------------------------------------------------------

_allowed_commands := {
    # POSIX file / text utilities
    "ls",
    "cat",
    "echo",
    "printf",
    "grep",
    "awk",
    "sed",
    "sort",
    "uniq",
    "wc",
    "head",
    "tail",
    "find",
    "stat",
    "file",
    "diff",
    "cp",
    "mv",
    "mkdir",
    "touch",
    "chmod",
    "chown",
    # Archive / compression
    "tar",
    "gzip",
    "gunzip",
    "zip",
    "unzip",
    # Runtimes / interpreters (adjust to your template)
    "python3",
    "python",
    "node",
    "npm",
    "npx",
    # Build / test tools (add only what your workload needs)
    "make",
    "go",
    "git",
    # Network diagnostics (read-only probes; NOT curl/wget — add below if needed)
    "ping",
    "dig",
    "nslookup",
    # Utilities
    "env",
    "printenv",
    "which",
    "date",
    "sleep",
    "true",
    "false",
    "test",
    "[",
    "bash",
    "sh",
}

deny contains msg if {
    not _allowed_commands[input.action.command]
    msg := sprintf(
        "command %q is not in the allowed-command list; update the policy allowlist to permit it",
        [input.action.command],
    )
}
