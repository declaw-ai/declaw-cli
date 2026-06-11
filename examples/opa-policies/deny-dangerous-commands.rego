# deny-dangerous-commands.rego
#
# Package: declaw.platform.cmd
#
# Purpose:
#   Blocks a narrow set of commands that can cause irreversible data loss or
#   expose cloud-instance metadata to the running process.  This policy
#   tightens the platform floor — it cannot relax any deny rule that the
#   platform itself already enforces.
#
# Input contract (declaw.platform.cmd):
#   input.action.command  string   — the binary name / argv[0]
#   input.action.args     [string] — argv[1..n]
#
# Rule convention:
#   deny is a *set* of reason strings.  Any non-empty deny set causes the
#   command to be rejected (HTTP 403 / non-zero exit from `declaw sandbox exec`).
#   A single policy may contribute multiple messages; all are returned in the
#   API error body so the caller knows exactly which rule fired.
#
# Enforcement point: COMMAND gate
#   The command is rejected before it reaches the sandbox.  The process is
#   never spawned.

package declaw.platform.cmd

import future.keywords.contains
import future.keywords.if

# ---------------------------------------------------------------------------
# Blocked binary names
# Add any binary whose unrestricted use poses an unacceptable risk.
# ---------------------------------------------------------------------------

_blocked_commands := {
    "rm",    # recursive / force removes; catastrophic in a shared-rootfs scenario
    "dd",    # raw-device writes — can wipe block devices
    "mkfs",  # filesystem formatting
    "shred", # secure-wipe — same risk as dd
    "wipefs",# partition-table erasure
}

deny contains msg if {
    _blocked_commands[input.action.command]
    msg := sprintf("command %q is blocked by policy deny-dangerous-commands", [input.action.command])
}

# ---------------------------------------------------------------------------
# IMDS / metadata endpoint access via curl / wget
#
# Blocks curl and wget calls whose argument list contains an address that
# looks like an AWS/GCP/Azure instance-metadata endpoint.
#
#   AWS  : 169.254.169.254
#   GCP  : metadata.google.internal  or  169.254.169.254
#   Azure: 169.254.169.254  or  169.254.169.253
#
# The check is intentionally broad: any argument token that *contains* the
# address string is matched, catching flags like --url=169.254.169.254/...
# ---------------------------------------------------------------------------

_metadata_fetchers := {"curl", "wget"}

_imds_patterns := {
    "169.254.169.254",
    "169.254.169.253",
    "metadata.google.internal",
    "metadata.goog",
}

_arg_contains_imds(arg) if {
    pattern := _imds_patterns[_]
    contains(arg, pattern)
}

deny contains msg if {
    _metadata_fetchers[input.action.command]
    arg := input.action.args[_]
    _arg_contains_imds(arg)
    msg := sprintf(
        "command %q with argument %q is blocked: IMDS endpoint access is not permitted",
        [input.action.command, arg],
    )
}
