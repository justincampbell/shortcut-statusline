# shortcut-statusline

A small Go CLI that prints info about the current Shortcut story for use in a
shell prompt or [Claude Code](https://claude.com/claude-code) statusline.

It reads the current git branch, resolves a Shortcut story ID from it
(first by parsing `sc-NNNNN` out of the branch, then — for branches like
`chore/sc-new-story/foo` — by asking Shortcut directly via
`/search/stories?query=branch:<name>`), calls the API for the story
(and, if your format references them, the epic and objective), and
prints the rendered template. Results are cached per-branch on disk so
subsequent invocations are <20ms.

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
shortcut-statusline                                  # default: {story.idName} ({epic.name})
shortcut-statusline -f '{story.name}'
shortcut-statusline -f '{story.name} • {epic.name} • {objective.name}'
shortcut-statusline -f '[{story.id}] {story.name}'
```

Empty enclosures (`()`, `[]`, `{}`) left by missing values are stripped from
the output, so the default format renders as `12345: Story Name` for a
story without an epic.

If the branch has no `sc-NNNNN` in it, or you're outside a git repo, the
command prints nothing and exits 0. Same on network errors: a stale cache is
used if available, otherwise nothing is printed — the prompt is never blocked.

### Tokens

| Token                | Value                          |
|----------------------|--------------------------------|
| `{story.name}`       | Story title                    |
| `{story.id}`         | Numeric ID                     |
| `{story.idName}`     | `id: name` (e.g. `12345: Build the thing`) |
| `{story.url}`        | App URL                        |
| `{story.state}`      | Workflow state name (e.g. "In Development") |
| `{story.type}`       | `Feature` / `Bug` / `Chore`    |
| `{story.typeChar}`   | One-letter form: `F` / `B` / `C` |
| `{story.owner}`      | First owner's mention handle (e.g. `justincampbell`) |
| `{story.ownerMention}` | `@`-prefixed handle, Shortcut notification form (e.g. `@justincampbell`) |
| `{story.ownerName}`  | Full display name (e.g. `Justin Campbell`) |
| `{story.requestor}`  | Requestor's mention handle     |
| `{story.requestorMention}` | `@`-prefixed handle          |
| `{story.requestorName}` | Full display name           |
| `{story.team}`       | Team handle (e.g. `platform`)  |
| `{story.teamMention}`| `@`-prefixed team handle       |
| `{story.teamName}`   | Team display name (e.g. `Platform`) |
| `{epic.name}`        | Epic title                     |
| `{epic.id}`          | Epic ID                        |
| `{epic.idName}`      | `id: name`                     |
| `{epic.url}`         | Epic app URL                   |
| `{epic.state}`       | Epic state name (e.g. "In Progress") |
| `{epic.owner}` / `{epic.ownerMention}` / `{epic.ownerName}` | Same three variants as story |
| `{epic.team}` / `{epic.teamMention}` / `{epic.teamName}` | Same three variants as story |
| `{objective.name}`   | Objective (milestone) title    |
| `{objective.id}`     | Objective ID                   |
| `{objective.idName}` | `id: name`                     |
| `{objective.url}`    | Objective app URL              |
| `{objective.state}`  | Objective state (e.g. "in progress") |

Epic and objective are only fetched if your format references them. State,
owner/requestor, and team tokens trigger one-time workspace-wide lookups
(`/workflows`, `/members`, `/groups`), each cached for 7 days (override the
shared TTL with `SHORTCUT_STATUSLINE_WORKFLOW_TTL`).

### Flags

- `-f, --format <string>` — default `{story.idName} ({epic.name})`
- `--no-cache` — bypass cache (for debugging)
- `--refresh` — clear cache for the current branch and refetch
- `--no-links` — disable OSC8 hyperlinks
- `--no-color` — disable ANSI color
- `-v, --version` — print version

### Color

Name, id, idName, and state tokens are wrapped in an ANSI color reflecting
the resource's workflow state, roughly matching Shortcut's web UI:

| State type                 | Color   |
|----------------------------|---------|
| `backlog` / `unstarted` / `to do` | gray    |
| `started` / `in progress` | magenta |
| `done`                     | green   |

`{story.type}` has its own palette:

| Story type | Color  |
|------------|--------|
| `bug`      | red    |
| `chore`    | yellow |
| `feature`  | cyan   |

On by default. Disable with `--no-color`, `SHORTCUT_STATUSLINE_NO_COLOR=1`,
or `NO_COLOR` (any non-empty value, per [no-color.org](https://no-color.org/)).

### Hyperlinks (OSC8)

`name` and `id` tokens (e.g. `{story.name}`, `{epic.id}`) are wrapped in
[OSC8 hyperlink escape sequences](https://gist.github.com/egmontkob/eb114294efbcd5adb1944c9f3cb5feda)
pointing at the object's app URL. Supporting terminals (iTerm2, WezTerm,
Kitty, Ghostty, Alacritty, modern Apple Terminal, VS Code, Windows Terminal,
…) render them as clickable links; others ignore the framing and show the
text. On by default.

Disable with `--no-links`, `SHORTCUT_STATUSLINE_NO_LINKS=1`, or the
cross-tool [`NO_COLOR`](https://no-color.org/) (set to any non-empty value).

If you're inside tmux and links don't render, tmux is likely stripping the
escapes. Add this to `~/.tmux.conf`:

```tmux
set -as terminal-features ',*:hyperlinks'
```

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
