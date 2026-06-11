# egress-allowlist.rego
#
# Package: declaw.platform.network
#
# Purpose:
#   Implements a strict egress allowlist: outbound connections to destinations
#   not matching an approved domain pattern are silently dropped before the TCP
#   SYN leaves the sandbox network namespace.
#
#   Combine this with a command policy using --opa-policy-module so both gates
#   are active simultaneously in a single sandbox.
#
# Input contract (declaw.platform.network):
#   input.action.destination  string  — target host or IP:port resolved by
#                                       the orchestrator at connection time
#                                       (e.g. "api.openai.com", "1.2.3.4:443")
#
# Enforcement point: NETWORK gate
#   Unlike the command gate (which returns 403), network denials drop the
#   egress connection silently (TCP reset / no route).  The command itself
#   continues to run and may produce a network-error in its output, but it is
#   NOT killed.  This matches the platform's iptables-layer enforcement model.
#
# Design notes:
#   - Rules match on the *hostname* portion of the destination; port is
#     stripped before matching so a rule for "pypi.org" covers any port.
#   - Regex patterns use RE2 syntax (OPA's `regex.match`).
#   - Add a bare IP rule if your workload resolves DNS itself and sends to IPs
#     (prefer DNS-based allowlisting and let the sandbox resolver handle it).

package declaw.platform.network

import future.keywords.contains
import future.keywords.if

# ---------------------------------------------------------------------------
# Allowed destination patterns (RE2 regex, matched against host portion only)
# ---------------------------------------------------------------------------

_allowed_patterns := [
    # Package registries
    `^(.*\.)?pypi\.org$`,
    `^(.*\.)?pythonhosted\.org$`,
    `^(.*\.)?npmjs\.com$`,
    `^(.*\.)?registry\.npmjs\.org$`,
    `^(.*\.)?pkg\.go\.dev$`,
    `^(.*\.)?proxy\.golang\.org$`,
    `^(.*\.)?sum\.golang\.org$`,

    # Source control
    `^(.*\.)?github\.com$`,
    `^(.*\.)?githubusercontent\.com$`,
    `^(.*\.)?gitlab\.com$`,

    # AI / LLM APIs (add only what your workload calls)
    `^(.*\.)?api\.openai\.com$`,
    `^(.*\.)?api\.anthropic\.com$`,

    # DNS (plain-text DNS on 53 should not reach here, but belt-and-suspenders)
    `^(.*\.)?cloudflare-dns\.com$`,
    `^(.*\.)?dns\.google$`,
]

# ---------------------------------------------------------------------------
# Helper: strip port suffix and extract host
# ---------------------------------------------------------------------------

_host(destination) := host if {
    # destination may be "hostname:port" — take everything before the last ":"
    # For bare hostnames with no port this is a no-op.
    parts := split(destination, ":")
    count(parts) >= 2
    # Reassemble everything except the last segment (handles IPv6 brackets too)
    host := concat(":", array.slice(parts, 0, count(parts)-1))
} else := destination  # no colon → return as-is

_matches_any(host) if {
    pattern := _allowed_patterns[_]
    regex.match(pattern, host)
}

deny contains msg if {
    host := _host(input.action.destination)
    not _matches_any(host)
    msg := sprintf(
        "egress to %q is blocked: destination does not match any allowed pattern; add a pattern to egress-allowlist.rego to permit it",
        [input.action.destination],
    )
}
