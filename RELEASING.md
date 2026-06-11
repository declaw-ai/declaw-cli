# Releasing the CLI

Checklist for cutting a new release of `declaw-ai/declaw-cli`. The public
repo is a mirror of `declaw/cli/`, updated by **snapshot sync at release
time only** (don't sync between trains: the CLI may use unreleased go-sdk
APIs, and the mirror must always compile against the pinned declaw-go
release).

Publishing is automated: pushing a `vX.Y.Z` tag to the mirror triggers
`publish.yml` there, which runs the tests, cross-compiles all 6
platforms, and creates a GitHub Release with the binaries attached. The
tag is the release button.

## Prerequisites

- If the go-sdk also changed this train, **release it first** (the public
  CLI go.mod pins a published go-sdk tag).

## Steps

1. **Update changelog**

   Add the new version block at the top of `CHANGELOG.md`
   ([Keep a Changelog](https://keepachangelog.com/) format).

2. **Pin the new go-sdk** (skip if the go-sdk didn't change this train)

   In `declaw/cli/go.mod`, bump `github.com/declaw-ai/declaw-go` to the
   tag released in the go-sdk step, then refresh sums:

   ```bash
   cd declaw/cli
   GOWORK=off go mod tidy
   ```

3. **Verify** — both the workspace build (local dev) and the standalone
   configuration (what the public mirror gets):

   ```bash
   cd declaw/cli
   make test && make build && ./bin/declaw version
   GOWORK=off go build ./... && GOWORK=off go test ./...
   ```

4. **Commit + push**

   ```bash
   git commit -am "release(cli): vX.Y.Z"
   git push origin main
   ```

5. **Sync the public mirror** (gated by junk/secret scans; the mirror
   gets one commit with this message, never internal history)

   ```bash
   gh workflow run sync-mirror.yml -f component=cli \
     -f message="release: vX.Y.Z" && gh run watch
   ```

6. **Tag the mirror — this builds + releases**

   ```bash
   SHA=$(git ls-remote cli-public main | cut -f1)
   gh api repos/declaw-ai/declaw-cli/git/refs \
     -f ref=refs/tags/vX.Y.Z -f sha=$SHA
   ```

   Watch the run (`gh run list --repo declaw-ai/declaw-cli -w publish.yml`).
   It attaches: darwin-arm64, darwin-amd64, linux-amd64, linux-arm64,
   windows-amd64.exe, windows-arm64.exe (the README install table depends
   on these exact names). If it fails, nothing was released — fix and
   `gh run rerun` (the tag stays).

7. **Wrap up**: update the docs capability matrix (`client_docs`) if this
   is part of a release train.

## Manual fallback

If CI is unavailable:

```bash
cd declaw/cli && VERSION=vX.Y.Z make build-all
gh release create vX.Y.Z ./bin/* --repo declaw-ai/declaw-cli \
  --title "Declaw CLI vX.Y.Z" \
  --notes "<paste the CHANGELOG block for this version>"
```
