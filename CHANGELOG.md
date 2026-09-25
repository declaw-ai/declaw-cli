# Changelog

All notable changes to the Declaw CLI are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [v0.9.0] — 2026-09

_2026-09b train: template rebuild._

### Added

- `declaw template rebuild <template-id>`: retry a failed template build,
  streaming its output like `template build`; `--no-wait` returns once the
  build has started. Previously the only retry was delete-and-build. (#919)

## [v0.8.0] — 2026-09

_2026-09 train: working template builds._

### Changed

- `--run-cmd` takes its value verbatim, one command per flag. Commas no longer
  split a command, so `--run-cmd 'pip install "numpy>=1.24,<2"'` works.
- `declaw template build` prints the build ID first, then streams the build
  output to stderr while it waits, so stdout stays clean for `--json`. Ctrl-C
  stops waiting and exits 130 while the build keeps running; a failed build
  exits 1 after its output. Status checks that fail temporarily are retried
  for up to two minutes (from go-sdk v0.7.0).

### Fixed

- `declaw template build` always failed with `alias is required`. The new
  required `--alias` flag names the template; create sandboxes from it with
  `declaw sandbox create --template <alias>` or
  `declaw mcp --template <alias>`.
- Packages given with `--apt-package` are now installed; they were sent under
  a name the API ignores.

## [v0.7.0] — 2026-08

_2026-08 train: idempotent sandbox creation._

### Changed

- `declaw sandbox create` inherits idempotent creation from the Go SDK: an
  interrupted or retried create no longer risks leaving a second running,
  billable sandbox behind.

## [v0.6.0] — 2026-07

_2026-07 train: credential vault client + injection domain scoping._

### Added

- `declaw vault` — manage credential-vault secrets by name: `create`, `list`,
  `rotate`, `update-scopes`, `delete`, and `presets`. Secret values are
  write-only. Reference a secret from a sandbox so its value is injected at the
  egress proxy and never enters the sandbox (#386, #399, #408, #456).
- `--injection-domain` — opt-in scoping of injection scanning to specific
  destination hosts.

## [v0.5.0] — 2026-06

_2026-06 train: file-granular volumes, OPA governance._

### Added

- Mode-based volumes: write-back, file-granular backend, mount mode (#344).
- OPA custom-policy support: `content_gate`, `policy_ref`, out-of-box AI
  governance packs (#279, #345).

## [v0.4.0] — 2026-05

### Added

- `declaw mcp` — run an MCP server in a sandbox: `--env KEY` credential
  forwarding, `--file` flag, tuned default timeout (#328).

### Fixed

- Terminate sandbox on stdin close in `mcp`.
- `install.sh` + goreleaser bare-binary alignment.

## [v0.3.0] — 2026-05

### Added

- Native interactive stdio support (rides go-sdk v0.3.0, #320).

## [v0.2.0] — 2026-04

### Added

- Volume + template management improvements; signup URL in `auth login`.

## [v0.1.0] — 2026-04

### Added

- Initial release: `auth`, `sandbox` (create/list/info/kill/pause/resume/
  exec/connect/files), `template`, `volume`, `account`, `version`.
  Cross-compiled for 6 platforms.

[Unreleased]: https://github.com/declaw-ai/declaw-cli/compare/v0.4.0...HEAD
[v0.4.0]: https://github.com/declaw-ai/declaw-cli/compare/v0.3.0...v0.4.0
[v0.3.0]: https://github.com/declaw-ai/declaw-cli/compare/v0.2.0...v0.3.0
[v0.2.0]: https://github.com/declaw-ai/declaw-cli/compare/v0.1.0...v0.2.0
[v0.1.0]: https://github.com/declaw-ai/declaw-cli/releases/tag/v0.1.0
