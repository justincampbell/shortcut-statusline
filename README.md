# shortcut-statusline

A small Go CLI that prints info about the current Shortcut story for use in a
shell prompt or [Claude Code](https://claude.com/claude-code) statusline.

It reads the current git branch, extracts the story ID (e.g. `sc-12345`),
calls the Shortcut API for the story (and, if your format references them, the
epic and objective), and prints the rendered template. Results are cached
per-branch on disk so subsequent invocations are <20ms.

## Install

```sh
brew install justincampbell/tap/shortcut-statusline
```

or:

```sh
go install github.com/justincampbell/shortcut-statusline@latest
```

## Auth

`shortcut-statusline` reads the token in this order:

1. `SHORTCUT_API_TOKEN` env var
2. `~/.config/shortcut-cli/config.json` (the file the [`short`](https://github.com/useshortcut/shortcut-cli) CLI manages)

## Usage

```sh
shortcut-statusline -f '{story.name}'
shortcut-statusline -f '{story.name} • {epic.name} • {objective.name}'
shortcut-statusline -f '[{story.id}] {story.name}'
```

If the branch has no `sc-NNNNN` in it, or you're outside a git repo, the
command prints nothing and exits 0. Same on network errors: a stale cache is
used if available, otherwise nothing is printed — the prompt is never blocked.

### Tokens

| Token              | Value                          |
|--------------------|--------------------------------|
| `{story.name}`     | Story title                    |
| `{story.id}`       | Numeric ID                     |
| `{story.url}`      | App URL                        |
| `{story.state}`    | Workflow state name (e.g. "In Development") |
| `{epic.name}`      | Epic title                     |
| `{epic.id}`        | Epic ID                        |
| `{epic.url}`       | Epic app URL                   |
| `{epic.state}`     | Epic state name (e.g. "In Progress") |
| `{objective.name}` | Objective (milestone) title    |
| `{objective.id}`   | Objective ID                   |
| `{objective.url}`  | Objective app URL              |
| `{objective.state}` | Objective state (e.g. "in progress") |

Epic and objective are only fetched if your format references them. State
fields trigger a one-time workspace-wide workflow lookup, cached for 7 days
(override with `SHORTCUT_STATUSLINE_WORKFLOW_TTL`).

### Flags

- `-f, --format <string>` — default `{story.name}`
- `--no-cache` — bypass cache (for debugging)
- `--refresh` — clear cache for the current branch and refetch
- `-v, --version` — print version

### Cache

JSON files in `$XDG_CACHE_HOME/shortcut-statusline/` (or
`~/Library/Caches/shortcut-statusline/` on macOS), 1-hour TTL. Override the
TTL with `SHORTCUT_STATUSLINE_TTL` (Go duration, e.g. `5m`).

## Claude Code statusline

Add to `~/.claude/settings.json`:

```json
{
  "statusLine": {
    "type": "command",
    "command": "shortcut-statusline -f '{story.name} • {epic.name}'"
  }
}
```

## License

MIT
