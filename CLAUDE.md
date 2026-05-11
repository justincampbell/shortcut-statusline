# shortcut-statusline

Go CLI that prints Shortcut story info (story → epic → objective) for use in
a shell prompt or Claude Code statusline. Detects the story from the current
git branch (`sc-NNNNN`), fetches via the Shortcut API, caches per-branch.

## Architecture

- `main.go` — flag parsing, orchestration, quiet-failure error handling.
- `internal/branch` — finds `.git` (walking up; understands worktrees), reads HEAD, extracts `sc-NNNNN`.
- `internal/config` — token resolution: `SHORTCUT_API_TOKEN` → `~/.config/shortcut-cli/config.json` (XDG-style on all platforms; matches the `short` CLI).
- `internal/shortcut` — minimal HTTP client for Stories, Epics, Objectives. Injectable `BaseURL` for tests.
- `internal/cache` — per-branch JSON cache. Filename = SHA256(branch)[:16]. Atomic write via temp file + rename.
- `internal/format` — `{ns.field}` tokenizer + renderer. Pre-pass `Namespaces()` tells main.go which resources to fetch (lazy).

## Invariants

- **The statusline is hot.** Cache hits must stay <20ms. Network calls have a 3s timeout. Any error after-the-fact prefers stale cache > empty output > error. Never block a shell prompt.
- **Exit 0 on every recoverable failure.** Errors go to stderr; the prompt keeps moving.
- **Lazy fetch.** Only fetch the epic if the format references `epic.*` or `objective.*`. Only fetch the objective if the format references `objective.*`. Only fetch workflows if the format references `story.state` or `epic.state`. Workflow lookups are cached separately with a much longer TTL (`DefaultWorkflowTTL`, 7d) since they change rarely.

## Build / test / install

```
make test
make lint
make install   # → $GOBIN/shortcut-statusline (and a versioned copy)
```

Versioning is `git describe --tags --dirty --always`, embedded via
`-ldflags -X main.version=...`. Release via tag push (`v*`) triggers
goreleaser (see `.github/workflows/release.yml`).
